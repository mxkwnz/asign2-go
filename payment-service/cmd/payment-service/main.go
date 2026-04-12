package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"payment-service/internal/repository"
	httpHandler "payment-service/internal/transport/http"
	"payment-service/internal/usecase"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:Taktalifum123@localhost:5432/payment_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewPaymentRepo(db)
	uc := usecase.NewPaymentUseCase(repo)
	handler := httpHandler.NewHandler(uc)

	r := gin.Default()
	r.POST("/payments", handler.CreatePayment)
	r.GET("/payments/:order_id", handler.GetPayment)

	r.Run(":8081")
}
