package grpc

import (
	"database/sql"
	"time"

	orderv1 "github.com/mxkwnz/ap2-generated/order/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGRPCServer struct {
	orderv1.UnimplementedOrderServiceServer
	db *sql.DB
}

func NewOrderGRPCServer(db *sql.DB) *OrderGRPCServer {
	return &OrderGRPCServer{db: db}
}

func (s *OrderGRPCServer) SubscribeToOrderUpdates(req *orderv1.OrderRequest, stream orderv1.OrderService_SubscribeToOrderUpdatesServer) error {
	if req.OrderId == "" {
		return status.Error(codes.InvalidArgument, "order_id required")
	}

	var lastStatus string
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
		}

		var currentStatus string
		err := s.db.QueryRowContext(stream.Context(),
			"SELECT status FROM orders WHERE id = $1", req.OrderId,
		).Scan(&currentStatus)

		if err != nil {
			return status.Error(codes.NotFound, "order not found")
		}

		if currentStatus != lastStatus {
			lastStatus = currentStatus
			if err := stream.Send(&orderv1.OrderStatusUpdate{
				OrderId: req.OrderId,
				Status:  currentStatus,
				Message: "Status updated to " + currentStatus,
			}); err != nil {
				return err
			}
		}

		time.Sleep(2 * time.Second)
	}
}
