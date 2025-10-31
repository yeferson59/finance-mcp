// Package validation provides common input validation functions for financial data.
package validation

import (
	"fmt"
	"strings"
)

// ValidateSymbol validates a stock symbol for common patterns and constraints.
// It checks for:
//   - Non-empty symbol
//   - Maximum length of 10 characters
//   - Only alphanumeric characters and dots
//
// Returns nil if valid, error with descriptive message otherwise.
func ValidateSymbol(symbol string) error {
	// Check if empty or whitespace only
	trimmed := strings.TrimSpace(symbol)
	if trimmed == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	// Check length constraint
	if len(trimmed) > 10 {
		return fmt.Errorf("symbol '%s' appears to be invalid (too long)", trimmed)
	}

	// Check for valid characters (alphanumeric and dot)
	for _, char := range trimmed {
		if !((char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '.') {
			return fmt.Errorf("symbol '%s' contains invalid characters", trimmed)
		}
	}

	return nil
}
