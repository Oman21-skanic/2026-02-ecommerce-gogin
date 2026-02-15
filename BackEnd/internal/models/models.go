package models

import "time"

// Product describes a sellable item
type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"` // Tambahan: Agar bisa difilter
	PriceCents  int64  `json:"price_cents"`
	SKU         string `json:"sku"`
	Stock       int    `json:"stock"`
	// Tambahan: Untuk tampilan UI yang cantik
	Thumbnail   string    `json:"thumbnail"`    // Gambar utama
	Images      string    `json:"images"`       // Disimpan sebagai string (JSON/Comma separated)
	Rating      float64   `json:"rating"`       // Dummy rating: 4.5
	ReviewCount int       `json:"review_count"` // Dummy count: 120 ulasan
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Cart and nested items
type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type Cart struct {
	UserID string     `json:"user_id"`
	Items  []CartItem `json:"items"`
}

// Checkout
type CheckoutRequest struct {
	PaymentMethod string `json:"payment_method"` // e.g., "card"
}

type Order struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Items      []CartItem `json:"items"`
	Amount     int64      `json:"amount_cents"`
	Status     string     `json:"status"` // pending, paid, failed
	PaymentRef string     `json:"payment_ref"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Review represents a user review for the coffeehouse
type Review struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	UserPhoto string    `json:"user_photo"`
	Rating    int       `json:"rating"` // 1-5
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}
