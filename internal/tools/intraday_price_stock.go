// Package tools implements MCP tools for financial data retrieval.
//
// This package provides implementations of MCP (Model Context Protocol) tools
// that can be called by AI models and other MCP clients to fetch real-time
// financial market data from external APIs like Alpha Vantage.
package tools

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/yeferson59/finance-mcp/internal/models"
	"github.com/yeferson59/finance-mcp/internal/validation"
	"github.com/yeferson59/finance-mcp/pkg/client"
	"github.com/yeferson59/finance-mcp/pkg/parser"
	"github.com/yeferson59/finance-mcp/pkg/request"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// IntradayPriceStock implements the "get-intraday-price-stock" MCP tool for retrieving
// intraday stock price data with time series information.
//
// This tool integrates with Alpha Vantage's TIME_SERIES_INTRADAY function to provide:
//   - Real-time intraday price data (1min, 5min, 15min, 30min, 60min intervals)
//   - Volume information for each time period
//   - OHLC (Open, High, Low, Close) prices
//   - Extended hours trading data (optional)
//   - Historical intraday data for specific months
//   - Adjustable output size (compact vs full)
//
// The tool handles HTTP communication, JSON parsing, error handling, and
// data validation automatically with proper context support for timeouts
// and cancellation.
type IntradayPriceStock struct {
	// alphaClient is the injected Alpha Vantage client
	alphaClient *request.AlphaVantageClient

	// mu protects concurrent access for thread safety
	mu sync.RWMutex
}

// NewIntradayPriceStock creates a new IntradayPriceStock tool instance with the provided
// Alpha Vantage API configuration using dependency injection.
//
// Parameters:
//   - apiURL: Base URL for Alpha Vantage API (e.g., "https://www.alphavantage.co")
//   - apiKey: Valid Alpha Vantage API key for authentication
//
// Returns:
//   - Configured IntradayPriceStock instance ready for use as MCP tool
//
// The returned instance includes a preconfigured HTTP client with optimized
// settings for intraday data retrieval that are reused across requests.
func NewIntradayPriceStock(apiURL, apiKey string) *IntradayPriceStock {
	// Create Alpha Vantage client configuration
	config := &request.AlphaVantageConfig{
		BaseURL: apiURL,
		APIKey:  apiKey,
		Timeout: 30 * time.Second,
	}

	// Create HTTP client with optimized settings for intraday data
	httpConfig := client.DefaultConfig()
	httpConfig.UserAgent = "Finance-MCP-Server/1.0"
	httpConfig.ReadTimeout = 30 * time.Second
	httpConfig.WriteTimeout = 30 * time.Second
	// Intraday data can be large, so we may need higher limits
	httpConfig.MaxResponseBodySize = 20 * 1024 * 1024 // 20MB for large datasets
	httpClient := client.NewFastHTTPClient(httpConfig)

	// Create Alpha Vantage client with dependency injection
	alphaClient := request.NewAlphaVantageClient(httpClient, config)

	return &IntradayPriceStock{
		alphaClient: alphaClient,
	}
}

// validateInput performs comprehensive input validation on the intraday price input
func (s *IntradayPriceStock) validateInput(input models.IntradayPriceInput) error {
	// Validate symbol using shared validation
	if err := validation.ValidateSymbol(input.Symbol); err != nil {
		return err
	}

	// Validate interval
	validIntervals := []string{"1min", "5min", "15min", "30min", "60min"}
	if !slices.Contains(validIntervals, input.Interval) {
		return fmt.Errorf("invalid interval '%s'. Valid intervals are: %s",
			input.Interval, strings.Join(validIntervals, ", "))
	}

	// Validate output size if provided
	if input.OutputSize != nil {
		validOutputSizes := []string{"compact", "full"}
		outputSizeValid := false

		if slices.Contains(validOutputSizes, *input.OutputSize) {
			outputSizeValid = true
		}

		if !outputSizeValid {
			return fmt.Errorf("invalid output size '%s'. Valid sizes are: %s",
				*input.OutputSize, strings.Join(validOutputSizes, ", "))
		}
	}

	// Validate month format if provided (should be YYYY-MM)
	if input.Month != nil {
		month := *input.Month
		if len(month) != 7 || month[4] != '-' {
			return fmt.Errorf("invalid month format '%s'. Expected format: YYYY-MM", month)
		}
		// Additional validation could check if it's a valid date
	}

	return nil
}

