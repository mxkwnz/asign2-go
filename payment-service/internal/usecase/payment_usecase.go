package usecase

import (
	"context"
	"payment-service/internal/domain"
	"payment-service/internal/repository"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo repository.PaymentRepository
}

func NewPaymentUseCase(r repository.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: r}
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

	err := uc.repo.Save(ctx, payment)
	return payment, err
}

func (uc *PaymentUseCase) GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error) {
	return uc.repo.GetByOrderID(ctx, orderID)
}
