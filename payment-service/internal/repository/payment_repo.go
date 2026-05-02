package repository

import (
	"context"
	"database/sql"
	"fmt"
	"payment-service/internal/domain"
)

type PaymentRepository interface {
	Save(ctx context.Context, p domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error)
	FindByAmountRange(ctx context.Context, min, max int64) ([]*domain.Payment, error)
	ListByAmountRange(ctx context.Context, minAmount, maxAmount int64) ([]domain.Payment, error)
}

type paymentRepo struct {
	db *sql.DB
}

func NewPaymentRepo(db *sql.DB) PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Save(ctx context.Context, p domain.Payment) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO payments (id, order_id, transaction_id, amount, status) VALUES ($1,$2,$3,$4,$5)",
		p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status,
	)
	return err
}

func (r *paymentRepo) GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error) {
	var p domain.Payment
	err := r.db.QueryRowContext(ctx,
		"SELECT id, order_id, transaction_id, amount, status FROM payments WHERE order_id=$1",
		orderID,
	).Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status)
	return p, err
}

func (r *paymentRepo) FindByAmountRange(ctx context.Context, min, max int64) ([]*domain.Payment, error) {
	query := "SELECT id, order_id, transaction_id, amount, status FROM payments WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if min > 0 {
		query += fmt.Sprintf(" AND amount >= $%d", argIdx)
		args = append(args, min)
		argIdx++
	}
	if max > 0 {
		query += fmt.Sprintf(" AND amount <= $%d", argIdx)
		args = append(args, max)
		argIdx++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status); err != nil {
			return nil, err
		}
		payments = append(payments, &p)
	}
	return payments, rows.Err()
}

func (r *paymentRepo) ListByAmountRange(ctx context.Context, minAmount, maxAmount int64) ([]domain.Payment, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, order_id, transaction_id, amount, status FROM payments WHERE amount >= $1 AND amount <= $2",
		minAmount, maxAmount,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}
