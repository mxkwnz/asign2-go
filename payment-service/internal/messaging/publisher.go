package messaging

import (
	"context"
	"encoding/json"
	"log"
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

type Publisher interface {
	Publish(ctx context.Context, event PaymentEvent) error
	Close()
}

type rabbitPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func NewRabbitPublisher(amqpURL, queueName string) (Publisher, error) {
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

	log.Printf("[Publisher] Connected to RabbitMQ, queue=%s", queueName)
	return &rabbitPublisher{conn: conn, channel: ch, queue: queueName}, nil
}

func (p *rabbitPublisher) Publish(ctx context.Context, event PaymentEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.channel.PublishWithContext(ctx,
		"",
		p.queue,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return err
	}

	log.Printf("[Publisher] Published event for order %s", event.OrderID)
	return nil
}

func (p *rabbitPublisher) Close() {
	p.channel.Close()
	p.conn.Close()
}
