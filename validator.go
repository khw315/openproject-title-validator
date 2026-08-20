package main

import (
	"fmt"
	"regexp"
	"strings"
)

// titlePattern matches the required format: [TC-NNN][FEATURE][SUB FEATURE][PIC TESTER]
// TC number must be at least 3 digits (e.g., TC-001, TC-1234).
// Each bracket section must contain at least one non-empty character.
var titlePattern = regexp.MustCompile(`^\[TC-\d{3,}\]\[[^\[\]]+\]\[[^\[\]]+\]\[[^\[\]]+\]$`)

// tcNumberPattern matches just the TC-NNN portion for specific validation.
var tcNumberPattern = regexp.MustCompile(`^\[TC-\d{3,}\]`)

// bracketPattern extracts all bracket sections from a title.
var bracketPattern = regexp.MustCompile(`\[[^\[\]]*\]`)

// ValidateTitle checks whether the given subject matches the required format:
// [TC-NNN][NAMA FEATURE][SUB FEATURE][PIC TESTER]
//
// Returns true if valid, or false with a list of violation descriptions.
func ValidateTitle(subject string) (bool, []string) {
	subject = strings.TrimSpace(subject)

	if subject == "" {
		return false, []string{"Judul tiket kosong"}
	}

	// Check if it matches the full pattern.
	if titlePattern.MatchString(subject) {
		return true, nil
	}

	// Provide specific violation messages.
	var violations []string

	// Extract all bracket sections.
	brackets := bracketPattern.FindAllString(subject, -1)

	if len(brackets) == 0 {
		violations = append(violations, "Judul tidak mengikuti format [TC-NNN][NAMA FEATURE][SUB FEATURE][PIC TESTER]")
		return false, violations
	}

	if len(brackets) < 4 {
		violations = append(violations, fmt.Sprintf("Judul harus memiliki 4 bagian dalam bracket, ditemukan %d", len(brackets)))
	}

	if len(brackets) > 4 {
		violations = append(violations, fmt.Sprintf("Judul harus memiliki tepat 4 bagian dalam bracket, ditemukan %d", len(brackets)))
	}

	// Check if title starts with the first bracket.
	if !strings.HasPrefix(subject, "[") {
		violations = append(violations, "Judul harus dimulai dengan bracket [TC-NNN]")
	}

	// Check TC number format in the first bracket.
	if len(brackets) >= 1 {
		first := brackets[0]
		tcContent := first[1 : len(first)-1] // Remove surrounding brackets.
		if !regexp.MustCompile(`^TC-\d{3,}$`).MatchString(tcContent) {
			violations = append(violations, fmt.Sprintf("NOMOR TEST CASE harus berformat TC-NNN (contoh: TC-001), ditemukan: %s", tcContent))
		}
	}

	// Check for empty brackets.
	fieldNames := []string{"NOMOR TEST CASE", "NAMA FEATURE", "SUB FEATURE", "PIC TESTER"}
	for i, bracket := range brackets {
		if i >= 4 {
			break
		}
		content := bracket[1 : len(bracket)-1] // Remove surrounding brackets.
		if strings.TrimSpace(content) == "" {
			violations = append(violations, fmt.Sprintf("%s tidak boleh kosong", fieldNames[i]))
		}
	}

	// Check for extra text after brackets.
	if len(brackets) >= 4 {
		reconstructed := strings.Join(brackets[:4], "")
		if subject != reconstructed {
			violations = append(violations, "Judul tidak boleh memiliki teks tambahan di luar bracket")
		}
	}

	if len(violations) == 0 {
		violations = append(violations, "Judul tidak mengikuti format [TC-NNN][NAMA FEATURE][SUB FEATURE][PIC TESTER]")
	}

	return false, violations
}

// BuildCommentMessage creates the comment message to post on an invalid work package.
func BuildCommentMessage(authorName string, violations []string) string {
	criteria := "[TC-NNN][NAMA FEATURE][SUB FEATURE][PIC TESTER]"

	var b strings.Builder
	b.WriteString(fmt.Sprintf("@%s, judul tiket ini tidak sesuai dengan kriteria **%s**! Harap perbarui judul tiket ini!\n\n", authorName, criteria))

	if len(violations) > 0 {
		b.WriteString("**Detail pelanggaran:**\n")
		for _, v := range violations {
			b.WriteString(fmt.Sprintf("- %s\n", v))
		}
	}

	return b.String()
}
