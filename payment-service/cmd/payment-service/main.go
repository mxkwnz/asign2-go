package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	paymentv1 "github.com/mxkwnz/ap2-generated/payment/v1"
	"payment-service/internal/repository"
	grpcHandler "payment-service/internal/transport/grpc"
	httpHandler "payment-service/internal/transport/http"
	"payment-service/internal/usecase"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:Taktalifum123@localhost:5432/payment_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPaymentRepo(db)
	uc := usecase.NewPaymentUseCase(repo)

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
		log.Printf("Payment gRPC server listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	handler := httpHandler.NewHandler(uc)
	r := gin.Default()
	r.GET("/payments/:order_id", handler.GetPayment)

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8081"
	}
	r.Run(httpAddr)
}
