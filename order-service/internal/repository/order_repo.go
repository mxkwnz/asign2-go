package repository

import (
	"context"
	"database/sql"
	"order-service/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, o domain.Order) error
	GetByID(ctx context.Context, id string) (domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	GetByIdempotencyKey(ctx context.Context, key string) (domain.Order, error)
	GetRevenueByCustomer(ctx context.Context, customerID string) (int64, int, error)
}

type orderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(ctx context.Context, o domain.Order) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO orders (id, customer_id, item_name, amount, status, created_at, idempotency_key) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		o.ID, o.CustomerID, o.ItemName, o.Amount, o.Status, o.CreatedAt, o.IdempotencyKey,
	)
	return err
}

func (r *orderRepo) GetByID(ctx context.Context, id string) (domain.Order, error) {
	var o domain.Order
	err := r.db.QueryRowContext(ctx,
		"SELECT id, customer_id, item_name, amount, status, created_at, idempotency_key FROM orders WHERE id=$1",
		id,
	).Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt, &o.IdempotencyKey)
	return o, err
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET status=$1 WHERE id=$2",
		status, id,
	)
	return err
}

func (r *orderRepo) GetByIdempotencyKey(ctx context.Context, key string) (domain.Order, error) {
	var o domain.Order
	err := r.db.QueryRowContext(ctx,
		"SELECT id, customer_id, item_name, amount, status, created_at, idempotency_key FROM orders WHERE idempotency_key=$1",
		key,
	).Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt, &o.IdempotencyKey)
	return o, err
}

func (r *orderRepo) GetRevenueByCustomer(ctx context.Context, customerID string) (int64, int, error) {
	var revenue int64
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM orders WHERE customer_id = $1 AND status = $2",
		customerID, "Paid",
	).Scan(&revenue, &count)

	if err != nil {
		return 0, 0, err
	}
	return revenue, count, err
}
