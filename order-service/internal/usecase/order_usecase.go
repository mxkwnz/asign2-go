package usecase

import (
	"context"
	"errors"
	"log"
	"order-service/internal/cache"
	"order-service/internal/domain"
	"order-service/internal/ports"
	"order-service/internal/repository"
)

type OrderUseCase struct {
	repo    repository.OrderRepository
	payment ports.PaymentPort
	cache   cache.OrderCache
}

func NewOrderUseCase(r repository.OrderRepository, p ports.PaymentPort, c cache.OrderCache) *OrderUseCase {
	return &OrderUseCase{repo: r, payment: p, cache: c}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, o domain.Order) (domain.Order, error) {
	if o.IdempotencyKey != "" {
		existing, err := uc.repo.GetByIdempotencyKey(ctx, o.IdempotencyKey)
		if err == nil && existing.ID != "" {
			return existing, nil
		}
	}

	if o.Amount <= 0 {
		return domain.Order{}, errors.New("invalid amount")
	}

	o.Status = "Pending"
	if o.IdempotencyKey == "" {
		o.IdempotencyKey = o.ID
	}
	if err := uc.repo.Create(ctx, o); err != nil {
		log.Printf("Failed to create order: %v", err)
		return domain.Order{}, err
	}

	log.Printf("Calling payment for order %s amount %d", o.ID, o.Amount)
	status, err := uc.payment.Pay(ctx, o.ID, o.Amount)
	if err != nil {
		log.Printf("Payment failed with error: %v", err)
		uc.repo.UpdateStatus(ctx, o.ID, "Failed")
		o.Status = "Failed"
		uc.cache.Delete(ctx, o.ID)
		return o, nil
	}

	log.Printf("Payment status: %s", status)
	if status == "Authorized" {
		uc.repo.UpdateStatus(ctx, o.ID, "Paid")
		o.Status = "Paid"
	} else {
		uc.repo.UpdateStatus(ctx, o.ID, "Failed")
		o.Status = "Failed"
	}

	uc.cache.Delete(ctx, o.ID)

	return o, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (domain.Order, error) {
	if cached, ok := uc.cache.Get(ctx, id); ok {
		return cached, nil
	}

	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}

	uc.cache.Set(ctx, order)

	return order, nil
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) error {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if order.Status == "Paid" {
		return errors.New("cannot cancel paid order")
	}

	err = uc.repo.UpdateStatus(ctx, id, "Cancelled")
	if err != nil {
		return err
	}

	uc.cache.Delete(ctx, id)

	return nil
}

func (uc *OrderUseCase) GetRevenueByCustomer(ctx context.Context, customerID string) (int64, int, error) {
	if customerID == "" {
		return 0, 0, errors.New("customer ID is required")
	}
	totalAmount, orderCount, err := uc.repo.GetRevenueByCustomer(ctx, customerID)
	if err != nil {
		return 0, 0, err
	}
	if orderCount == 0 {
		return 0, 0, err
	}
	return totalAmount, orderCount, nil
}
