E-Commerce API (Go + Gin)

Features
- JWT authentication (signup, login) with role-based access control
- Default admin user created on startup
- Admin endpoints for managing products and inventory
- Product search and listing
- Cart management (add/remove/list items)
- Checkout and payment (mock gateway with structure ready for Stripe)
- MySQL persistence with automatic schema creation
- In-memory storage option for development

Tech Stack
- Go 1.21+
- Gin Web Framework
- MySQL (or in-memory storage)

Setup
1. Create MySQL database:
   CREATE DATABASE ecommerce CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

2. Create env file:
   - Copy `.env.example` to `.env`
   - Fill all required values (`JWT_SECRET`, `ADMIN_EMAIL`, `ADMIN_PASSWORD`)
   - Set storage backend via `STORE_BACKEND` and corresponding DSN (`MYSQL_DSN` or `POSTGRES_DSN`)
   - Optional email must be complete pair: `SMTP_FROM` + `SMTP_PASSWORD`

3. Install dependencies and run:
   - go mod tidy
   - go run ./cmd/server

Default Admin
- Email: from `ADMIN_EMAIL`
- Password: from `ADMIN_PASSWORD`
- Role: admin (created automatically on startup)
- Login to get JWT token with admin privileges

Sample API Flow
1. Login as admin:
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H 'Content-Type: application/json' \
     -d '{"email":"admin@example.com","password":"admin123"}'

2. Create a product (admin only):
   curl -X POST http://localhost:8080/api/v1/admin/products \
     -H 'Content-Type: application/json' \
     -H 'Authorization: Bearer <token>' \
     -d '{"name":"Laptop","description":"High-end laptop","price_cents":150000,"sku":"LAPTOP-001","stock":10}'

3. Signup as regular user:
   curl -X POST http://localhost:8080/api/v1/auth/signup \
     -H 'Content-Type: application/json' \
     -d '{"email":"user@example.com","password":"pass123"}'

4. Login as user:
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H 'Content-Type: application/json' \
     -d '{"email":"user@example.com","password":"pass123"}'

5. List products:
   curl http://localhost:8080/api/v1/products

6. Add to cart:
   curl -X POST http://localhost:8080/api/v1/me/cart/add \
     -H 'Content-Type: application/json' \
     -H 'Authorization: Bearer <user-token>' \
     -d '{"product_id":"<product-id>","quantity":2}'

7. View cart:
   curl http://localhost:8080/api/v1/me/cart \
     -H 'Authorization: Bearer <user-token>'

8. Checkout:
   curl -X POST http://localhost:8080/api/v1/me/checkout \
     -H 'Content-Type: application/json' \
     -H 'Authorization: Bearer <user-token>' \
     -d '{"payment_method":"card"}'

9. View orders:
   curl http://localhost:8080/api/v1/me/orders \
     -H 'Authorization: Bearer <user-token>'

Project Structure
- cmd/server/main.go         -> Entry point
- internal/config/config.go  -> Config loader (.env support)
- internal/models/models.go  -> Data models (User with role, Product, Cart, Order)
- internal/store/store.go    -> Store interface + in-memory implementation
- internal/store/mysql.go    -> MySQL implementation with auto-migration
- internal/auth/jwt.go       -> JWT token generation and verification
- internal/auth/password.go  -> Password hashing and verification
- internal/payment/payment.go -> Payment gateway interface + mock
- internal/handlers/         -> HTTP handlers (auth, products, cart, checkout)
- internal/middleware/jwt.go -> JWT auth middleware and admin guard
- internal/routes/routes.go  -> Route wiring and admin seeding

Database Schema
- users: id, email, password_hash, role (user/admin), created_at
- products: id, name, description, price_cents, sku, stock, created_at, updated_at
- carts: user_id, updated_at
- cart_items: user_id, product_id, quantity
- orders: id, user_id, amount_cents, status, payment_ref, created_at
- order_items: order_id, product_id, quantity

Notes
- Admin user is automatically created on startup if it doesn't exist
- Users have role "user" by default; only admin role can manage products
- MySQL tables are auto-created on first connection
- Payment is mocked but ready to integrate Stripe
- Switch between MySQL and in-memory via STORE_BACKEND in .env
