package store

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/example/ecommerce-api/internal/models"
)

// InMemoryStore implements users, products, carts, and orders in memory
type InMemoryStore struct {
	mu                 sync.RWMutex
	users              map[string]*models.User
	byEmail            map[string]*models.User
	emailVerifications map[string]*models.EmailVerification
	passwordResets     map[string]*models.PasswordReset
	products           map[string]*models.Product
	carts              map[string]*models.Cart
	orders             map[string]*models.Order
	reviews            map[string]*models.Review
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		users:              make(map[string]*models.User),
		byEmail:            make(map[string]*models.User),
		emailVerifications: make(map[string]*models.EmailVerification),
		passwordResets:     make(map[string]*models.PasswordReset),
		products:           make(map[string]*models.Product),
		carts:              make(map[string]*models.Cart),
		orders:             make(map[string]*models.Order),
		reviews:            make(map[string]*models.Review),
	}
}

// Users

func (s *InMemoryStore) CreateUser(fullName, phone, email, passwordHash, role, provider, googleID string, emailVerified bool) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byEmail[email]; exists {
		return nil, errors.New("email already registered")
	}

	if role == "" {
		role = "user"
	}

	if provider == "" {
		provider = "email"
	}

	u := &models.User{
		ID:            uuid.NewString(),
		FullName:      fullName,
		Phone:         phone,
		Email:         email,
		Password:      passwordHash,
		Role:          role,
		AuthProvider:  provider,
		GoogleID:      googleID,
		EmailVerified: emailVerified,
		CreatedAt:     time.Now(),
	}

	s.users[u.ID] = u
	s.byEmail[email] = u
	return u, nil
}

