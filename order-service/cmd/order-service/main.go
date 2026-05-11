package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"order-service/internal/cache"
	"order-service/internal/middleware"
	"order-service/internal/ports"
	"order-service/internal/repository"
	grpcHandler "order-service/internal/transport/grpc"
	httpHandler "order-service/internal/transport/http"
	"order-service/internal/usecase"

	orderv1 "github.com/mxkwnz/ap2-generated/order/v1"
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

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("Connected to Redis at %s", redisAddr)

	orderCache := cache.NewRedisOrderCache()
	uc := usecase.NewOrderUseCase(repo, paymentClient, orderCache)

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
		log.Printf("Order gRPC server on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	rateLimiter := middleware.NewRateLimiter(redisClient)
	handler := httpHandler.NewHandler(uc)

	r := gin.Default()
	r.Use(rateLimiter.Middleware())

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
