package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/example/ecommerce-api/internal/models"
	"github.com/example/ecommerce-api/internal/store"
)

type AdminOrdersHandler struct {
	store store.Store
}

func NewAdminOrdersHandler(st store.Store) *AdminOrdersHandler {
	return &AdminOrdersHandler{store: st}
}

type adminOrderResp struct {
	OrderID    string            `json:"order_id"`
	UserID     string            `json:"user_id"`
	Status     string            `json:"status"`
	Amount     int64             `json:"amount_cents"`
	PaymentRef string            `json:"payment_ref,omitempty"`
	Items      []models.CartItem `json:"items,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

type orderStatusReq struct {
	Status string `json:"status"`
}

func (h *AdminOrdersHandler) List(c *gin.Context) {
	orders, err := h.store.ListOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]adminOrderResp, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, adminOrderResp{
			OrderID:    o.ID,
			UserID:     o.UserID,
			Status:     o.Status,
			Amount:     o.Amount,
			PaymentRef: o.PaymentRef,
			Items:      o.Items,
			CreatedAt:  o.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AdminOrdersHandler) UpdateStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order id required"})
		return
	}

	var req orderStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status required"})
		return
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status required"})
		return
	}

	switch status {
	case "pending", "paid", "failed", "done", "completed":
		if status == "completed" {
			status = "done"
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	if err := h.store.UpdateOrderStatus(orderID, status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
