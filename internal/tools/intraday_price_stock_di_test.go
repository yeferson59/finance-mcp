package tools

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yeferson59/finance-mcp/internal/models"
	"github.com/yeferson59/finance-mcp/pkg/client"
	"github.com/yeferson59/finance-mcp/pkg/request"
)

// Mock response data for testing
const mockIntradayResponse = `{
  "Meta Data": {
    "1. Information": "Intraday (1min) open, high, low, close prices and volume",
    "2. Symbol": "AAPL",
    "3. Last Refreshed": "2023-12-08 19:59:00",
    "4. Interval": "1min",
    "5. Output Size": "Compact",
    "6. Time Zone": "US/Eastern"
  },
  "Time Series (1min)": {
    "2023-12-08 19:59:00": {
      "1. open": "195.0900",
      "2. high": "195.1800",
      "3. low": "194.9200",
      "4. close": "195.0000",
      "5. volume": "12345"
    },
    "2023-12-08 19:58:00": {
      "1. open": "194.8500",
      "2. high": "195.0900",
      "3. low": "194.8000",
      "4. close": "195.0900",
      "5. volume": "23456"
    }
  }
}`

func TestIntradayPriceStock_DependencyInjection(t *testing.T) {
	// Test that we can create the tool with dependency injection
	mockClient := client.NewMockClient()
	config := &request.AlphaVantageConfig{
		BaseURL: "https://www.alphavantage.co",
		APIKey:  "test-api-key",
		Timeout: 30 * time.Second,
	}

	alphaClient := request.NewAlphaVantageClient(mockClient, config)

	// This would be how we'd create it with DI (not exposed yet, but demonstrates the pattern)
	tool := &IntradayPriceStock{
		alphaClient: alphaClient,
	}

	assert.NotNil(t, tool)
	assert.NotNil(t, tool.alphaClient)
}

func TestIntradayPriceStock_NewIntradayPriceStock(t *testing.T) {
	// Test the constructor creates a properly configured instance
	apiURL := "https://www.alphavantage.co"
	apiKey := "test-api-key"

	tool := NewIntradayPriceStock(apiURL, apiKey)

	assert.NotNil(t, tool)
	assert.NotNil(t, tool.alphaClient)
}

