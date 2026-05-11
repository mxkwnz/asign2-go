package provider

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type MockEmailSender struct {
	FailureRate float64
}

func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{FailureRate: 0.3}
}

func (m *MockEmailSender) Send(ctx context.Context, to, subject, body string) error {
	latency := time.Duration(100+rand.Intn(400)) * time.Millisecond
	log.Printf("[MockProvider] Simulating %s network latency...", latency)
	time.Sleep(latency)

	if rand.Float64() < m.FailureRate {
		return fmt.Errorf("mock provider: simulated transient network error")
	}

	log.Printf("[MockProvider] Email sent to=%s subject=%q", to, subject)
	return nil
}
