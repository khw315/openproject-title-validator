package main

import (
	"os"
	"testing"
)

func TestLoadConfig_RequiredFields(t *testing.T) {
	// Clear environment.
	os.Unsetenv("OPENPROJECT_URL")
	os.Unsetenv("OPENPROJECT_API_KEY")
	os.Unsetenv("WEBHOOK_SECRET")
	os.Unsetenv("PORT")
	os.Unsetenv("TITLE_PATTERN")
	os.Unsetenv("TITLE_CRITERIA_DESC")
	os.Unsetenv("COMMENT_TEMPLATE")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error when OPENPROJECT_URL is missing")
	}

	os.Setenv("OPENPROJECT_URL", "http://localhost")
	_, err = LoadConfig()
	if err == nil {
		t.Fatal("expected error when OPENPROJECT_API_KEY is missing")
	}

	os.Setenv("OPENPROJECT_API_KEY", "test-key")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.TitlePattern != DefaultTitlePattern {
		t.Errorf("expected default title pattern, got %s", cfg.TitlePattern)
	}
	if cfg.TitleCriteriaDesc != DefaultTitleCriteriaDesc {
		t.Errorf("expected default title criteria, got %s", cfg.TitleCriteriaDesc)
	}
	if cfg.CommentTemplate != DefaultCommentTemplate {
		t.Errorf("expected default comment template, got %s", cfg.CommentTemplate)
	}
}

func TestLoadConfig_TrimsTrailingSlash(t *testing.T) {
	os.Setenv("OPENPROJECT_URL", "http://localhost:3000/")
	os.Setenv("OPENPROJECT_API_KEY", "test-key")
	os.Setenv("PORT", "9090")
	defer func() {
		os.Unsetenv("OPENPROJECT_URL")
		os.Unsetenv("OPENPROJECT_API_KEY")
		os.Unsetenv("PORT")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.OpenProjectURL != "http://localhost:3000" {
		t.Errorf("expected URL without trailing slash, got %s", cfg.OpenProjectURL)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
}

func TestLoadConfig_AllFields(t *testing.T) {
	os.Setenv("OPENPROJECT_URL", "https://op.example.com")
	os.Setenv("OPENPROJECT_API_KEY", "my-api-key")
	os.Setenv("WEBHOOK_SECRET", "my-secret")
	os.Setenv("PORT", "3000")
	os.Setenv("TITLE_PATTERN", `^\[PROJ-\d+\] .+$`)
	os.Setenv("TITLE_CRITERIA_DESC", `[PROJ-123] Description`)
	os.Setenv("COMMENT_TEMPLATE", `Hi @{author}, please follow {criteria}!`)
	defer func() {
		os.Unsetenv("OPENPROJECT_URL")
		os.Unsetenv("OPENPROJECT_API_KEY")
		os.Unsetenv("WEBHOOK_SECRET")
		os.Unsetenv("PORT")
		os.Unsetenv("TITLE_PATTERN")
		os.Unsetenv("TITLE_CRITERIA_DESC")
		os.Unsetenv("COMMENT_TEMPLATE")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.OpenProjectURL != "https://op.example.com" {
		t.Errorf("expected https://op.example.com, got %s", cfg.OpenProjectURL)
	}
	if cfg.OpenProjectAPIKey != "my-api-key" {
		t.Errorf("expected my-api-key, got %s", cfg.OpenProjectAPIKey)
	}
	if cfg.WebhookSecret != "my-secret" {
		t.Errorf("expected my-secret, got %s", cfg.WebhookSecret)
	}
	if cfg.Port != "3000" {
		t.Errorf("expected 3000, got %s", cfg.Port)
	}
	if cfg.TitlePattern != `^\[PROJ-\d+\] .+$` {
		t.Errorf("expected custom title pattern, got %s", cfg.TitlePattern)
	}
	if cfg.TitleCriteriaDesc != `[PROJ-123] Description` {
		t.Errorf("expected custom criteria desc, got %s", cfg.TitleCriteriaDesc)
	}
	if cfg.CommentTemplate != `Hi @{author}, please follow {criteria}!` {
		t.Errorf("expected custom comment template, got %s", cfg.CommentTemplate)
	}
}

func TestLoadConfig_InvalidRegexPattern(t *testing.T) {
	os.Setenv("OPENPROJECT_URL", "https://op.example.com")
	os.Setenv("OPENPROJECT_API_KEY", "my-api-key")
	os.Setenv("TITLE_PATTERN", `[invalid(regex`)
	defer func() {
		os.Unsetenv("OPENPROJECT_URL")
		os.Unsetenv("OPENPROJECT_API_KEY")
		os.Unsetenv("TITLE_PATTERN")
	}()

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error on invalid regex pattern, got nil")
	}
}
