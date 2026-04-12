package ports

import (
	"context"
	"log"
	"os"

	paymentv1 "github.com/mxkwnz/ap2-generated/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCPaymentClient struct {
	client paymentv1.PaymentServiceClient
}

func NewGRPCPaymentClient() (*GRPCPaymentClient, error) {
	addr := os.Getenv("PAYMENT_GRPC_ADDR")
	if addr == "" {
		addr = "localhost:9091"
	}
	log.Printf("Connecting to payment gRPC at %s", addr)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCPaymentClient{client: paymentv1.NewPaymentServiceClient(conn)}, nil
}

func (c *GRPCPaymentClient) Pay(ctx context.Context, orderID string, amount int64) (string, error) {
	log.Printf("Calling payment gRPC for order %s amount %d", orderID, amount)
	resp, err := c.client.ProcessPayment(ctx, &paymentv1.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		log.Printf("Payment gRPC error: %v", err)
		return "", err
	}
	log.Printf("Payment gRPC response: %s", resp.Status)
	return resp.Status, nil
}
