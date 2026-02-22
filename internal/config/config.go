package config

import (
	"os"
	"strconv"
)

// DANAConfig menyimpan konfigurasi untuk integrasi DANA
type DANAConfig struct {
	// DANA API Credentials
	PartnerID    string
	ClientID     string
	ClientSecret string
	PrivateKey   string
	PublicKey    string
	MerchantID   string

	// Environment: "sandbox" atau "production"
	Environment string

	// Base URLs
	SandboxBaseURL    string
	ProductionBaseURL string

	// Server config
	ServerPort int
	ServerHost string
}

// DefaultConfig mengembalikan konfigurasi default dari environment variables
func DefaultConfig() *DANAConfig {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	} else if p := os.Getenv("SERVER_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = os.Getenv("DANA_ENV")
	}
	if env == "" {
		env = "sandbox"
	}

	return &DANAConfig{
		PartnerID:    getEnv("DANA_PARTNER_ID", os.Getenv("X_PARTNER_ID")),
		ClientID:     getEnv("DANA_CLIENT_ID", os.Getenv("X_PARTNER_ID")),
		ClientSecret: getEnv("DANA_CLIENT_SECRET", os.Getenv("CLIENT_SECRET")),
		PrivateKey:   getEnv("DANA_PRIVATE_KEY", os.Getenv("PRIVATE_KEY")),
		PublicKey:    getEnv("DANA_PUBLIC_KEY", os.Getenv("PUBLIC_KEY")),
		MerchantID:   getEnv("DANA_MERCHANT_ID", ""),
		Environment:  env,

		SandboxBaseURL:    "https://api.sandbox.dana.id",
		ProductionBaseURL: "https://api.dana.id",

		ServerPort: port,
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
	}
}

// BaseURL mengembalikan base URL sesuai environment
func (c *DANAConfig) BaseURL() string {
	if c.Environment == "PRODUCTION" || c.Environment == "production" {
		return c.ProductionBaseURL
	}
	return c.SandboxBaseURL
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
