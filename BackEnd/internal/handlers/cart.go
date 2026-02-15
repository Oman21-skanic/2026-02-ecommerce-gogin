package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/ecommerce-api/internal/middleware"
	"github.com/example/ecommerce-api/internal/store"
)

type CartHandler struct {
	store store.Store
}

func NewCartHandler(st store.Store) *CartHandler {
	return &CartHandler{store: st}
}

type cartUpdateReq struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (h *CartHandler) Add(c *gin.Context) {
	var req cartUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil || req.ProductID == "" || req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	userID := c.GetString(string(middleware.UserIDKey))
	if err := h.store.AddToCart(userID, req.ProductID, req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CartHandler) Remove(c *gin.Context) {
	var req cartUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil || req.ProductID == "" || req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	userID := c.GetString(string(middleware.UserIDKey))
	if err := h.store.RemoveFromCart(userID, req.ProductID, req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CartHandler) View(c *gin.Context) {
	userID := c.GetString(string(middleware.UserIDKey))
	cart, err := h.store.GetCart(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cart)
}
