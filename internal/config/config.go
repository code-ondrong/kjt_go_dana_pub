package config

import (
	"fmt"
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
	ShopID       string
	DivisionID   string
	ChargeTarget string // "MERCHANT" (default) or "DIVISION"
	ChannelID    string

	// Environment: "sandbox" atau "production"
	Environment string

	// Base URLs
	SandboxBaseURL    string
	ProductionBaseURL string

	// Server config
	ServerPort int
	ServerHost string

	// Origin untuk Merchant Management API (required by DANA)
	Origin string
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

	// X_PARTNER_ID is also known as clientId per DANA SDK docs
	partnerID := getEnv("X_PARTNER_ID", os.Getenv("DANA_PARTNER_ID"))
	clientID := getEnv("DANA_CLIENT_ID", os.Getenv("CLIENT_ID"))
	if clientID == "" {
		clientID = partnerID // X_PARTNER_ID doubles as clientId
	}

	return &DANAConfig{
		PartnerID:    partnerID,
		ClientID:     clientID,
		ClientSecret: getEnv("CLIENT_SECRET", os.Getenv("DANA_CLIENT_SECRET")),
		PrivateKey:   getEnv("PRIVATE_KEY", os.Getenv("DANA_PRIVATE_KEY")),
		PublicKey:    getEnv("DANA_PUBLIC_KEY", os.Getenv("PUBLIC_KEY")),
		MerchantID:   getEnv("DANA_MERCHANT_ID", os.Getenv("MERCHANT_ID")),
		ShopID:       getEnv("SHOP_ID", os.Getenv("DANA_SHOP_ID")),
		DivisionID:   getEnv("DIVISION_ID", os.Getenv("DANA_DIVISION_ID")),
		ChargeTarget: getEnv("CHARGE_TARGET", os.Getenv("DANA_CHARGE_TARGET")), // "MERCHANT" or "DIVISION"
		ChannelID:    getEnv("CHANNEL_ID", "95221"),                            // Default CHANNEL_ID for Payment Gateway
		Environment:  env,

		SandboxBaseURL:    "http://api.sandbox.dana.id", // BUGFIX #1: Sandbox uses HTTP, not HTTPS
		ProductionBaseURL: "https://api.dana.id",

		ServerPort: port,
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		Origin:     getEnv("ORIGIN", "http://localhost:8888"), // Required for Merchant Management
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

// ValidatePaymentGateway validates configuration for Payment Gateway UAT
func (c *DANAConfig) ValidatePaymentGateway() error {
	if c.PartnerID == "" {
		return fmt.Errorf("X_PARTNER_ID (or DANA_PARTNER_ID) is required for Payment Gateway")
	}
	if c.MerchantID == "" {
		return fmt.Errorf("DANA_MERCHANT_ID is required for Payment Gateway")
	}
	if c.PrivateKey == "" {
		return fmt.Errorf("PRIVATE_KEY (or DANA_PRIVATE_KEY) is required for Payment Gateway signature")
	}
	if c.Origin == "" {
		return fmt.Errorf("ORIGIN is required for Payment Gateway")
	}
	if c.ChannelID == "" {
		return fmt.Errorf("CHANNEL_ID is required for Payment Gateway")
	}
	if c.Environment != "sandbox" && c.Environment != "production" {
		return fmt.Errorf("DANA_ENV must be 'sandbox' or 'production', got: %s", c.Environment)
	}
	return nil
}
