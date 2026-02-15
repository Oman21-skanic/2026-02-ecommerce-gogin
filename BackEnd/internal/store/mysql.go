package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"github.com/example/ecommerce-api/internal/models"
)

type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(dsn string) (*MySQLStore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	ms := &MySQLStore{db: db}
	if err := ms.autoMigrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return ms, nil
}

func (s *MySQLStore) autoMigrate() error {
	// Create tables
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id CHAR(36) PRIMARY KEY,
			full_name VARCHAR(255) NOT NULL DEFAULT '',
			phone VARCHAR(50) NOT NULL DEFAULT '',
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'user',
			auth_provider VARCHAR(30) NOT NULL DEFAULT 'email',
			google_id VARCHAR(255) NOT NULL DEFAULT '',
			email_verified BOOLEAN NOT NULL DEFAULT FALSE,
			created_at DATETIME NOT NULL,
			INDEX idx_email (email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

		`CREATE TABLE IF NOT EXISTS email_verifications (
			token CHAR(36) PRIMARY KEY,
			user_id CHAR(36) NOT NULL,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_id (user_id),
			INDEX idx_expires_at (expires_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

		`CREATE TABLE IF NOT EXISTS password_resets (
			token CHAR(36) PRIMARY KEY,
			user_id CHAR(36) NOT NULL,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_user_id (user_id),
			INDEX idx_expires_at (expires_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

		`CREATE TABLE IF NOT EXISTS products (
			id CHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			category VARCHAR(100),
			price_cents BIGINT NOT NULL,
			sku VARCHAR(100) NOT NULL,
			stock INT NOT NULL,
			thumbnail TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			INDEX idx_sku (sku),
			INDEX idx_name (name),
			INDEX idx_category (category)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

		`CREATE TABLE IF NOT EXISTS carts (
			user_id CHAR(36) PRIMARY KEY,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

		`CREATE TABLE IF NOT EXISTS cart_items (
			user_id CHAR(36) NOT NULL,
			product_id CHAR(36) NOT NULL,
			quantity INT NOT NULL,
			PRIMARY KEY (user_id, product_id),
			FOREIGN KEY (user_id) REFERENCES carts(user_id) ON DELETE CASCADE,
			FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

		`CREATE TABLE IF NOT EXISTS orders (
			id CHAR(36) PRIMARY KEY,
			user_id CHAR(36) NOT NULL,
			amount_cents BIGINT NOT NULL,
			status VARCHAR(20) NOT NULL,
			payment_ref VARCHAR(255) NOT NULL,
			created_at DATETIME NOT NULL,
			INDEX idx_user_id (user_id),
			INDEX idx_created_at (created_at),
			FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

		`CREATE TABLE IF NOT EXISTS order_items (
			order_id CHAR(36) NOT NULL,
			product_id CHAR(36) NOT NULL,
			quantity INT NOT NULL,
			price_cents BIGINT NOT NULL,
			PRIMARY KEY (order_id, product_id),
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
			FOREIGN KEY (product_id) REFERENCES products(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,

		`CREATE TABLE IF NOT EXISTS reviews (
			id CHAR(36) PRIMARY KEY,
			user_id CHAR(36) NOT NULL,
			user_name VARCHAR(255) NOT NULL,
			user_photo TEXT,
			rating INT NOT NULL,
			comment TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			INDEX idx_reviews_created_at (created_at),
			INDEX idx_reviews_user_id (user_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`,
	}

	for _, st := range stmts {
		if _, err := s.db.Exec(st); err != nil {
			// Ignore "table already exists" errors
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}

	// Check and add email_verified column if it doesn't exist
	var emailVerifiedCount int
	row := s.db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'email_verified'")
	if err := row.Scan(&emailVerifiedCount); err != nil {
		return err
	}

	if emailVerifiedCount == 0 {
		_, err := s.db.Exec(`ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE AFTER role`)
		if err != nil && !strings.Contains(err.Error(), "Duplicate column") {
			return err
		}
	}

	// Check and add role column if it doesn't exist
	var roleCount int
	row = s.db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'role'")
	if err := row.Scan(&roleCount); err != nil {
		return err
	}

	if roleCount == 0 {
		_, err := s.db.Exec(`ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user' AFTER password_hash`)
		if err != nil && !strings.Contains(err.Error(), "Duplicate column") {
			return err
		}
	}

	var fullNameCount int
	row = s.db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'full_name'")
	if err := row.Scan(&fullNameCount); err != nil {
		return err
	}
	if fullNameCount == 0 {
		_, err := s.db.Exec(`ALTER TABLE users ADD COLUMN full_name VARCHAR(255) NOT NULL DEFAULT '' AFTER id`)
		if err != nil && !strings.Contains(err.Error(), "Duplicate column") {
			return err
		}
	}

	var phoneCount int
	row = s.db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'phone'")
	if err := row.Scan(&phoneCount); err != nil {
		return err
	}
	if phoneCount == 0 {
		_, err := s.db.Exec(`ALTER TABLE users ADD COLUMN phone VARCHAR(50) NOT NULL DEFAULT '' AFTER full_name`)
		if err != nil && !strings.Contains(err.Error(), "Duplicate column") {
			return err
		}
	}

	var authProviderCount int
	row = s.db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'auth_provider'")
	if err := row.Scan(&authProviderCount); err != nil {
		return err
	}
	if authProviderCount == 0 {
		_, err := s.db.Exec(`ALTER TABLE users ADD COLUMN auth_provider VARCHAR(30) NOT NULL DEFAULT 'email' AFTER role`)
		if err != nil && !strings.Contains(err.Error(), "Duplicate column") {
			return err
		}
	}

	var googleIDCount int
	row = s.db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'google_id'")
	if err := row.Scan(&googleIDCount); err != nil {
		return err
	}
	if googleIDCount == 0 {
		_, err := s.db.Exec(`ALTER TABLE users ADD COLUMN google_id VARCHAR(255) NOT NULL DEFAULT '' AFTER auth_provider`)
		if err != nil && !strings.Contains(err.Error(), "Duplicate column") {
			return err
		}
	}

	var categoryCount int
	row = s.db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'products' AND COLUMN_NAME = 'category'")
	if err := row.Scan(&categoryCount); err != nil {
		return err
	}
	if categoryCount == 0 {
		_, err := s.db.Exec(`ALTER TABLE products ADD COLUMN category VARCHAR(100) AFTER description`)
		if err != nil && !strings.Contains(err.Error(), "Duplicate column") {
			return err
		}
	}

	var thumbnailCount int
	row = s.db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'products' AND COLUMN_NAME = 'thumbnail'")
	if err := row.Scan(&thumbnailCount); err != nil {
		return err
	}
	if thumbnailCount == 0 {
		_, err := s.db.Exec(`ALTER TABLE products ADD COLUMN thumbnail TEXT AFTER stock`)
		if err != nil && !strings.Contains(err.Error(), "Duplicate column") {
			return err
		}
	}

	return nil
}

// Users

func (s *MySQLStore) CreateUser(fullName, phone, email, passwordHash, role, provider, googleID string, emailVerified bool) (*models.User, error) {
	id := uuid.NewString()
	now := time.Now()
	if role == "" {
		role = "user"
	}
	if provider == "" {
		provider = "email"
	}

	_, err := s.db.Exec(
		`INSERT INTO users (id, full_name, phone, email, password_hash, role, auth_provider, google_id, email_verified, created_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		id, fullName, phone, email, passwordHash, role, provider, googleID, emailVerified, now,
	)
	if err != nil {
		if isDuplicate(err) {
			return nil, errors.New("email already registered")
		}
		return nil, err
	}

	return &models.User{
		ID:            id,
		FullName:      fullName,
		Phone:         phone,
		Email:         email,
		Password:      passwordHash,
		Role:          role,
		AuthProvider:  provider,
		GoogleID:      googleID,
		EmailVerified: emailVerified,
		CreatedAt:     now,
	}, nil
}

func (s *MySQLStore) GetUserByEmail(email string) (*models.User, error) {
	row := s.db.QueryRow(
		`SELECT id, full_name, phone, email, password_hash, role, auth_provider, google_id, email_verified, created_at FROM users WHERE email=?`,
		email,
	)

	u := models.User{}
	if err := row.Scan(&u.ID, &u.FullName, &u.Phone, &u.Email, &u.Password, &u.Role, &u.AuthProvider, &u.GoogleID, &u.EmailVerified, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (s *MySQLStore) GetUserByID(id string) (*models.User, error) {
	row := s.db.QueryRow(
		`SELECT id, full_name, phone, email, password_hash, role, auth_provider, google_id, email_verified, created_at FROM users WHERE id=?`,
		id,
	)

	u := models.User{}
	if err := row.Scan(&u.ID, &u.FullName, &u.Phone, &u.Email, &u.Password, &u.Role, &u.AuthProvider, &u.GoogleID, &u.EmailVerified, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (s *MySQLStore) SeedAdminUser(email, passwordHash string) error {
	row := s.db.QueryRow(`SELECT id FROM users WHERE email=?`, email)
	var id string
	err := row.Scan(&id)
	if err == nil {
		return nil // already exists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	adminID := uuid.NewString()
	now := time.Now()
	_, err = s.db.Exec(
		`INSERT INTO users (id, full_name, phone, email, password_hash, role, auth_provider, google_id, email_verified, created_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		adminID, "Administrator", "", email, passwordHash, "admin", "email", "", true, now,
	)
	if err != nil {
		if isDuplicate(err) {
			return nil // race condition, already created
		}
		return err
	}
	return nil
}

func (s *MySQLStore) UpdateUserPassword(userID, newPasswordHash string) error {
	res, err := s.db.Exec(`UPDATE users SET password_hash=? WHERE id=?`, newPasswordHash, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (s *MySQLStore) MarkEmailVerified(userID string) error {
	res, err := s.db.Exec(`UPDATE users SET email_verified=TRUE WHERE id=?`, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("user not found")
	}
	return nil
}

// Email Verification

func (s *MySQLStore) CreateEmailVerification(userID, token string, expiresAt time.Time) error {
	_, err := s.db.Exec(
		`INSERT INTO email_verifications (token, user_id, expires_at) VALUES (?,?,?)`,
		token, userID, expiresAt,
	)
	return err
}

func (s *MySQLStore) GetEmailVerification(token string) (*models.EmailVerification, error) {
	row := s.db.QueryRow(
		`SELECT token, user_id, expires_at FROM email_verifications WHERE token=?`,
		token,
	)

	v := models.EmailVerification{}
	if err := row.Scan(&v.Token, &v.UserID, &v.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("verification token not found")
		}
		return nil, err
	}
	return &v, nil
}

func (s *MySQLStore) GetLatestEmailVerificationByUser(userID string) (*models.EmailVerification, error) {
	row := s.db.QueryRow(
		`SELECT token, user_id, expires_at FROM email_verifications WHERE user_id=? ORDER BY expires_at DESC LIMIT 1`,
		userID,
	)

	v := models.EmailVerification{}
	if err := row.Scan(&v.Token, &v.UserID, &v.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("verification token not found")
		}
		return nil, err
	}
	return &v, nil
}

func (s *MySQLStore) DeleteEmailVerification(token string) error {
	_, err := s.db.Exec(`DELETE FROM email_verifications WHERE token=?`, token)
	return err
}

// Password Reset

func (s *MySQLStore) CreatePasswordReset(userID, token string, expiresAt time.Time) error {
	_, err := s.db.Exec(
		`INSERT INTO password_resets (token, user_id, expires_at) VALUES (?,?,?)`,
		token, userID, expiresAt,
	)
	return err
}

func (s *MySQLStore) GetPasswordReset(token string) (*models.PasswordReset, error) {
	row := s.db.QueryRow(
		`SELECT token, user_id, expires_at FROM password_resets WHERE token=?`,
		token,
	)

	r := models.PasswordReset{}
	if err := row.Scan(&r.Token, &r.UserID, &r.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reset token not found")
		}
		return nil, err
	}
	return &r, nil
}

func (s *MySQLStore) DeletePasswordReset(token string) error {
	_, err := s.db.Exec(`DELETE FROM password_resets WHERE token=?`, token)
	return err
}

// Products

func (s *MySQLStore) CreateProduct(p *models.Product) (*models.Product, error) {
	p.ID = uuid.NewString()
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := s.db.Exec(
		`INSERT INTO products (id, name, description, category, price_cents, sku, stock, thumbnail, created_at, updated_at) 
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		p.ID, p.Name, p.Description, p.Category, p.PriceCents, p.SKU, p.Stock, p.Thumbnail, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *MySQLStore) UpdateProduct(id string, updateFn func(p *models.Product) error) (*models.Product, error) {
	row := s.db.QueryRow(
		`SELECT id, name, description, category, price_cents, sku, stock, thumbnail, created_at, updated_at 
		FROM products WHERE id=?`,
		id,
	)

	p := models.Product{}
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.PriceCents, &p.SKU, &p.Stock, &p.Thumbnail, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	if err := updateFn(&p); err != nil {
		return nil, err
	}

	p.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE products SET name=?, description=?, category=?, price_cents=?, sku=?, stock=?, thumbnail=?, updated_at=? WHERE id=?`,
		p.Name, p.Description, p.Category, p.PriceCents, p.SKU, p.Stock, p.Thumbnail, p.UpdatedAt, p.ID,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *MySQLStore) DeleteProduct(id string) error {
	res, err := s.db.Exec(`DELETE FROM products WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("product not found")
	}
	return nil
}

func (s *MySQLStore) GetProduct(id string) (*models.Product, error) {
	row := s.db.QueryRow(
		`SELECT id, name, description, category, price_cents, sku, stock, thumbnail, created_at, updated_at 
		FROM products WHERE id=?`,
		id,
	)

	p := models.Product{}
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.PriceCents, &p.SKU, &p.Stock, &p.Thumbnail, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	return &p, nil
}

func (s *MySQLStore) ListProducts(query string) ([]*models.Product, error) {
	var rows *sql.Rows
	var err error

	if strings.TrimSpace(query) == "" {
		rows, err = s.db.Query(
			`SELECT id, name, description, category, price_cents, sku, stock, thumbnail, created_at, updated_at 
			FROM products ORDER BY created_at DESC`,
		)
	} else {
		like := "%" + strings.ToLower(query) + "%"
		rows, err = s.db.Query(
			`SELECT id, name, description, category, price_cents, sku, stock, thumbnail, created_at, updated_at 
			FROM products 
			WHERE LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(sku) LIKE ? OR LOWER(category) LIKE ? 
			ORDER BY created_at DESC`,
			like, like, like, like,
		)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []*models.Product{}
	for rows.Next() {
		p := models.Product{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.PriceCents, &p.SKU, &p.Stock, &p.Thumbnail, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		cp := p
		res = append(res, &cp)
	}
	return res, nil
}

// Carts

func (s *MySQLStore) GetOrCreateCart(userID string) *models.Cart {
	_, _ = s.db.Exec(
		`INSERT IGNORE INTO carts (user_id, updated_at) VALUES (?, ?)`,
		userID, time.Now(),
	)

	c, _ := s.GetCart(userID)
	if c == nil {
		return &models.Cart{UserID: userID, Items: []models.CartItem{}}
	}
	return c
}

func (s *MySQLStore) AddToCart(userID, productID string, qty int) error {
	if qty <= 0 {
		return errors.New("quantity must be positive")
	}

	p, err := s.GetProduct(productID)
	if err != nil {
		return err
	}

	// Check current cart quantity
	row := s.db.QueryRow(
		`SELECT COALESCE(quantity, 0) FROM cart_items WHERE user_id=? AND product_id=?`,
		userID, productID,
	)
	var currentQty int
	row.Scan(&currentQty)

	if p.Stock < (currentQty + qty) {
		return errors.New("insufficient stock")
	}

	_, err = s.db.Exec(
		`INSERT INTO carts (user_id, updated_at) VALUES (?, ?) 
		ON DUPLICATE KEY UPDATE updated_at=VALUES(updated_at)`,
		userID, time.Now(),
	)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		`INSERT INTO cart_items (user_id, product_id, quantity) VALUES (?,?,?) 
		ON DUPLICATE KEY UPDATE quantity = quantity + VALUES(quantity)`,
		userID, productID, qty,
	)
	return err
}

func (s *MySQLStore) RemoveFromCart(userID, productID string, qty int) error {
	if qty <= 0 {
		return errors.New("quantity must be positive")
	}

	row := s.db.QueryRow(
		`SELECT quantity FROM cart_items WHERE user_id=? AND product_id=?`,
		userID, productID,
	)
	var cur int
	err := row.Scan(&cur)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("item not in cart")
		}
		return err
	}

	if cur <= qty {
		_, err = s.db.Exec(
			`DELETE FROM cart_items WHERE user_id=? AND product_id=?`,
			userID, productID,
		)
		return err
	}

	_, err = s.db.Exec(
		`UPDATE cart_items SET quantity = quantity - ? WHERE user_id=? AND product_id=?`,
		qty, userID, productID,
	)
	return err
}

func (s *MySQLStore) ClearCart(userID string) {
	_, _ = s.db.Exec(`DELETE FROM cart_items WHERE user_id=?`, userID)
	_, _ = s.db.Exec(`DELETE FROM carts WHERE user_id=?`, userID)
}

func (s *MySQLStore) GetCart(userID string) (*models.Cart, error) {
	rows, err := s.db.Query(
		`SELECT product_id, quantity FROM cart_items WHERE user_id=?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.CartItem{}
	for rows.Next() {
		var pid string
		var q int
		if err := rows.Scan(&pid, &q); err != nil {
			return nil, err
		}
		items = append(items, models.CartItem{ProductID: pid, Quantity: q})
	}
	return &models.Cart{UserID: userID, Items: items}, nil
}

// Orders

func (s *MySQLStore) CreateOrder(userID string, items []models.CartItem, amount int64, status, paymentRef string) (*models.Order, error) {
	id := uuid.NewString()
	now := time.Now()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	_, err = tx.Exec(
		`INSERT INTO orders (id, user_id, amount_cents, status, payment_ref, created_at) 
		VALUES (?,?,?,?,?,?)`,
		id, userID, amount, status, paymentRef, now,
	)
	if err != nil {
		return nil, err
	}

	for _, it := range items {
		// Get current price
		var priceCents int64
		row := tx.QueryRow(`SELECT price_cents FROM products WHERE id=?`, it.ProductID)
		if err := row.Scan(&priceCents); err != nil {
			return nil, err
		}

		_, err = tx.Exec(
			`INSERT INTO order_items (order_id, product_id, quantity, price_cents) VALUES (?,?,?,?)`,
			id, it.ProductID, it.Quantity, priceCents,
		)
		if err != nil {
			return nil, err
		}

		// Decrement stock
		res, err := tx.Exec(
			`UPDATE products SET stock = stock - ? WHERE id=? AND stock >= ?`,
			it.Quantity, it.ProductID, it.Quantity,
		)
		if err != nil {
			return nil, err
		}

		affected, _ := res.RowsAffected()
		if affected == 0 {
			return nil, errors.New("insufficient stock for product: " + it.ProductID)
		}
	}

	return &models.Order{
		ID:         id,
		UserID:     userID,
		Items:      items,
		Amount:     amount,
		Status:     status,
		PaymentRef: paymentRef,
		CreatedAt:  now,
	}, nil
}

func (s *MySQLStore) ListOrdersByUser(userID string) ([]*models.Order, error) {
	rows, err := s.db.Query(
		`SELECT id, amount_cents, status, payment_ref, created_at 
		FROM orders WHERE user_id=? ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []*models.Order{}
	for rows.Next() {
		var id, status, pref string
		var amount int64
		var created time.Time
		if err := rows.Scan(&id, &amount, &status, &pref, &created); err != nil {
			return nil, err
		}

		// Fetch items
		ir, err := s.db.Query(
			`SELECT product_id, quantity FROM order_items WHERE order_id=?`,
			id,
		)
		if err != nil {
			return nil, err
		}

		items := []models.CartItem{}
		for ir.Next() {
			var pid string
			var q int
			if err := ir.Scan(&pid, &q); err != nil {
				_ = ir.Close()
				return nil, err
			}
			items = append(items, models.CartItem{ProductID: pid, Quantity: q})
		}
		_ = ir.Close()

		res = append(res, &models.Order{
			ID:         id,
			UserID:     userID,
			Items:      items,
			Amount:     amount,
			Status:     status,
			PaymentRef: pref,
			CreatedAt:  created,
		})
	}
	return res, nil
}

func (s *MySQLStore) ListOrders() ([]*models.Order, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, amount_cents, status, payment_ref, created_at 
		FROM orders ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []*models.Order{}
	for rows.Next() {
		var id, userID, status, pref string
		var amount int64
		var created time.Time
		if err := rows.Scan(&id, &userID, &amount, &status, &pref, &created); err != nil {
			return nil, err
		}

		ir, err := s.db.Query(
			`SELECT product_id, quantity FROM order_items WHERE order_id=?`,
			id,
		)
		if err != nil {
			return nil, err
		}

		items := []models.CartItem{}
		for ir.Next() {
			var pid string
			var q int
			if err := ir.Scan(&pid, &q); err != nil {
				_ = ir.Close()
				return nil, err
			}
			items = append(items, models.CartItem{ProductID: pid, Quantity: q})
		}
		_ = ir.Close()

		res = append(res, &models.Order{
			ID:         id,
			UserID:     userID,
			Items:      items,
			Amount:     amount,
			Status:     status,
			PaymentRef: pref,
			CreatedAt:  created,
		})
	}
	return res, nil
}

func (s *MySQLStore) UpdateOrderStatus(orderID, status string) error {
	res, err := s.db.Exec(`UPDATE orders SET status=? WHERE id=?`, status, orderID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (s *MySQLStore) UpdateOrderPaymentRef(orderID, paymentRef string) error {
	res, err := s.db.Exec(`UPDATE orders SET payment_ref=? WHERE id=?`, paymentRef, orderID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("order not found")
	}
	return nil
}

// Reviews

func (s *MySQLStore) CreateReview(userID, userName, userPhoto string, rating int, comment string) (*models.Review, error) {
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	id := uuid.NewString()
	now := time.Now()

	_, err := s.db.Exec(
		`INSERT INTO reviews (id, user_id, user_name, user_photo, rating, comment, created_at) VALUES (?,?,?,?,?,?,?)`,
		id, userID, userName, userPhoto, rating, comment, now,
	)
	if err != nil {
		return nil, err
	}

	return &models.Review{
		ID:        id,
		UserID:    userID,
		UserName:  userName,
		UserPhoto: userPhoto,
		Rating:    rating,
		Comment:   comment,
		CreatedAt: now,
	}, nil
}

func (s *MySQLStore) ListReviews(limit int) ([]*models.Review, error) {
	query := `SELECT id, user_id, user_name, user_photo, rating, comment, created_at FROM reviews ORDER BY created_at DESC`
	args := []any{}
	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := []*models.Review{}
	for rows.Next() {
		r := models.Review{}
		if err := rows.Scan(&r.ID, &r.UserID, &r.UserName, &r.UserPhoto, &r.Rating, &r.Comment, &r.CreatedAt); err != nil {
			return nil, err
		}
		review := r
		reviews = append(reviews, &review)
	}

	return reviews, nil
}

func (s *MySQLStore) GetUserReviewCount(userID string) (int, error) {
	row := s.db.QueryRow(`SELECT COUNT(*) FROM reviews WHERE user_id=?`, userID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// helpers

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "error 1062") ||
		strings.Contains(errStr, "duplicate entry") ||
		strings.Contains(errStr, "duplicate key")
}
