package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSymbol(t *testing.T) {
	testCases := []struct {
		name        string
		symbol      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid symbol uppercase",
			symbol:      "AAPL",
			expectError: false,
		},
		{
			name:        "valid symbol lowercase",
			symbol:      "aapl",
			expectError: false,
		},
		{
			name:        "valid symbol with dot",
			symbol:      "BRK.A",
			expectError: false,
		},
		{
			name:        "valid symbol with numbers",
			symbol:      "GOOG1",
			expectError: false,
		},
		{
			name:        "empty symbol",
			symbol:      "",
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "whitespace only symbol",
			symbol:      "   ",
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name:        "symbol too long",
			symbol:      "VERYLONGSYMBOL",
			expectError: true,
			errorMsg:    "too long",
		},
		{
			name:        "invalid characters",
			symbol:      "AAPL!",
			expectError: true,
			errorMsg:    "invalid characters",
		},
		{
			name:        "symbol with spaces",
			symbol:      "AA PL",
			expectError: true,
			errorMsg:    "invalid characters",
		},
		{
			name:        "symbol with special characters",
			symbol:      "AAPL@",
			expectError: true,
			errorMsg:    "invalid characters",
		},
		{
			name:        "valid symbol with leading/trailing spaces",
			symbol:      "  AAPL  ",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSymbol(tc.symbol)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func BenchmarkValidateSymbol(b *testing.B) {
	symbols := []string{"AAPL", "GOOGL", "MSFT", "BRK.A", "TSM"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateSymbol(symbols[i%len(symbols)])
	}
}
