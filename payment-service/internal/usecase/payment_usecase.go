package usecase

import (
	"context"
	"fmt"
	"payment-service/internal/domain"
	"payment-service/internal/messaging"
	"payment-service/internal/repository"
	"time"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo      repository.PaymentRepository
	publisher messaging.Publisher
}

func NewPaymentUseCase(r repository.PaymentRepository, pub messaging.Publisher) *PaymentUseCase {
	return &PaymentUseCase{repo: r, publisher: pub}
}

func (uc *PaymentUseCase) CreatePayment(ctx context.Context, orderID string, amount int64) (domain.Payment, error) {
	payment := domain.Payment{
		ID:      uuid.New().String(),
		OrderID: orderID,
		Amount:  amount,
	}

	if amount > 100000 {
		payment.Status = "Declined"
	} else {
		payment.Status = "Authorized"
		payment.TransactionID = uuid.New().String()
	}

	if err := uc.repo.Save(ctx, payment); err != nil {
		return payment, err
	}

	if payment.Status == "Authorized" {
		event := messaging.PaymentEvent{
			EventID:       uuid.New().String(),
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			CustomerEmail: fmt.Sprintf("customer-%s@example.com", orderID),
			Status:        payment.Status,
			OccurredAt:    time.Now(),
		}
		if err := uc.publisher.Publish(ctx, event); err != nil {
			_ = err
		}
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error) {
	return uc.repo.GetByOrderID(ctx, orderID)
}

func (uc *PaymentUseCase) ListPayments(ctx context.Context, minAmount, maxAmount int64) ([]domain.Payment, error) {
	return uc.repo.ListByAmountRange(ctx, minAmount, maxAmount)
}
