package worker

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"notification-service/internal/provider"

	"github.com/redis/go-redis/v9"
)

type NotificationWorker struct {
	emailSender provider.EmailSender
	redis       *redis.Client
	maxRetries  int
}

func NewNotificationWorker(sender provider.EmailSender, redisClient *redis.Client) *NotificationWorker {
	maxRetries := 3
	if v := os.Getenv("MAX_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxRetries = n
		}
	}
	return &NotificationWorker{
		emailSender: sender,
		redis:       redisClient,
		maxRetries:  maxRetries,
	}
}

func (w *NotificationWorker) idempotencyKey(paymentID string) string {
	return fmt.Sprintf("notification:sent:%s", paymentID)
}

func (w *NotificationWorker) ProcessNotification(ctx context.Context, eventID, orderID, email string, amount int64) error {
	key := w.idempotencyKey(eventID)

	val, err := w.redis.Get(ctx, key).Result()
	if err == nil && val == "sent" {
		log.Printf("[Worker] Event %s already processed — skipping (idempotent)", eventID)
		return nil
	}

	subject := fmt.Sprintf("Payment Confirmed for Order #%s", orderID)
	body := fmt.Sprintf(
		"Hello,\n\nYour payment of $%.2f for order #%s has been confirmed.\n\nThank you!",
		float64(amount)/100.0, orderID,
	)

	var lastErr error
	for attempt := 1; attempt <= w.maxRetries; attempt++ {
		log.Printf("[Worker] Attempt %d/%d sending notification for event %s to %s",
			attempt, w.maxRetries, eventID, email)

		lastErr = w.emailSender.Send(ctx, email, subject, body)
		if lastErr == nil {
			w.redis.Set(ctx, key, "sent", 24*time.Hour)
			log.Printf("[Worker] Notification sent successfully for event %s (attempt %d)", eventID, attempt)
			return nil
		}

		log.Printf("[Worker] Attempt %d failed: %v", attempt, lastErr)

		if attempt < w.maxRetries {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			log.Printf("[Worker] Retrying in %s...", backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}

	log.Printf("[Worker] All %d attempts failed for event %s: %v", w.maxRetries, eventID, lastErr)
	return fmt.Errorf("notification failed after %d attempts: %w", w.maxRetries, lastErr)
}
