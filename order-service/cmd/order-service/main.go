package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"order-service/internal/ports"
	"order-service/internal/repository"
	httpHandler "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:Taktalifum123@localhost:5432/order_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewOrderRepo(db)

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	paymentClient := ports.NewPaymentClient("http://localhost:8081", httpClient)

	uc := usecase.NewOrderUseCase(repo, paymentClient)
	handler := httpHandler.NewHandler(uc)

	r := gin.Default()
	r.POST("/orders", handler.CreateOrder)
	r.GET("/orders/:id", handler.GetOrder)
	r.PATCH("/orders/:id/cancel", handler.CancelOrder)
	r.GET("orders/revenue", handler.GetRevenue)

	r.Run(":8080")
}
