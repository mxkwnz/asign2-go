package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"payment-service/internal/usecase"

	paymentv1 "github.com/mxkwnz/ap2-generated/payment/v1"
)

type PaymentGRPCServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewPaymentGRPCServer(uc *usecase.PaymentUseCase) *PaymentGRPCServer {
	return &PaymentGRPCServer{uc: uc}
}

func (s *PaymentGRPCServer) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	payment, err := s.uc.CreatePayment(ctx, req.OrderId, req.Amount)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &paymentv1.PaymentResponse{
		TransactionId: payment.TransactionID,
		Status:        payment.Status,
		ProcessedAt:   timestamppb.New(time.Now()),
	}, nil
}
