package models

import "time"

// User represents an account in the system
type User struct {
	ID            string    `json:"id"`
	FullName      string    `json:"full_name"`
	Phone         string    `json:"phone"`
	Email         string    `json:"email"`
	Password      string    `json:"-"`
	Role          string    `json:"role"` // "user" or "admin"
	AuthProvider  string    `json:"auth_provider"`
	GoogleID      string    `json:"google_id,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// EmailVerification token
type EmailVerification struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

// PasswordReset token
type PasswordReset struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}
