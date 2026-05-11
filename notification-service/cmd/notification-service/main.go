package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/consumer"
	"notification-service/internal/provider"
	"notification-service/internal/worker"

	"github.com/redis/go-redis/v9"
)

func main() {
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		if err := redisClient.Ping(ctx).Err(); err == nil {
			log.Printf("Connected to Redis at %s", redisAddr)
			break
		} else {
			log.Printf("Waiting for Redis... attempt %d: %v", i+1, err)
			time.Sleep(3 * time.Second)
		}
	}

	emailSender := provider.NewEmailSender()

	notifWorker := worker.NewNotificationWorker(emailSender, redisClient)

	var c *consumer.NotificationConsumer
	var err error
	for i := 0; i < 10; i++ {
		c, err = consumer.NewNotificationConsumer(amqpURL, "payment.completed", redisClient, notifWorker)
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

	runCtx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down Notification Service...")
		cancel()
	}()

	if err := c.Start(runCtx); err != nil {
		log.Fatal(err)
	}
	log.Println("Notification Service stopped.")
}
