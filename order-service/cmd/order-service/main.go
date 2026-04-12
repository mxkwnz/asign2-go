package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	orderv1 "github.com/mxkwnz/ap2-generated/order/v1"
	"order-service/internal/ports"
	"order-service/internal/repository"
	grpcHandler "order-service/internal/transport/grpc"
	httpHandler "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:Taktalifum123@localhost:5432/order_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewOrderRepo(db)

	paymentClient, err := ports.NewGRPCPaymentClient()
	if err != nil {
		log.Fatal("failed to connect to payment gRPC:", err)
	}

	uc := usecase.NewOrderUseCase(repo, paymentClient)

	grpcAddr := os.Getenv("ORDER_GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":9090"
	}
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(grpcServer, grpcHandler.NewOrderGRPCServer(db))

	go func() {
		log.Printf("Order gRPC streaming server on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	handler := httpHandler.NewHandler(uc)
	r := gin.Default()
	r.POST("/orders", handler.CreateOrder)
	r.GET("/orders/:id", handler.GetOrder)
	r.PATCH("/orders/:id/cancel", handler.CancelOrder)
	r.GET("/orders/revenue", handler.GetRevenue)

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}
	r.Run(httpAddr)
}
