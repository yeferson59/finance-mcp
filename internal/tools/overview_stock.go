// Package tools implements MCP tools for financial data retrieval.
//
// This package provides implementations of MCP (Model Context Protocol) tools
// that can be called by AI models and other MCP clients to fetch real-time
// financial market data from external APIs like Alpha Vantage.
package tools

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/yeferson59/finance-mcp/internal/models"
	"github.com/yeferson59/finance-mcp/pkg/parser"
	"github.com/yeferson59/finance-mcp/pkg/requests"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// OverviewStock implements the "get-stock" MCP tool for retrieving comprehensive
// stock market data and company information.
//
// This tool integrates with Alpha Vantage's OVERVIEW function to provide:
//   - Company fundamentals (name, description, sector, industry)
//   - Financial metrics (P/E ratio, market cap, dividends, etc.)
//   - Market data (52-week highs/lows, moving averages)
//   - Performance ratios (ROE, ROA, profit margins)
//
// The tool handles HTTP communication, JSON parsing, error handling, and
// data validation automatically with proper context support for timeouts
// and cancellation.
type OverviewStock struct {
	// APIURL is the base URL for Alpha Vantage API endpoints
	// (typically "https://www.alphavantage.co")
	APIURL string `json:"apiURL"`

	// APIKey is the authentication key for Alpha Vantage API access
	// Required for all API requests
	APIKey string `json:"apiKey"`

	// parser is a reusable JSON parser instance to avoid allocation overhead
	parser *parser.JSON

	// mu protects the parser for concurrent access
	mu sync.RWMutex
}

// NewOverviewStock creates a new OverviewStock tool instance with the provided
// Alpha Vantage API configuration.
//
// Parameters:
//   - apiURL: Base URL for Alpha Vantage API (e.g., "https://www.alphavantage.co")
//   - apiKey: Valid Alpha Vantage API key for authentication
//
// Returns:
//   - Configured OverviewStock instance ready for use as MCP tool
//
// The returned instance includes a preconfigured JSON parser that is reused
// across requests for better performance.
func NewOverviewStock(apiURL, apiKey string) *OverviewStock {
	return &OverviewStock{
		APIURL: apiURL,
		APIKey: apiKey,
		parser: parser.NewJSON(),
	}
}

// validateInput performs input validation on the symbol input
func (os *OverviewStock) validateInput(input models.SymbolInput) error {
	if strings.TrimSpace(input.Symbol) == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	symbol := strings.TrimSpace(input.Symbol)

	if len(symbol) > 10 {
		return fmt.Errorf("symbol '%s' appears to be invalid (too long)", symbol)
	}

	for _, char := range symbol {
		if !((char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '.') {
			return fmt.Errorf("symbol '%s' contains invalid characters", symbol)
		}
	}

	return nil
}

// validateResponse checks if the API response contains error information
func (os *OverviewStock) validateResponse(data models.OverviewOutput, symbol string) error {
	if data.Symbol == "" && data.Name == "" {
		return fmt.Errorf("no data returned for symbol '%s' - symbol may not exist or API limit reached", symbol)
	}

	return nil
}

// Get retrieves comprehensive stock market data for the specified stock symbol.
//
// This method implements the MCP tool interface for the "get-stock" tool,
// handling the complete request lifecycle from input validation to API response parsing.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout handling
//   - req: MCP tool request metadata (unused but required by interface)
//   - input: Stock symbol input containing the ticker to query (e.g., "AAPL", "GOOGL")
//
// Returns:
//   - *mcp.CallToolResult: Always nil (result data is in second return value)
//   - models.OverviewOutput: Complete stock and company data structure
//   - error: Any error encountered during the request or parsing process
//
// Error conditions:
//   - Invalid or empty stock symbol
//   - Context cancellation or timeout
//   - Network connectivity issues
//   - Alpha Vantage API errors (rate limits, invalid API key, etc.)
//   - JSON parsing errors
//   - Invalid response data
//
// The method automatically converts stock symbols to uppercase and handles
// various Alpha Vantage response formats including error responses.
// It respects the context for cancellation and timeout control.
func (os *OverviewStock) Get(ctx context.Context, req *mcp.CallToolRequest, input models.SymbolInput) (*mcp.CallToolResult, models.OverviewOutput, error) {
	if err := os.validateInput(input); err != nil {
		return nil, models.OverviewOutput{}, fmt.Errorf("input validation failed: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, models.OverviewOutput{}, ctx.Err()
	default:
	}

	client := requests.NewAlpha(
		os.APIURL,
		os.APIKey,
		input.Symbol,
		[]requests.Query{
			requests.NewQuery("function", "OVERVIEW"),
		},
	)

	res, err := client.Get()
	if err != nil {
		return nil, models.OverviewOutput{}, fmt.Errorf("failed to fetch stock data for symbol '%s': %w", input.Symbol, err)
	}

	select {
	case <-ctx.Done():
		return nil, models.OverviewOutput{}, ctx.Err()
	default:
	}

	var data models.OverviewOutput

	os.mu.Lock()
	err = os.parser.Parse(&data, bytes.NewReader(res))
	os.mu.Unlock()

	if err != nil {
		return nil, models.OverviewOutput{}, fmt.Errorf("failed to parse stock data for symbol '%s': %w", input.Symbol, err)
	}

	if err := os.validateResponse(data, input.Symbol); err != nil {
		return nil, models.OverviewOutput{}, err
	}

	return nil, data, nil
}
