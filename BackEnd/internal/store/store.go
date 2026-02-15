package store

import (
	"time"

	"github.com/example/ecommerce-api/internal/models"
)

// Store abstracts data storage backends
type Store interface {
	// Users
	CreateUser(fullName, phone, email, passwordHash, role, provider, googleID string, emailVerified bool) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	SeedAdminUser(email, passwordHash string) error
	UpdateUserPassword(userID, newPasswordHash string) error
	MarkEmailVerified(userID string) error

	// Email Verification
	CreateEmailVerification(userID, token string, expiresAt time.Time) error
	GetEmailVerification(token string) (*models.EmailVerification, error)
	GetLatestEmailVerificationByUser(userID string) (*models.EmailVerification, error)
	DeleteEmailVerification(token string) error

	// Password Reset
	CreatePasswordReset(userID, token string, expiresAt time.Time) error
	GetPasswordReset(token string) (*models.PasswordReset, error)
	DeletePasswordReset(token string) error

	// Products
	CreateProduct(p *models.Product) (*models.Product, error)
	UpdateProduct(id string, update func(p *models.Product) error) (*models.Product, error)
	DeleteProduct(id string) error
	GetProduct(id string) (*models.Product, error)
	ListProducts(query string) ([]*models.Product, error)

	// Carts
	GetOrCreateCart(userID string) *models.Cart
	AddToCart(userID, productID string, qty int) error
	RemoveFromCart(userID, productID string, qty int) error
	ClearCart(userID string)
	GetCart(userID string) (*models.Cart, error)

	// Orders
	CreateOrder(userID string, items []models.CartItem, amount int64, status, paymentRef string) (*models.Order, error)
	ListOrdersByUser(userID string) ([]*models.Order, error)
	ListOrders() ([]*models.Order, error)
	UpdateOrderStatus(orderID, status string) error
	UpdateOrderPaymentRef(orderID, paymentRef string) error

	// Reviews
	CreateReview(userID, userName, userPhoto string, rating int, comment string) (*models.Review, error)
	ListReviews(limit int) ([]*models.Review, error)
	GetUserReviewCount(userID string) (int, error)
}
