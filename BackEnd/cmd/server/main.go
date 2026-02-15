package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/example/ecommerce-api/internal/config"
	"github.com/example/ecommerce-api/internal/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("❌ Failed to load config: %v", err)
		os.Exit(1)
	}

	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // Frontend URLs
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // 12 hours
	}))

	// Register routes
	if err := routes.Register(r, cfg); err != nil {
		log.Printf("❌ Failed to register routes: %v", err)
		os.Exit(1)
	}

	addr := ":" + cfg.Port
	log.Printf("🚀 Server starting on %s", addr)
	log.Printf("📦 Store backend: %s", cfg.StoreBackend)
	log.Printf("👤 Admin email: %s", cfg.AdminEmail)

	if err := r.Run(addr); err != nil {
		log.Printf("❌ Server failed: %v", err)
		os.Exit(1)
	}
}