func (s *InMemoryStore) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.byEmail[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (s *InMemoryStore) GetUserByID(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (s *InMemoryStore) SeedAdminUser(email, passwordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byEmail[email]; exists {
		return nil // already exists
	}

	u := &models.User{
		ID:            uuid.NewString(),
		FullName:      "Administrator",
		Email:         email,
		Password:      passwordHash,
		Role:          "admin",
		AuthProvider:  "email",
		EmailVerified: true,
		CreatedAt:     time.Now(),
	}

	s.users[u.ID] = u
	s.byEmail[email] = u
	return nil
}

func (s *InMemoryStore) UpdateUserPassword(userID, newPasswordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[userID]
	if !ok {
		return errors.New("user not found")
	}

	u.Password = newPasswordHash
	return nil
}

func (s *InMemoryStore) MarkEmailVerified(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[userID]
	if !ok {
		return errors.New("user not found")
	}

	u.EmailVerified = true
	return nil
}

// Email Verification

func (s *InMemoryStore) CreateEmailVerification(userID, token string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.emailVerifications[token] = &models.EmailVerification{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
	return nil
}

func (s *InMemoryStore) GetEmailVerification(token string) (*models.EmailVerification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.emailVerifications[token]
	if !ok {
		return nil, errors.New("verification token not found")
	}
	return v, nil
}

func (s *InMemoryStore) GetLatestEmailVerificationByUser(userID string) (*models.EmailVerification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latest *models.EmailVerification
	for _, verification := range s.emailVerifications {
		if verification.UserID != userID {
			continue
		}
		if latest == nil || verification.ExpiresAt.After(latest.ExpiresAt) {
			copyVerification := *verification
			latest = &copyVerification
		}
	}

	if latest == nil {
		return nil, errors.New("verification token not found")
	}

	return latest, nil
}

func (s *InMemoryStore) DeleteEmailVerification(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.emailVerifications, token)
	return nil
}

// Password Reset

func (s *InMemoryStore) CreatePasswordReset(userID, token string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.passwordResets[token] = &models.PasswordReset{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
	return nil
}

func (s *InMemoryStore) GetPasswordReset(token string) (*models.PasswordReset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.passwordResets[token]
	if !ok {
		return nil, errors.New("reset token not found")
	}
	return r, nil
}

func (s *InMemoryStore) DeletePasswordReset(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.passwordResets, token)
	return nil
}

// Products

func (s *InMemoryStore) CreateProduct(p *models.Product) (*models.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p.ID = uuid.NewString()
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	s.products[p.ID] = p
	return p, nil
}

func (s *InMemoryStore) UpdateProduct(id string, update func(p *models.Product) error) (*models.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.products[id]
	if !ok {
		return nil, errors.New("product not found")
	}

	if err := update(p); err != nil {
		return nil, err
	}

	p.UpdatedAt = time.Now()
	return p, nil
}

func (s *InMemoryStore) DeleteProduct(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.products[id]; !ok {
		return errors.New("product not found")
	}

	delete(s.products, id)
	return nil
}

func (s *InMemoryStore) GetProduct(id string) (*models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.products[id]
	if !ok {
		return nil, errors.New("product not found")
	}
	return p, nil
}

func (s *InMemoryStore) ListProducts(query string) ([]*models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := []*models.Product{}
	for _, p := range s.products {
		if query == "" || containsFold(p.Name, query) || containsFold(p.Description, query) || containsFold(p.SKU, query) || containsFold(p.Category, query) {
			cp := *p
			res = append(res, &cp)
		}
	}
	return res, nil
}

// Carts

func (s *InMemoryStore) GetOrCreateCart(userID string) *models.Cart {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.carts[userID]
	if !ok {
		c = &models.Cart{UserID: userID, Items: []models.CartItem{}}
		s.carts[userID] = c
	}
	return c
}

func (s *InMemoryStore) AddToCart(userID, productID string, qty int) error {
	if qty <= 0 {
		return errors.New("quantity must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.carts[userID]
	if !ok {
		c = &models.Cart{UserID: userID, Items: []models.CartItem{}}
		s.carts[userID] = c
	}

	// Validate product and stock
	p, ok := s.products[productID]
	if !ok {
		return errors.New("product not found")
	}

	// Check current cart quantity
	currentQty := 0
	for i := range c.Items {
		if c.Items[i].ProductID == productID {
			currentQty = c.Items[i].Quantity
			break
		}
	}

	if p.Stock < (currentQty + qty) {
		return errors.New("insufficient stock")
	}

	// Add or update
	found := false
	for i := range c.Items {
		if c.Items[i].ProductID == productID {
			c.Items[i].Quantity += qty
			found = true
			break
		}
	}

	if !found {
		c.Items = append(c.Items, models.CartItem{ProductID: productID, Quantity: qty})
	}

	return nil
}

func (s *InMemoryStore) RemoveFromCart(userID, productID string, qty int) error {
	if qty <= 0 {
		return errors.New("quantity must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.carts[userID]
	if !ok {
		return errors.New("cart not found")
	}

	for i := range c.Items {
		if c.Items[i].ProductID == productID {
			if c.Items[i].Quantity <= qty {
				// Remove item
				c.Items = append(c.Items[:i], c.Items[i+1:]...)
				return nil
			}
			c.Items[i].Quantity -= qty
			return nil
		}
	}

	return errors.New("item not in cart")
}

func (s *InMemoryStore) ClearCart(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.carts, userID)
}

func (s *InMemoryStore) GetCart(userID string) (*models.Cart, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.carts[userID]
	if !ok {
		return &models.Cart{UserID: userID, Items: []models.CartItem{}}, nil
	}

	copyItems := make([]models.CartItem, len(c.Items))
	copy(copyItems, c.Items)
	return &models.Cart{UserID: userID, Items: copyItems}, nil
}

// Orders

func (s *InMemoryStore) CreateOrder(userID string, items []models.CartItem, amount int64, status, paymentRef string) (*models.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.NewString()
	o := &models.Order{
		ID:         id,
		UserID:     userID,
		Items:      append([]models.CartItem(nil), items...),
		Amount:     amount,
		Status:     status,
		PaymentRef: paymentRef,
		CreatedAt:  time.Now(),
	}

	s.orders[id] = o

	// Decrement stock
	for _, it := range items {
		if p, ok := s.products[it.ProductID]; ok {
			if p.Stock >= it.Quantity {
				p.Stock -= it.Quantity
			}
		}
	}

	return o, nil
}

func (s *InMemoryStore) ListOrdersByUser(userID string) ([]*models.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := []*models.Order{}
	for _, o := range s.orders {
		if o.UserID == userID {
			copyO := *o
			res = append(res, &copyO)
		}
	}
	return res, nil
}

func (s *InMemoryStore) ListOrders() ([]*models.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*models.Order, 0, len(s.orders))
	for _, o := range s.orders {
		copyO := *o
		res = append(res, &copyO)
	}
	return res, nil
}

func (s *InMemoryStore) UpdateOrderStatus(orderID, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	o, ok := s.orders[orderID]
	if !ok {
		return errors.New("order not found")
	}

	o.Status = status
	return nil
}

func (s *InMemoryStore) UpdateOrderPaymentRef(orderID, paymentRef string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	o, ok := s.orders[orderID]
	if !ok {
		return errors.New("order not found")
	}

	o.PaymentRef = paymentRef
	return nil
}

// Reviews

func (s *InMemoryStore) CreateReview(userID, userName, userPhoto string, rating int, comment string) (*models.Review, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	id := uuid.NewString()
	r := &models.Review{
		ID:        id,
		UserID:    userID,
		UserName:  userName,
		UserPhoto: userPhoto,
		Rating:    rating,
		Comment:   comment,
		CreatedAt: time.Now(),
	}

	s.reviews[id] = r
	return r, nil
}

func (s *InMemoryStore) ListReviews(limit int) ([]*models.Review, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := []*models.Review{}
	for _, r := range s.reviews {
		cp := *r
		res = append(res, &cp)
	}

	// Sort by created date descending
	for i := 0; i < len(res)-1; i++ {
		for j := i + 1; j < len(res); j++ {
			if res[i].CreatedAt.Before(res[j].CreatedAt) {
				res[i], res[j] = res[j], res[i]
			}
		}
	}

	if limit > 0 && len(res) > limit {
		res = res[:limit]
	}

	return res, nil
}

func (s *InMemoryStore) GetUserReviewCount(userID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, r := range s.reviews {
		if r.UserID == userID {
			count++
		}
	}
	return count, nil
}

// Utils

func containsFold(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	bs := []byte(s)
	for i, b := range bs {
		if 'A' <= b && b <= 'Z' {
			bs[i] = b + 32
		}
	}
	return string(bs)
}

func contains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
