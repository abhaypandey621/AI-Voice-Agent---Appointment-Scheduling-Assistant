package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// EmailValidator provides email validation functionality
type EmailValidator struct {
	// Using simplified RFC 5322 regex for email validation
	// Covers most common use cases
	emailRegex *regexp.Regexp
}

// NewEmailValidator creates a new email validator
func NewEmailValidator() *EmailValidator {
	// Simplified email regex that covers most common patterns
	// Format: localpart@domain.extension
	emailRegex := regexp.MustCompile(
		`^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`,
	)

	return &EmailValidator{
		emailRegex: emailRegex,
	}
}

// ValidateEmail validates and normalizes an email address
// Returns (isValid bool, normalizedEmail string, error)
func (ev *EmailValidator) ValidateEmail(email string) (bool, string, error) {
	// Check for null or empty
	if email == "" {
		return false, "", fmt.Errorf("email cannot be empty")
	}

	// Reject string "null"
	if strings.EqualFold(email, "null") {
		return false, "", fmt.Errorf("email cannot be 'null'")
	}

	// Trim whitespace
	email = strings.TrimSpace(email)

	// Remove trailing punctuation (periods, commas) that users might add
	email = strings.TrimRight(email, ".,;:")

	// Re-trim whitespace after removing punctuation
	email = strings.TrimSpace(email)

	// Check minimum length
	if len(email) < 5 {
		// Minimum valid email: a@b.c
		return false, "", fmt.Errorf("email is too short")
	}

	// Check maximum length (RFC 5321)
	if len(email) > 254 {
		return false, "", fmt.Errorf("email is too long")
	}

	// Validate format with regex
	if !ev.emailRegex.MatchString(email) {
		return false, "", fmt.Errorf("invalid email format")
	}

	// Additional validation: domain must have at least one dot
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false, "", fmt.Errorf("invalid email format: must contain exactly one @")
	}

	localPart := parts[0]
	domain := parts[1]

	// Check local part (before @)
	if len(localPart) == 0 || len(localPart) > 64 {
		return false, "", fmt.Errorf("invalid email: local part must be 1-64 characters")
	}

	// Check domain part (after @)
	if len(domain) == 0 {
		return false, "", fmt.Errorf("invalid email: domain cannot be empty")
	}

	// Domain must have at least one dot
	if !strings.Contains(domain, ".") {
		return false, "", fmt.Errorf("invalid email: domain must contain at least one dot")
	}

	// Check for consecutive dots
	if strings.Contains(email, "..") {
		return false, "", fmt.Errorf("invalid email: cannot contain consecutive dots")
	}

	// Normalize to lowercase
	normalizedEmail := strings.ToLower(email)

	// Additional security checks
	// Local part cannot start or end with dot
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return false, "", fmt.Errorf("invalid email: local part cannot start or end with dot")
	}

	// Domain part cannot start or end with hyphen or dot
	if strings.HasPrefix(domain, ".") || strings.HasPrefix(domain, "-") ||
		strings.HasSuffix(domain, ".") || strings.HasSuffix(domain, "-") {
		return false, "", fmt.Errorf("invalid email: domain cannot start or end with dot or hyphen")
	}

	return true, normalizedEmail, nil
}

// IsValidEmail is a convenience method that returns only a boolean
func (ev *EmailValidator) IsValidEmail(email string) bool {
	isValid, _, _ := ev.ValidateEmail(email)
	return isValid
}
