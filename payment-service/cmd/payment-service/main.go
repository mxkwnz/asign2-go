package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"payment-service/internal/messaging"
	"payment-service/internal/repository"
	grpcHandler "payment-service/internal/transport/grpc"
	httpHandler "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	paymentv1 "github.com/mxkwnz/ap2-generated/payment/v1"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:Taktalifum123@localhost:5432/payment_db?sslmode=disable"
	}
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var publisher messaging.Publisher
	for i := 0; i < 10; i++ {
		publisher, err = messaging.NewRabbitPublisher(amqpURL, "payment.completed")
		if err == nil {
			break
		}
		log.Printf("Waiting for RabbitMQ... attempt %d: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer publisher.Close()

	repo := repository.NewPaymentRepo(db)
	uc := usecase.NewPaymentUseCase(repo, publisher)

	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":9091"
	}
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(grpcServer, grpcHandler.NewPaymentGRPCServer(uc))
	go func() {
		log.Printf("Payment gRPC server on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	handler := httpHandler.NewHandler(uc)
	r := gin.Default()
	r.GET("/payments/:order_id", handler.GetPayment)
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8081"
	}
	srv := &http.Server{Addr: httpAddr, Handler: r}
	go func() {
		log.Printf("Payment HTTP server on %s", httpAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Payment Service...")

	grpcServer.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Payment Service stopped.")
}
