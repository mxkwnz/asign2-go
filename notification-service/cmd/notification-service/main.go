package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/consumer"
)

func main() {
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	var c *consumer.NotificationConsumer
	var err error
	for i := 0; i < 10; i++ {
		c, err = consumer.NewNotificationConsumer(amqpURL, "payment.completed")
		if err == nil {
			break
		}
		log.Printf("Waiting for RabbitMQ... attempt %d: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down Notification Service...")
		cancel()
	}()

	if err := c.Start(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("Notification Service stopped.")
}
