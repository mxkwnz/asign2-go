package http

import (
	"order-service/internal/domain"
	"order-service/internal/usecase"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	uc *usecase.OrderUseCase
}

func NewHandler(uc *usecase.OrderUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req struct {
		CustomerID string `json:"customer_id"`
		ItemName   string `json:"item_name"`
		Amount     int64  `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	key := c.GetHeader("Idempotency-Key")

	order := domain.Order{
		ID:             uuid.New().String(),
		CustomerID:     req.CustomerID,
		ItemName:       req.ItemName,
		Amount:         req.Amount,
		CreatedAt:      time.Now(),
		IdempotencyKey: key,
	}

	result, err := h.uc.CreateOrder(c, order)

	if err != nil {
		c.JSON(503, gin.H{"error": "payment service unavailable"})
		return
	}

	c.JSON(201, result)
}

func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.uc.GetOrder(c, id)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	c.JSON(200, order)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	id := c.Param("id")

	err := h.uc.CancelOrder(c, id)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "cancelled"})
}

func (h *Handler) GetRevenue(c *gin.Context) {
	customerID := c.Query("customer_id")

	totalAmount, orderCount, err := h.uc.GetRevenueByCustomer(c, customerID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"customer_id":  customerID,
		"total_amount": totalAmount,
		"order_count":  orderCount,
	})
}
