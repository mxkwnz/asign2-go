package http

import (
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc *usecase.PaymentUseCase
}

func NewHandler(uc *usecase.PaymentUseCase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) CreatePayment(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id"`
		Amount  int64  `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.uc.CreatePayment(c, req.OrderID, req.Amount)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if payment.Status == "Declined" {
		c.JSON(400, payment)
		return
	}

	c.JSON(200, payment)
}

func (h *Handler) GetPayment(c *gin.Context) {
	orderID := c.Param("order_id")

	payment, err := h.uc.GetByOrderID(c, orderID)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	c.JSON(200, payment)
}
