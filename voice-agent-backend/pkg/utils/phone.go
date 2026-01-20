package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// PhoneValidator handles phone number validation and formatting
type PhoneValidator struct {
	// Map of country codes to regex patterns
	countryPatterns map[string]string
}

// NewPhoneValidator creates a new phone validator
func NewPhoneValidator() *PhoneValidator {
	return &PhoneValidator{
		countryPatterns: map[string]string{
			"US":   `^(\+1)?[-.\s]?\(?[2-9]\d{2}\)?[-.\s]?\d{3}[-.\s]?\d{4}$`,
			"IN":   `^(\+91)?[-.\s]?[6-9]\d{9}$`,
			"UK":   `^(\+44)?[-.\s]?(?:\(\d+\)|\d+)[-.\s]?\d{3,4}[-.\s]?\d{3,4}$`,
			"CA":   `^(\+1)?[-.\s]?\(?[2-9]\d{2}\)?[-.\s]?\d{3}[-.\s]?\d{4}$`,
			"AU":   `^(\+61)?[-.\s]?(?:2|3|7|8)\d{8}$`,
			"INTL": `^(\+\d{1,3})?[-.\s]?\d{6,14}$`, // Generic international format
		},
	}
}

// ValidatePhoneNumber validates a phone number and returns normalized format
func (pv *PhoneValidator) ValidatePhoneNumber(phone string) (bool, string, error) {
	if phone == "" {
		return false, "", fmt.Errorf("phone number cannot be empty")
	}

	// Remove whitespace
	normalized := strings.TrimSpace(phone)

	// Try to match against patterns
	for _, pattern := range pv.countryPatterns {
		matched, err := regexp.MatchString(pattern, normalized)
		if err != nil {
			return false, "", fmt.Errorf("validation error: %w", err)
		}
		if matched {
			// Normalize the phone number
			normalized = normalizePhoneNumber(normalized)
			return true, normalized, nil
		}
	}

	return false, "", fmt.Errorf("invalid phone number format")
}

// ValidatePhoneNumberWithCountry validates phone number for specific country
func (pv *PhoneValidator) ValidatePhoneNumberWithCountry(phone, countryCode string) (bool, string, error) {
	if phone == "" {
		return false, "", fmt.Errorf("phone number cannot be empty")
	}

	countryCode = strings.ToUpper(countryCode)
	pattern, exists := pv.countryPatterns[countryCode]
	if !exists {
		// Fall back to international pattern
		pattern = pv.countryPatterns["INTL"]
	}

	normalized := strings.TrimSpace(phone)
	matched, err := regexp.MatchString(pattern, normalized)
	if err != nil {
		return false, "", fmt.Errorf("validation error: %w", err)
	}

	if !matched {
		return false, "", fmt.Errorf("invalid phone number for country %s", countryCode)
	}

	normalized = normalizePhoneNumber(normalized)
	return true, normalized, nil
}

// normalizePhoneNumber converts phone to E.164 format (+1234567890)
func normalizePhoneNumber(phone string) string {
	// Remove all non-digit characters except leading +
	normalized := ""
	for i, char := range phone {
		if (char >= '0' && char <= '9') || (i == 0 && char == '+') {
			normalized += string(char)
		}
	}

	// Ensure it starts with +
	if !strings.HasPrefix(normalized, "+") {
		// If no country code, assume +1 (US/Canada)
		if len(normalized) == 10 {
			normalized = "+1" + normalized
		} else if !strings.HasPrefix(normalized, "+") {
			normalized = "+" + normalized
		}
	}

	return normalized
}

// IsValidPhoneFormat checks if phone is in valid format (quick check)
func IsValidPhoneFormat(phone string) bool {
	if len(phone) < 10 {
		return false
	}
	normalized := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '+' {
			return r
		}
		return -1
	}, phone)
	return len(normalized) >= 10 && len(normalized) <= 15
}
