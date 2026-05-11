package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"notification-service/internal/worker"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type PaymentEvent struct {
	EventID       string    `json:"event_id"`
	OrderID       string    `json:"order_id"`
	Amount        int64     `json:"amount"`
	CustomerEmail string    `json:"customer_email"`
	Status        string    `json:"status"`
	OccurredAt    time.Time `json:"occurred_at"`
}

type NotificationConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
	worker  *worker.NotificationWorker
}

func NewNotificationConsumer(amqpURL, queueName string, redisClient *redis.Client, w *worker.NotificationWorker) (*NotificationConsumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	ch.Qos(1, 0, false)

	log.Printf("[Consumer] Connected to RabbitMQ, queue=%s", queueName)
	return &NotificationConsumer{
		conn:    conn,
		channel: ch,
		queue:   queueName,
		worker:  w,
	}, nil
}

func (c *NotificationConsumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Println("[Consumer] Waiting for messages...")
	for {
		select {
		case <-ctx.Done():
			log.Println("[Consumer] Context cancelled, stopping.")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Println("[Consumer] Channel closed.")
				return nil
			}
			c.handleMessage(ctx, msg)
		}
	}
}

func (c *NotificationConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	var event PaymentEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("[Consumer] Bad message: %v — discarding", err)
		msg.Nack(false, false) // discard poison message
		return
	}

	log.Printf("[Consumer] Processing event %s for order %s", event.EventID, event.OrderID)

	err := c.worker.ProcessNotification(ctx, event.EventID, event.OrderID, event.CustomerEmail, event.Amount)
	if err != nil {
		log.Printf("[Consumer] Failed to process event %s: %v — NACKing for requeue", event.EventID, err)
		msg.Nack(false, true) // requeue for later
		return
	}

	msg.Ack(false)
}

func (c *NotificationConsumer) Close() {
	c.channel.Close()
	c.conn.Close()
}
