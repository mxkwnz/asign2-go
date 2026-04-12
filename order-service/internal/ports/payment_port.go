package ports

import "context"

type PaymentPort interface {
	Pay(ctx context.Context, orderID string, amount int64) (string, error)
}
