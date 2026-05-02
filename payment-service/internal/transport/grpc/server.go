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
		Amount:        payment.Amount,
	}, nil
}

func (s *PaymentGRPCServer) ListPayments(ctx context.Context, req *paymentv1.ListPaymentsRequest) (*paymentv1.ListPaymentsResponse, error) {
	payments, err := s.uc.ListPayments(ctx, req.MinAmount, req.MaxAmount)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var responses []*paymentv1.PaymentResponse
	for _, p := range payments {
		responses = append(responses, &paymentv1.PaymentResponse{
			TransactionId: p.TransactionID,
			Status:        p.Status,
			Amount:        p.Amount,
			ProcessedAt:   timestamppb.New(time.Now()),
		})
	}

	return &paymentv1.ListPaymentsResponse{Payments: responses}, nil
}
