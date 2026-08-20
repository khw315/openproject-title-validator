package main

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all configuration values for the webhook service.
type Config struct {
	// OpenProjectURL is the base URL of the OpenProject instance.
	OpenProjectURL string

	// OpenProjectAPIKey is the API key for authenticating with OpenProject API.
	OpenProjectAPIKey string

	// WebhookSecret is the optional shared secret for signature verification.
	WebhookSecret string

	// Port is the HTTP server listen port.
	Port string
}

// LoadConfig reads configuration from environment variables and validates required fields.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		OpenProjectURL:    os.Getenv("OPENPROJECT_URL"),
		OpenProjectAPIKey: os.Getenv("OPENPROJECT_API_KEY"),
		WebhookSecret:     os.Getenv("WEBHOOK_SECRET"),
		Port:              os.Getenv("PORT"),
	}

	if cfg.OpenProjectURL == "" {
		return nil, fmt.Errorf("OPENPROJECT_URL is required")
	}
	if cfg.OpenProjectAPIKey == "" {
		return nil, fmt.Errorf("OPENPROJECT_API_KEY is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	// Strip trailing slash from OpenProject URL if present.
	cfg.OpenProjectURL = strings.TrimRight(cfg.OpenProjectURL, "/")

	return cfg, nil
}
