package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/example/ecommerce-api/internal/store"
)

type ReviewsHandler struct {
	store store.Store
}

func NewReviewsHandler(st store.Store) *ReviewsHandler {
	return &ReviewsHandler{store: st}
}

type CreateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"required"`
}

// Create handles POST /api/v1/me/reviews
func (h *ReviewsHandler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user has placed at least one order (optional validation)
	orders, err := h.store.ListOrdersByUser(userID.(string))
	if err != nil || len(orders) == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "you must complete at least one order to leave a review"})
		return
	}

	// Get user info for display name
	user, err := h.store.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	// Limit reviews per user (optional: 1 review per user)
	reviewCount, _ := h.store.GetUserReviewCount(userID.(string))
	if reviewCount > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "you have already submitted a review"})
		return
	}

	// Default user photo placeholder
	userPhoto := "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&w=240&q=80"

	review, err := h.store.CreateReview(userID.(string), user.Email, userPhoto, req.Rating, req.Comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// List handles GET /api/v1/reviews
func (h *ReviewsHandler) List(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	reviews, err := h.store.ListReviews(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reviews)
}
