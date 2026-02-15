package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/example/ecommerce-api/internal/config"
	"github.com/example/ecommerce-api/internal/email"
	"github.com/example/ecommerce-api/internal/middleware"
	"github.com/example/ecommerce-api/internal/models"
	"github.com/example/ecommerce-api/internal/payment"
	"github.com/example/ecommerce-api/internal/store"
)

type CheckoutHandler struct {
	cfg          *config.Config
	store        store.Store
	pay          payment.Gateway
	emailService *email.Service
}

func NewCheckoutHandler(cfg *config.Config, st store.Store, gw payment.Gateway, es *email.Service) *CheckoutHandler {
	return &CheckoutHandler{
		cfg:          cfg,
		store:        st,
		pay:          gw,
		emailService: es,
	}
}

type checkoutReq struct {
	PaymentMethod string `json:"payment_method"` // gopay, shopeepay, qris, bank_transfer
}

type orderResp struct {
	OrderID     string            `json:"order_id"`
	Status      string            `json:"status"`
	Amount      int64             `json:"amount_cents"`
	PaymentRef  string            `json:"payment_ref,omitempty"`
	PaymentURL  string            `json:"payment_url,omitempty"`
	RedirectURL string            `json:"redirect_url,omitempty"`
	Items       []models.CartItem `json:"items,omitempty"`
	CreatedAt   time.Time         `json:"created_at,omitempty"`
}

func (h *CheckoutHandler) Checkout(c *gin.Context) {
	var req checkoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment method required"})
		return
	}

	userID := c.GetString(string(middleware.UserIDKey))
	email := c.GetString(string(middleware.EmailKey))

	// Get user cart
	cart, err := h.store.GetCart(userID)
	if err != nil || len(cart.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keranjang kosong"})
		return
	}

	// Calculate total amount and get product details
	var amount int64
	var itemsStr string
	midtransItems := []payment.MidtransItem{}

	for _, it := range cart.Items {
		p, err := h.store.GetProduct(it.ProductID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product in cart: " + it.ProductID})
			return
		}

		// Check stock
		if p.Stock < it.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Stock tidak cukup untuk %s (tersedia: %d)", p.Name, p.Stock),
			})
			return
		}

		itemAmount := int64(it.Quantity) * p.PriceCents
		amount += itemAmount

		// Build items string for email
		itemsStr += fmt.Sprintf("- %s x%d = Rp %d\n", p.Name, it.Quantity, itemAmount/100)

		// Build Midtrans items
		midtransItems = append(midtransItems, payment.MidtransItem{
			ID:       p.ID,
			Price:    p.PriceCents,
			Quantity: it.Quantity,
			Name:     p.Name,
		})
	}

	// Create order (status: pending)
	o, err := h.store.CreateOrder(userID, cart.Items, amount, "pending", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat order: " + err.Error()})
		return
	}

	// Process payment with Midtrans
	ctx := context.Background()
	metadata := map[string]string{
		"order_id":      o.ID,
		"email":         email,
		"customer_name": email, // Could get from user profile
	}

	// If using Midtrans, try to create Snap transaction
	var paymentURL string
	var paymentRef string

	if midtransGw, ok := h.pay.(*payment.MidtransGateway); ok {
		// Use Snap for better UX
		snapResp, err := midtransGw.CreateSnapTransaction(ctx, o.ID, amount, email, midtransItems)
		if err != nil {
			// Fallback to direct charge
			paymentRef, err = h.pay.Charge(ctx, amount, req.PaymentMethod, metadata)
			if err != nil {
				c.JSON(http.StatusPaymentRequired, gin.H{"error": "Payment gagal: " + err.Error()})
				return
			}
		} else {
			paymentURL = snapResp.PaymentURL
			paymentRef = snapResp.OrderID
		}
	} else {
		// Mock gateway
		paymentRef, err = h.pay.Charge(ctx, amount, req.PaymentMethod, metadata)
		if err != nil {
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "Payment gagal"})
			return
		}
	}

	// Update order with payment reference
	_ = h.store.UpdateOrderPaymentRef(o.ID, paymentRef)

	// Clear cart
	h.store.ClearCart(userID)

	// Send order confirmation email (async)
	if h.emailService != nil {
		go func() {
			_ = h.emailService.SendOrderConfirmation(email, o.ID, amount, itemsStr)
		}()
	}

	response := orderResp{
		OrderID:    o.ID,
		Status:     "pending",
		Amount:     amount,
		PaymentRef: paymentRef,
		Items:      o.Items,
		CreatedAt:  o.CreatedAt,
	}

	if paymentURL != "" {
		response.PaymentURL = paymentURL
		response.RedirectURL = paymentURL
	}

	c.JSON(http.StatusOK, response)
}

func (h *CheckoutHandler) MyOrders(c *gin.Context) {
	userID := c.GetString(string(middleware.UserIDKey))
	orders, err := h.store.ListOrdersByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]orderResp, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, orderResp{
			OrderID:    o.ID,
			Status:     o.Status,
			Amount:     o.Amount,
			PaymentRef: o.PaymentRef,
			Items:      o.Items,
			CreatedAt:  o.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, resp)
}

// MidtransCallback handles payment notifications from Midtrans
func (h *CheckoutHandler) MidtransCallback(c *gin.Context) {
	var notification map[string]interface{}
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification"})
		return
	}

	orderID, _ := notification["order_id"].(string)
	transactionStatus, _ := notification["transaction_status"].(string)
	fraudStatus, _ := notification["fraud_status"].(string)

	// Map Midtrans status to our order status
	var status string
	switch transactionStatus {
	case "capture":
		if fraudStatus == "accept" {
			status = "paid"
		}
	case "settlement":
		status = "paid"
	case "pending":
		status = "pending"
	case "deny", "cancel", "expire":
		status = "failed"
	default:
		status = "pending"
	}

	// Update order status
	if err := h.store.UpdateOrderStatus(orderID, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}
