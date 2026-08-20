package main

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// DefaultTitlePattern matches [TC-NNN][FEATURE][SUB FEATURE][PIC TESTER]
	DefaultTitlePattern = `^\[TC-\d{3,}\]\[[^\[\]]+\]\[[^\[\]]+\]\[[^\[\]]+\]$`

	// DefaultTitleCriteriaDesc is the default human-readable format string.
	DefaultTitleCriteriaDesc = `[TC-NNN][NAMA FEATURE][SUB FEATURE][PIC TESTER]`

	// DefaultCommentTemplate is the default template for comments on invalid work packages.
	DefaultCommentTemplate = `@{author}, judul tiket ini tidak sesuai dengan kriteria **{criteria}**! Harap perbarui judul tiket ini!`
)

var (
	// bracketPattern extracts all bracket sections from a title.
	bracketPattern = regexp.MustCompile(`\[[^\[\]]*\]`)

	// tcFormatRegex validates the TC-NNN format within the first bracket.
	tcFormatRegex = regexp.MustCompile(`^TC-\d{3,}$`)

	// fieldNames maps bracket index to field description.
	fieldNames = [4]string{"NOMOR TEST CASE", "NAMA FEATURE", "SUB FEATURE", "PIC TESTER"}
)

// Validator validates work package titles against a regex pattern and criteria description.
type Validator struct {
	pattern         *regexp.Regexp
	criteriaDesc    string
	commentTemplate string
}

// NewValidator creates a new Validator with the given pattern, criteria description, and comment template.
func NewValidator(patternStr, criteriaDesc, commentTemplate string) (*Validator, error) {
	if patternStr == "" {
		patternStr = DefaultTitlePattern
	}
	if criteriaDesc == "" {
		criteriaDesc = DefaultTitleCriteriaDesc
	}
	if commentTemplate == "" {
		commentTemplate = DefaultCommentTemplate
	}

	re, err := regexp.Compile(patternStr)
	if err != nil {
		return nil, fmt.Errorf("compile regex pattern %q: %w", patternStr, err)
	}

	// Support literal \n escaped sequences in template
	commentTemplate = strings.ReplaceAll(commentTemplate, `\n`, "\n")

	return &Validator{
		pattern:         re,
		criteriaDesc:    criteriaDesc,
		commentTemplate: commentTemplate,
	}, nil
}

// Validate checks whether the given subject matches the configured regex pattern.
// Returns true if valid, or false with a list of violation descriptions.
func (v *Validator) Validate(subject string) (bool, []string) {
	subject = strings.TrimSpace(subject)

	if subject == "" {
		return false, []string{"Judul tiket kosong"}
	}

	// Check if it matches the configured pattern.
	if v.pattern.MatchString(subject) {
		return true, nil
	}

	// If using the default 4-bracket format, provide detailed structural error messages.
	if v.pattern.String() == DefaultTitlePattern {
		return false, v.validateDefaultBreakdown(subject)
	}

	// For custom patterns, provide general violation message.
	return false, []string{fmt.Sprintf("Judul tidak memenuhi kriteria format: %s", v.criteriaDesc)}
}

// validateDefaultBreakdown coordinates structural checks for the standard [TC-NNN][...][...][...] format.
func (v *Validator) validateDefaultBreakdown(subject string) []string {
	brackets := bracketPattern.FindAllString(subject, -1)
	if len(brackets) == 0 {
		return []string{"Judul tidak mengikuti format [TC-NNN][NAMA FEATURE][SUB FEATURE][PIC TESTER]"}
	}

	var violations []string
	violations = append(violations, checkBracketCount(brackets, subject)...)
	violations = append(violations, checkTestCaseFormat(brackets[0])...)
	violations = append(violations, checkEmptyFields(brackets)...)
	violations = append(violations, checkExtraText(brackets, subject)...)

	if len(violations) == 0 {
		return []string{"Judul tidak mengikuti format [TC-NNN][NAMA FEATURE][SUB FEATURE][PIC TESTER]"}
	}

	return violations
}

// checkBracketCount validates the number of brackets and prefix.
func checkBracketCount(brackets []string, subject string) []string {
	var violations []string

	if len(brackets) != 4 {
		violations = append(violations, fmt.Sprintf("Judul harus memiliki 4 bagian dalam bracket, ditemukan %d", len(brackets)))
	}

	if !strings.HasPrefix(subject, "[") {
		violations = append(violations, "Judul harus dimulai dengan bracket [TC-NNN]")
	}

	return violations
}

// checkTestCaseFormat validates that the first bracket contains a valid TC-NNN identifier.
func checkTestCaseFormat(firstBracket string) []string {
	tcContent := firstBracket[1 : len(firstBracket)-1]
	if !tcFormatRegex.MatchString(tcContent) {
		return []string{fmt.Sprintf("NOMOR TEST CASE harus berformat TC-NNN (contoh: TC-001), ditemukan: %s", tcContent)}
	}
	return nil
}

// checkEmptyFields checks that none of the 4 required brackets are empty.
func checkEmptyFields(brackets []string) []string {
	var violations []string
	limit := min(len(brackets), 4)

	for i := 0; i < limit; i++ {
		content := brackets[i][1 : len(brackets[i])-1]
		if strings.TrimSpace(content) == "" {
			violations = append(violations, fmt.Sprintf("%s tidak boleh kosong", fieldNames[i]))
		}
	}

	return violations
}

// checkExtraText checks if there are trailing characters outside the 4 brackets.
func checkExtraText(brackets []string, subject string) []string {
	if len(brackets) < 4 {
		return nil
	}

	reconstructed := strings.Join(brackets[:4], "")
	if subject != reconstructed {
		return []string{"Judul tidak boleh memiliki teks tambahan di luar bracket"}
	}

	return nil
}

// BuildCommentMessage creates the comment message to post on an invalid work package.
func (v *Validator) BuildCommentMessage(authorName string, violations []string) string {
	msg := strings.ReplaceAll(v.commentTemplate, "{author}", authorName)
	msg = strings.ReplaceAll(msg, "{criteria}", v.criteriaDesc)

	if strings.Contains(msg, "{violations}") {
		var violationDetails strings.Builder
		for _, violation := range violations {
			violationDetails.WriteString(fmt.Sprintf("- %s\n", violation))
		}
		return strings.ReplaceAll(msg, "{violations}", strings.TrimRight(violationDetails.String(), "\n"))
	}

	var b strings.Builder
	b.WriteString(msg)
	b.WriteString("\n\n")

	if len(violations) > 0 {
		b.WriteString("**Detail pelanggaran:**\n")
		for _, violation := range violations {
			b.WriteString(fmt.Sprintf("- %s\n", violation))
		}
	}

	return b.String()
}
