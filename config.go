package main

import (
	"fmt"
	"os"
	"regexp"
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

	// TitlePattern is the regex pattern for validating ticket titles.
	TitlePattern string

	// TitleCriteriaDesc is the human-readable criteria format description.
	TitleCriteriaDesc string

	// CommentTemplate is the customizable comment template for invalid titles.
	CommentTemplate string
}

// LoadConfig reads configuration from environment variables and validates required fields.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		OpenProjectURL:    os.Getenv("OPENPROJECT_URL"),
		OpenProjectAPIKey: os.Getenv("OPENPROJECT_API_KEY"),
		WebhookSecret:     os.Getenv("WEBHOOK_SECRET"),
		Port:              os.Getenv("PORT"),
		TitlePattern:      os.Getenv("TITLE_PATTERN"),
		TitleCriteriaDesc: os.Getenv("TITLE_CRITERIA_DESC"),
		CommentTemplate:   os.Getenv("COMMENT_TEMPLATE"),
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

	if cfg.TitlePattern == "" {
		cfg.TitlePattern = DefaultTitlePattern
	}
	if cfg.TitleCriteriaDesc == "" {
		cfg.TitleCriteriaDesc = DefaultTitleCriteriaDesc
	}
	if cfg.CommentTemplate == "" {
		cfg.CommentTemplate = DefaultCommentTemplate
	}

	// Validate that the title pattern is a valid regular expression.
	if _, err := regexp.Compile(cfg.TitlePattern); err != nil {
		return nil, fmt.Errorf("invalid TITLE_PATTERN regex: %w", err)
	}

	// Strip trailing slash from OpenProject URL if present.
	cfg.OpenProjectURL = strings.TrimRight(cfg.OpenProjectURL, "/")

	return cfg, nil
}
