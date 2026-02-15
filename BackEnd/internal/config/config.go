package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	JWTSecret     string
	AdminEmail    string
	AdminPassword string
	BaseURL       string // e.g. http://localhost:8080
	FrontendURL   string // e.g. http://localhost:4321

	StoreBackend string // "memory" or "mysql" or "postgres"
	MySQLDSN     string
	PostgresDSN  string

	// Midtrans
	MidtransServerKey    string
	MidtransIsProduction bool

	// Email (Gmail SMTP)
	SMTPFrom     string
	SMTPPassword string // App Password dari Gmail
	SMTPHost     string
	SMTPPort     string

	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func requireEnv(key string) (string, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return v, nil
}

func Load() (*Config, error) {
	// Load .env if present; ignore error to allow pure env-based config
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getenv("PORT", "8080"),
		BaseURL:     getenv("BASE_URL", "http://localhost:8080"),
		FrontendURL: getenv("FRONTEND_URL", "http://localhost:4321"),

		StoreBackend: getenv("STORE_BACKEND", "mysql"),
		MySQLDSN:     strings.TrimSpace(os.Getenv("MYSQL_DSN")),
		PostgresDSN:  strings.TrimSpace(os.Getenv("POSTGRES_DSN")),

		// Midtrans
		MidtransServerKey:    strings.TrimSpace(os.Getenv("MIDTRANS_SERVER_KEY")),
		MidtransIsProduction: getenv("MIDTRANS_IS_PRODUCTION", "false") == "true",

		// Email
		SMTPFrom:     strings.TrimSpace(os.Getenv("SMTP_FROM")),
		SMTPPassword: strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		SMTPHost:     getenv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getenv("SMTP_PORT", "587"),

		// Google OAuth
		GoogleClientID:     strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),
		GoogleClientSecret: strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")),
		GoogleRedirectURL:  strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL")),
	}

	var err error
	cfg.JWTSecret, err = requireEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}
	cfg.AdminEmail, err = requireEnv("ADMIN_EMAIL")
	if err != nil {
		return nil, err
	}
	cfg.AdminPassword, err = requireEnv("ADMIN_PASSWORD")
	if err != nil {
		return nil, err
	}

	// Validate store backend
	validBackends := map[string]bool{
		"memory":   true,
		"mysql":    true,
		"postgres": true,
	}

	if !validBackends[cfg.StoreBackend] {
		return nil, fmt.Errorf("unknown STORE_BACKEND %q", cfg.StoreBackend)
	}

	switch cfg.StoreBackend {
	case "mysql":
		if cfg.MySQLDSN == "" {
			return nil, fmt.Errorf("MYSQL_DSN is required when STORE_BACKEND=mysql")
		}
	case "postgres":
		if cfg.PostgresDSN == "" {
			return nil, fmt.Errorf("POSTGRES_DSN is required when STORE_BACKEND=postgres")
		}
	}

	if (cfg.SMTPFrom == "") != (cfg.SMTPPassword == "") {
		return nil, fmt.Errorf("SMTP_FROM and SMTP_PASSWORD must be set together")
	}

	if (cfg.GoogleClientID == "") != (cfg.GoogleClientSecret == "") || (cfg.GoogleClientID == "") != (cfg.GoogleRedirectURL == "") {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and GOOGLE_REDIRECT_URL must be set together")
	}

	log.Printf("config loaded: backend=%s, port=%s", cfg.StoreBackend, cfg.Port)
	return cfg, nil
}
