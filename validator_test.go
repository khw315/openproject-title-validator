package main

import (
	"strings"
	"testing"
)

func TestValidateTitle_ValidTitles(t *testing.T) {
	v, err := NewValidator("", "", "")
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	tests := []struct {
		name    string
		subject string
	}{
		{
			name:    "standard format",
			subject: "[TC-001][Login][Form Validation][Ahmad]",
		},
		{
			name:    "longer TC number",
			subject: "[TC-1234][Dashboard][Chart Rendering][Budi]",
		},
		{
			name:    "TC number with many digits",
			subject: "[TC-99999][Payment][Checkout Flow][Siti]",
		},
		{
			name:    "feature with spaces",
			subject: "[TC-010][User Management][Add New User][Andi Wijaya]",
		},
		{
			name:    "sub feature with special chars",
			subject: "[TC-100][API][GET /users endpoint][Dewi]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, violations := v.Validate(tt.subject)
			if !valid {
				t.Errorf("expected valid title %q, got violations: %v", tt.subject, violations)
			}
			if len(violations) != 0 {
				t.Errorf("expected no violations for %q, got: %v", tt.subject, violations)
			}
		})
	}
}

func TestValidateTitle_InvalidTitles(t *testing.T) {
	v, err := NewValidator("", "", "")
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	tests := []struct {
		name           string
		subject        string
		wantViolations bool
	}{
		{
			name:           "empty title",
			subject:        "",
			wantViolations: true,
		},
		{
			name:           "plain text without brackets",
			subject:        "Fix login bug",
			wantViolations: true,
		},
		{
			name:           "missing sub feature and PIC",
			subject:        "[TC-001][Login]",
			wantViolations: true,
		},
		{
			name:           "missing PIC only",
			subject:        "[TC-001][Login][Form Validation]",
			wantViolations: true,
		},
		{
			name:           "wrong TC format - no dash",
			subject:        "[TC001][Login][Form Validation][Ahmad]",
			wantViolations: true,
		},
		{
			name:           "wrong TC format - too few digits",
			subject:        "[TC-01][Login][Form Validation][Ahmad]",
			wantViolations: true,
		},
		{
			name:           "wrong TC format - no TC prefix",
			subject:        "[001][Login][Form Validation][Ahmad]",
			wantViolations: true,
		},
		{
			name:           "empty feature bracket",
			subject:        "[TC-001][][Form Validation][Ahmad]",
			wantViolations: true,
		},
		{
			name:           "empty sub feature bracket",
			subject:        "[TC-001][Login][][Ahmad]",
			wantViolations: true,
		},
		{
			name:           "empty PIC bracket",
			subject:        "[TC-001][Login][Form Validation][]",
			wantViolations: true,
		},
		{
			name:           "extra text after brackets",
			subject:        "[TC-001][Login][Form Validation][Ahmad] - extra text",
			wantViolations: true,
		},
		{
			name:           "extra text before brackets",
			subject:        "prefix [TC-001][Login][Form Validation][Ahmad]",
			wantViolations: true,
		},
		{
			name:           "spaces only title",
			subject:        "   ",
			wantViolations: true,
		},
		{
			name:           "five brackets instead of four",
			subject:        "[TC-001][Login][Form Validation][Ahmad][Extra]",
			wantViolations: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, violations := v.Validate(tt.subject)
			if valid {
				t.Errorf("expected invalid title %q, but got valid", tt.subject)
			}
			if tt.wantViolations && len(violations) == 0 {
				t.Errorf("expected violations for %q, got none", tt.subject)
			}
		})
	}
}

func TestValidateTitle_CustomPattern(t *testing.T) {
	// Custom pattern: [BUG] or [FEATURE] followed by description
	customPattern := `^\[(BUG|FEATURE)\] .+$`
	customDesc := `[BUG/FEATURE] Description`

	v, err := NewValidator(customPattern, customDesc, "")
	if err != nil {
		t.Fatalf("failed to create custom validator: %v", err)
	}

	// Valid custom titles
	validTitles := []string{
		"[BUG] User cannot login with special characters",
		"[FEATURE] Export reports to PDF",
	}
	for _, title := range validTitles {
		valid, violations := v.Validate(title)
		if !valid {
			t.Errorf("expected valid for %q, got violations: %v", title, violations)
		}
	}

	// Invalid custom titles
	invalidTitles := []string{
		"[TASK] Fix CSS styling",
		"Fix login bug",
		"",
	}
	for _, title := range invalidTitles {
		valid, _ := v.Validate(title)
		if valid {
			t.Errorf("expected invalid for %q, got valid", title)
		}
	}
}

func TestBuildCommentMessage(t *testing.T) {
	v, err := NewValidator("", "", "")
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	violations := []string{
		"NOMOR TEST CASE harus berformat TC-NNN (contoh: TC-001), ditemukan: TC01",
	}

	message := v.BuildCommentMessage("Ahmad", violations)

	if message == "" {
		t.Fatal("expected non-empty comment message")
	}

	// Check that author name is mentioned.
	if !strings.Contains(message, "@Ahmad") {
		t.Error("expected message to contain @Ahmad")
	}

	// Check that criteria format is mentioned.
	if !strings.Contains(message, "[TC-NNN][NAMA FEATURE][SUB FEATURE][PIC TESTER]") {
		t.Error("expected message to contain criteria format")
	}

	// Check that violation details are included.
	if !strings.Contains(message, "TC-NNN") {
		t.Error("expected message to contain violation detail about TC-NNN format")
	}
}

func TestBuildCommentMessage_MultipleViolations(t *testing.T) {
	v, err := NewValidator("", "", "")
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	violations := []string{
		"Judul harus memiliki 4 bagian dalam bracket, ditemukan 2",
		"NOMOR TEST CASE harus berformat TC-NNN (contoh: TC-001), ditemukan: BUG",
	}

	message := v.BuildCommentMessage("Siti", violations)

	if !strings.Contains(message, "@Siti") {
		t.Error("expected message to contain @Siti")
	}

	for _, violation := range violations {
		if !strings.Contains(message, violation) {
			t.Errorf("expected message to contain violation: %s", violation)
		}
	}
}

func TestBuildCommentMessage_CustomTemplate(t *testing.T) {
	template := "Hello {author}! Format must be: {criteria}\n\nList:\n{violations}"
	v, err := NewValidator("", "[FEATURE] desc", template)
	if err != nil {
		t.Fatalf("failed to create validator: %v", err)
	}

	violations := []string{
		"Missing feature prefix",
	}

	message := v.BuildCommentMessage("John", violations)

	if !strings.Contains(message, "Hello John!") {
		t.Errorf("expected message to contain 'Hello John!', got: %s", message)
	}
	if !strings.Contains(message, "Format must be: [FEATURE] desc") {
		t.Errorf("expected message to contain criteria description, got: %s", message)
	}
	if !strings.Contains(message, "- Missing feature prefix") {
		t.Errorf("expected message to contain violation detail, got: %s", message)
	}
}