// buildQueries constructs the query parameters for the Alpha Vantage API request
func (s *IntradayPriceStock) buildQueries(input models.IntradayPriceInput) []request.Query {
	queries := []request.Query{
		request.NewQuery("function", "TIME_SERIES_INTRADAY"),
		request.NewQuery("interval", input.Interval),
	}

	// Add optional parameters
	if input.Adjusted != nil {
		queries = append(queries, request.NewQuery("adjusted", fmt.Sprintf("%t", *input.Adjusted)))
	}

	if input.ExtendedHours != nil {
		queries = append(queries, request.NewQuery("extended_hours", fmt.Sprintf("%t", *input.ExtendedHours)))
	}

	if input.Month != nil {
		queries = append(queries, request.NewQuery("month", *input.Month))
	}

	if input.OutputSize != nil {
		queries = append(queries, request.NewQuery("outputsize", *input.OutputSize))
	}

	return queries
}

// Get retrieves intraday stock price data for the specified stock symbol and parameters.
//
// This method implements the MCP tool interface for the "get-intraday-price-stock" tool,
// handling the complete request lifecycle from input validation to API response parsing.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout handling
//   - req: MCP tool request metadata (unused but required by interface)
//   - input: Intraday price input containing symbol, interval, and optional parameters
//
// Returns:
//   - *mcp.CallToolResult: Always nil (result data is in second return value)
//   - models.IntradayStockOutput: Complete intraday stock data with time series
//   - error: Any error encountered during the request or parsing process
//
// Error conditions:
//   - Invalid or empty stock symbol
//   - Invalid interval specification
//   - Context cancellation or timeout
//   - Network connectivity issues
//   - Alpha Vantage API errors (rate limits, invalid API key, etc.)
//   - JSON parsing errors
//   - Invalid response data
//
// The method automatically converts stock symbols to uppercase and handles
// various Alpha Vantage response formats including error responses.
// It respects the context for cancellation and timeout control.
func (s *IntradayPriceStock) Get(ctx context.Context, req *mcp.CallToolRequest, input models.IntradayPriceInput) (*mcp.CallToolResult, models.IntradayStockOutput, error) {
	// Validate input before making any external requests
	if err := s.validateInput(input); err != nil {
		return nil, models.IntradayStockOutput{}, fmt.Errorf("input validation failed: %w", err)
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return nil, models.IntradayStockOutput{}, ctx.Err()
	default:
	}

	// Build query parameters
	queries := s.buildQueries(input)

	// Create request with proper query parameters using injected client
	requestClient := request.NewAlphaWithClient(
		s.alphaClient,
		input.Symbol,
		queries,
	)

	// Make API request with context support
	res, err := requestClient.GetWithContext(ctx)
	if err != nil {
		return nil, models.IntradayStockOutput{}, fmt.Errorf("failed to fetch intraday data for symbol '%s': %w", input.Symbol, err)
	}

	// Check context again before parsing (in case request took a long time)
	select {
	case <-ctx.Done():
		return nil, models.IntradayStockOutput{}, ctx.Err()
	default:
	}

	// Parse the raw intraday data using the specialized parser
	rawData, err := parser.IntradayPrices(res)
	if err != nil {
		return nil, models.IntradayStockOutput{}, fmt.Errorf("failed to parse intraday data for symbol '%s': %w", input.Symbol, err)
	}

	// Process the time series data into the final output format
	data, err := rawData.ProcessTimeSeries()
	if err != nil {
		return nil, models.IntradayStockOutput{}, fmt.Errorf("failed to process time series data for symbol '%s': %w", input.Symbol, err)
	}

	// Validate that we received data
	if err := s.validateResponse(*data, input.Symbol); err != nil {
		return nil, models.IntradayStockOutput{}, err
	}

	// Return successful result
	return nil, *data, nil
}

// validateResponse checks if the API response contains valid data
func (s *IntradayPriceStock) validateResponse(data models.IntradayStockOutput, symbol string) error {
	// Check if response contains basic required fields
	if data.MetaData.Symbol == "" {
		return fmt.Errorf("no data returned for symbol '%s' - symbol may not exist or API limit reached", symbol)
	}

	if data.MetaData.Interval == "" {
		return fmt.Errorf("invalid response: missing interval information for symbol '%s'", symbol)
	}

	if len(data.TimeSeries) == 0 {
		return fmt.Errorf("no time series data returned for symbol '%s' - check if market is open or try a different time period", symbol)
	}

	return nil
}

// GetStats returns HTTP client statistics for monitoring
func (s *IntradayPriceStock) GetStats() client.ClientStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alphaClient.GetStats()
}

// Close cleans up resources used by the tool
func (s *IntradayPriceStock) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.alphaClient.Close()
}

// SetTimeout configures the request timeout for this tool instance
func (s *IntradayPriceStock) SetTimeout(timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alphaClient.SetTimeout(timeout)
}
