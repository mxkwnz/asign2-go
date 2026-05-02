package consumer

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
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
	conn         *amqp.Connection
	channel      *amqp.Channel
	queue        string
	processedIDs map[string]struct{}
	mu           sync.Mutex
}

func NewNotificationConsumer(amqpURL, queueName string) (*NotificationConsumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	ch.Qos(1, 0, false)

	log.Printf("[Consumer] Connected to RabbitMQ, queue=%s", queueName)
	return &NotificationConsumer{
		conn:         conn,
		channel:      ch,
		queue:        queueName,
		processedIDs: make(map[string]struct{}),
	}, nil
}

func (c *NotificationConsumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("[Consumer] Waiting for messages...")
	for {
		select {
		case <-ctx.Done():
			log.Println("[Consumer] Context cancelled, stopping consumer.")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Println("[Consumer] Channel closed.")
				return nil
			}
			c.handleMessage(msg)
		}
	}
}

func (c *NotificationConsumer) handleMessage(msg amqp.Delivery) {
	var event PaymentEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("[Consumer] Failed to parse message: %v — NACKing (no requeue)", err)
		msg.Nack(false, false)
		return
	}

	c.mu.Lock()
	_, alreadyProcessed := c.processedIDs[event.EventID]
	if !alreadyProcessed {
		c.processedIDs[event.EventID] = struct{}{}
	}
	c.mu.Unlock()

	if alreadyProcessed {
		log.Printf("[Consumer] Duplicate event %s — skipping, ACKing", event.EventID)
		msg.Ack(false)
		return
	}

	log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f",
		event.CustomerEmail,
		event.OrderID,
		float64(event.Amount)/100.0,
	)

	msg.Ack(false)
}

func (c *NotificationConsumer) Close() {
	c.channel.Close()
	c.conn.Close()
}