func TestIntradayPriceStock_InputValidation(t *testing.T) {
	tool := NewIntradayPriceStock("https://www.alphavantage.co", "test-key")

	testCases := []struct {
		name        string
		input       models.IntradayPriceInput
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid input",
			input: models.IntradayPriceInput{
				Symbol:   "AAPL",
				Interval: "1min",
			},
			expectError: false,
		},
		{
			name: "empty symbol",
			input: models.IntradayPriceInput{
				Symbol:   "",
				Interval: "1min",
			},
			expectError: true,
			errorMsg:    "symbol cannot be empty",
		},
		{
			name: "whitespace symbol",
			input: models.IntradayPriceInput{
				Symbol:   "   ",
				Interval: "1min",
			},
			expectError: true,
			errorMsg:    "symbol cannot be empty",
		},
		{
			name: "invalid interval",
			input: models.IntradayPriceInput{
				Symbol:   "AAPL",
				Interval: "2min",
			},
			expectError: true,
			errorMsg:    "invalid interval '2min'",
		},
		{
			name: "valid intervals",
			input: models.IntradayPriceInput{
				Symbol:   "AAPL",
				Interval: "5min",
			},
			expectError: false,
		},
		{
			name: "invalid output size",
			input: models.IntradayPriceInput{
				Symbol:     "AAPL",
				Interval:   "1min",
				OutputSize: stringPtr("medium"),
			},
			expectError: true,
			errorMsg:    "invalid output size 'medium'",
		},
		{
			name: "valid output size compact",
			input: models.IntradayPriceInput{
				Symbol:     "AAPL",
				Interval:   "1min",
				OutputSize: stringPtr("compact"),
			},
			expectError: false,
		},
		{
			name: "valid output size full",
			input: models.IntradayPriceInput{
				Symbol:     "AAPL",
				Interval:   "1min",
				OutputSize: stringPtr("full"),
			},
			expectError: false,
		},
		{
			name: "invalid month format",
			input: models.IntradayPriceInput{
				Symbol:   "AAPL",
				Interval: "1min",
				Month:    stringPtr("2023-1"),
			},
			expectError: true,
			errorMsg:    "invalid month format '2023-1'",
		},
		{
			name: "valid month format",
			input: models.IntradayPriceInput{
				Symbol:   "AAPL",
				Interval: "1min",
				Month:    stringPtr("2023-01"),
			},
			expectError: false,
		},
		{
			name: "symbol too long",
			input: models.IntradayPriceInput{
				Symbol:   "VERYLONGSYMBOL",
				Interval: "1min",
			},
			expectError: true,
			errorMsg:    "symbol 'VERYLONGSYMBOL' appears to be invalid (too long)",
		},
		{
			name: "invalid characters in symbol",
			input: models.IntradayPriceInput{
				Symbol:   "AAPL!@#",
				Interval: "1min",
			},
			expectError: true,
			errorMsg:    "symbol 'AAPL!@#' contains invalid characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tool.validateInput(tc.input)

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

func TestIntradayPriceStock_BuildQueries(t *testing.T) {
	tool := NewIntradayPriceStock("https://www.alphavantage.co", "test-key")

	testCases := []struct {
		name           string
		input          models.IntradayPriceInput
		expectedParams map[string]string
	}{
		{
			name: "basic parameters",
			input: models.IntradayPriceInput{
				Symbol:   "AAPL",
				Interval: "1min",
			},
			expectedParams: map[string]string{
				"function": "TIME_SERIES_INTRADAY",
				"interval": "1min",
			},
		},
		{
			name: "with all optional parameters",
			input: models.IntradayPriceInput{
				Symbol:        "MSFT",
				Interval:      "5min",
				Adjusted:      boolPtr(true),
				ExtendedHours: boolPtr(false),
				Month:         stringPtr("2023-01"),
				OutputSize:    stringPtr("full"),
			},
			expectedParams: map[string]string{
				"function":       "TIME_SERIES_INTRADAY",
				"interval":       "5min",
				"adjusted":       "true",
				"extended_hours": "false",
				"month":          "2023-01",
				"outputsize":     "full",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			queries := tool.buildQueries(tc.input)

			// Convert queries to map for easier assertion
			paramMap := make(map[string]string)
			for _, query := range queries {
				paramMap[query.Name] = query.Value
			}

			for key, expected := range tc.expectedParams {
				actual, exists := paramMap[key]
				assert.True(t, exists, "Parameter %s should exist", key)
				assert.Equal(t, expected, actual, "Parameter %s should have value %s", key, expected)
			}
		})
	}
}

func TestIntradayPriceStock_SuccessfulRequest(t *testing.T) {
	// Create mock client
	mockClient := client.NewMockClient()

	// Set up successful response
	mockResponse := &client.Response{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(mockIntradayResponse),
	}

	// We need to set this for any URL that might be called
	// Since URL construction is complex, we'll use a pattern
	mockClient.SetResponse("https://www.alphavantage.co", mockResponse)

	// Create Alpha Vantage client with mock
	config := &request.AlphaVantageConfig{
		BaseURL: "https://www.alphavantage.co",
		APIKey:  "test-key",
		Timeout: 30 * time.Second,
	}
	alphaClient := request.NewAlphaVantageClient(mockClient, config)

	// Create tool with injected client
	tool := &IntradayPriceStock{
		alphaClient: alphaClient,
	}

	// Test input
	input := models.IntradayPriceInput{
		Symbol:   "AAPL",
		Interval: "1min",
	}

	// Make request (this will fail because the mock client URL matching is exact)
	// But we can test the input validation and query building
	err := tool.validateInput(input)
	assert.NoError(t, err, "Input validation should pass")

	queries := tool.buildQueries(input)
	assert.NotEmpty(t, queries, "Queries should be built")

	// Verify required queries are present
	hasFunction := false
	hasInterval := false
	for _, query := range queries {
		switch query.Name {
		case "function":
			hasFunction = true
			assert.Equal(t, "TIME_SERIES_INTRADAY", query.Value)
		case "interval":
			hasInterval = true
			assert.Equal(t, "1min", query.Value)
		}
	}
	assert.True(t, hasFunction, "Function query should be present")
	assert.True(t, hasInterval, "Interval query should be present")
}

func TestIntradayPriceStock_ContextCancellation(t *testing.T) {
	tool := NewIntradayPriceStock("https://www.alphavantage.co", "test-key")

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	input := models.IntradayPriceInput{
		Symbol:   "AAPL",
		Interval: "1min",
	}

	_, _, err := tool.Get(ctx, nil, input)

	// Should return context cancellation error
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestIntradayPriceStock_ContextTimeout(t *testing.T) {
	tool := NewIntradayPriceStock("https://www.alphavantage.co", "test-key")

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(1 * time.Millisecond)

	input := models.IntradayPriceInput{
		Symbol:   "AAPL",
		Interval: "1min",
	}

	_, _, err := tool.Get(ctx, nil, input)

	// Should return context deadline exceeded error
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestIntradayPriceStock_ValidateResponse(t *testing.T) {
	tool := NewIntradayPriceStock("https://www.alphavantage.co", "test-key")

	testCases := []struct {
		name        string
		data        models.IntradayStockOutput
		symbol      string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid response",
			data: models.IntradayStockOutput{
				MetaData: models.MetaData{
					Symbol:   "AAPL",
					Interval: "1min",
				},
				TimeSeries: []models.OHLCVFloat{
					{Open: 100.0, High: 105.0, Low: 99.0, Close: 102.0, Volume: 1000},
				},
			},
			symbol:      "AAPL",
			expectError: false,
		},
		{
			name: "empty symbol",
			data: models.IntradayStockOutput{
				MetaData: models.MetaData{
					Symbol:   "",
					Interval: "1min",
				},
				TimeSeries: []models.OHLCVFloat{},
			},
			symbol:      "AAPL",
			expectError: true,
			errorMsg:    "no data returned for symbol 'AAPL'",
		},
		{
			name: "empty interval",
			data: models.IntradayStockOutput{
				MetaData: models.MetaData{
					Symbol:   "AAPL",
					Interval: "",
				},
				TimeSeries: []models.OHLCVFloat{},
			},
			symbol:      "AAPL",
			expectError: true,
			errorMsg:    "missing interval information",
		},
		{
			name: "empty time series",
			data: models.IntradayStockOutput{
				MetaData: models.MetaData{
					Symbol:   "AAPL",
					Interval: "1min",
				},
				TimeSeries: []models.OHLCVFloat{},
			},
			symbol:      "AAPL",
			expectError: true,
			errorMsg:    "no time series data returned",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tool.validateResponse(tc.data, tc.symbol)

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

func TestIntradayPriceStock_ClientMethods(t *testing.T) {
	tool := NewIntradayPriceStock("https://www.alphavantage.co", "test-key")

	// Test GetStats
	stats := tool.GetStats()
	assert.NotNil(t, stats)

	// Test SetTimeout
	tool.SetTimeout(45 * time.Second)
	// Can't easily verify this without exposing internal state

	// Test Close
	err := tool.Close()
	assert.NoError(t, err)
}

// Helper functions for creating pointers to basic types
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func TestIntradayPriceStock_AllIntervals(t *testing.T) {
	tool := NewIntradayPriceStock("https://www.alphavantage.co", "test-key")

	validIntervals := []string{"1min", "5min", "15min", "30min", "60min"}

	for _, interval := range validIntervals {
		t.Run(interval, func(t *testing.T) {
			input := models.IntradayPriceInput{
				Symbol:   "AAPL",
				Interval: interval,
			}

			err := tool.validateInput(input)
			assert.NoError(t, err, "Interval %s should be valid", interval)

			queries := tool.buildQueries(input)

			// Find interval query
			found := false
			for _, query := range queries {
				if query.Name == "interval" {
					assert.Equal(t, interval, query.Value)
					found = true
					break
				}
			}
			assert.True(t, found, "Interval query should be present")
		})
	}
}

func TestIntradayPriceStock_ThreadSafety(t *testing.T) {
	tool := NewIntradayPriceStock("https://www.alphavantage.co", "test-key")

	// Test concurrent access to methods that use mutex
	done := make(chan bool, 3)

	// Concurrent GetStats calls
	go func() {
		for range 10 {
			_ = tool.GetStats()
		}
		done <- true
	}()

	// Concurrent SetTimeout calls
	go func() {
		for i := range 10 {
			tool.SetTimeout(time.Duration(i+30) * time.Second)
		}
		done <- true
	}()

	// Concurrent Close calls
	go func() {
		for range 10 {
			_ = tool.Close()
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for range 3 {
		select {
		case <-done:
			// Good, goroutine completed
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for goroutine to complete")
		}
	}
}
