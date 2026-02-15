package routes

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/example/ecommerce-api/internal/auth"
	"github.com/example/ecommerce-api/internal/config"
	"github.com/example/ecommerce-api/internal/email"
	"github.com/example/ecommerce-api/internal/handlers"
	"github.com/example/ecommerce-api/internal/middleware"
	"github.com/example/ecommerce-api/internal/payment"
	"github.com/example/ecommerce-api/internal/store"
)

func Register(r *gin.Engine, cfg *config.Config) error {
	var st store.Store

	// Initialize store based on backend type
	switch cfg.StoreBackend {
	case "mysql":
		log.Printf("Connecting to MySQL...")
		ms, err := store.NewMySQLStore(cfg.MySQLDSN)
		if err != nil {
			return fmt.Errorf("mysql connection failed: %w", err)
		} else {
			log.Println("✅ MySQL connected successfully")
			st = ms
		}
	case "postgres":
		log.Printf("⚠️  PostgreSQL not yet implemented, using in-memory store")
		st = store.NewInMemoryStore()
	default:
		log.Println("📝 Using in-memory store")
		st = store.NewInMemoryStore()
	}

	// Seed default admin user
	adminHash, _ := auth.HashPassword(cfg.AdminPassword)
	if err := st.SeedAdminUser(cfg.AdminEmail, adminHash); err != nil {
		log.Printf("⚠️  Failed to seed admin: %v", err)
	} else {
		log.Printf("✅ Admin user seeded: %s", cfg.AdminEmail)
	}

	// Initialize services
	jwtm := auth.NewJWTManager(cfg.JWTSecret)

	// Initialize payment gateway
	var pay payment.Gateway
	if cfg.MidtransServerKey != "" {
		pay = payment.NewMidtransGateway(cfg.MidtransServerKey, cfg.MidtransIsProduction)
		log.Println("✅ Midtrans payment gateway initialized")
	} else {
		pay = &payment.MockGateway{}
		log.Println("⚠️  Using Mock payment gateway (set MIDTRANS_SERVER_KEY for real payments)")
	}

	// Initialize email service
	var emailSvc *email.Service
	if cfg.SMTPFrom != "" && cfg.SMTPPassword != "" {
		emailSvc = email.NewService(cfg.SMTPFrom, cfg.SMTPPassword, cfg.SMTPHost, cfg.SMTPPort)
		log.Println("✅ Email service initialized")
	} else {
		log.Println("⚠️  Email service disabled (set SMTP_FROM and SMTP_PASSWORD to enable)")
	}

	// Initialize handlers
	authH := handlers.NewAuthHandler(cfg, st, jwtm, emailSvc)
	prodH := handlers.NewProductsHandler(st)
	cartH := handlers.NewCartHandler(st)
	checkH := handlers.NewCheckoutHandler(cfg, st, pay, emailSvc)
	reviewH := handlers.NewReviewsHandler(st)
	adminOrdersH := handlers.NewAdminOrdersHandler(st)
	uploadsH := handlers.NewUploadsHandler(cfg)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"backend": cfg.StoreBackend,
			"payment": func() string {
				if cfg.MidtransServerKey != "" {
					return "midtrans"
				}
				return "mock"
			}(),
			"email": emailSvc != nil,
		})
	})

	r.Static("/uploads", "./uploads")

	api := r.Group("/api/v1")

	// Auth routes
	authG := api.Group("/auth")
	{
		authG.POST("/signup", authH.Signup)
		authG.POST("/login", authH.Login)
		authG.POST("/login-google", authH.GoogleLogin)
		authG.GET("/google/start", authH.GoogleStart)
		authG.GET("/google/callback", authH.GoogleCallback)
		authG.POST("/verify-otp", authH.VerifyOTP)
		authG.POST("/resend-otp", authH.ResendOTP)
		authG.GET("/verify-email", authH.VerifyEmail)
		authG.POST("/request-password-reset", authH.RequestPasswordReset)
		authG.POST("/reset-password", authH.ResetPassword)
	}

	// Public product routes
	prod := api.Group("/products")
	{
		prod.GET("", prodH.List)
		prod.GET("/:id", prodH.Get)
	}

	// Public review routes
	api.GET("/reviews", reviewH.List)

	// Admin routes
	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth(jwtm), middleware.RequireAdmin())
	{
		admin.GET("/products", prodH.AdminList)
		admin.POST("/products", prodH.Create)
		admin.PUT("/products/:id", prodH.Update)
		admin.DELETE("/products/:id", prodH.Delete)
		admin.GET("/orders", adminOrdersH.List)
		admin.PUT("/orders/:id/status", adminOrdersH.UpdateStatus)
		admin.POST("/uploads/thumbnail", uploadsH.UploadProductThumbnail)
	}

	// User routes (authenticated)
	user := api.Group("/me")
	user.Use(middleware.JWTAuth(jwtm))
	{
		user.GET("/cart", cartH.View)
		user.POST("/cart/add", cartH.Add)
		user.POST("/cart/remove", cartH.Remove)
		user.POST("/checkout", checkH.Checkout)
		user.GET("/orders", checkH.MyOrders)
		user.POST("/reviews", reviewH.Create)
	}

	// Midtrans webhook
	api.POST("/webhooks/midtrans", checkH.MidtransCallback)

	log.Println("✅ Routes registered successfully")
	return nil
}
