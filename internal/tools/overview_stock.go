// Package tools implements MCP tools for financial data retrieval.
//
// This package provides implementations of MCP (Model Context Protocol) tools
// that can be called by AI models and other MCP clients to fetch real-time
// financial market data from external APIs like Alpha Vantage.
package tools

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bytedance/sonic"

	"github.com/yeferson59/finance-mcp/internal/models"

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
// data validation automatically.
type OverviewStock struct {
	// APIURL is the base URL for Alpha Vantage API endpoints
	// (typically "https://www.alphavantage.co")
	APIURL string `json:"apiURL"`

	// APIKey is the authentication key for Alpha Vantage API access
	// Required for all API requests
	APIKey string `json:"apiKey"`

	// httpClient is the configured HTTP client for API requests
	// Includes timeout settings for reliable operation
	httpClient *http.Client
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
// The returned instance includes a preconfigured HTTP client with reasonable
// timeout settings (30 seconds) to handle network latency and API response times.
func NewOverviewStock(apiURL, apiKey string) *OverviewStock {
	return &OverviewStock{
		APIURL: apiURL,
		APIKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // 30-second timeout for API requests
		},
	}
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
//   - Network connectivity issues
//   - Alpha Vantage API errors (rate limits, invalid API key, etc.)
//   - JSON parsing errors
//   - HTTP timeout (30 second limit)
//
// The method automatically converts stock symbols to uppercase and handles
// various Alpha Vantage response formats including error responses.
func (os *OverviewStock) Get(ctx context.Context, req *mcp.CallToolRequest, input models.SymbolInput) (*mcp.CallToolResult, models.OverviewOutput, error) {
	// Validate input parameters
	if strings.TrimSpace(input.Symbol) == "" {
		return nil, models.OverviewOutput{}, fmt.Errorf("stock symbol cannot be empty")
	}

	// Normalize symbol to uppercase for API consistency
	symbol := strings.ToUpper(strings.TrimSpace(input.Symbol))

	// Validate symbol format (basic check for common patterns)
	if len(symbol) > 10 {
		return nil, models.OverviewOutput{}, fmt.Errorf("stock symbol '%s' appears to be invalid (too long)", symbol)
	}

	// Construct Alpha Vantage API URL
	apiURL := fmt.Sprintf("%s/query?function=OVERVIEW&symbol=%s&apikey=%s",
		os.APIURL, symbol, os.APIKey)

	// Create HTTP request with context for proper cancellation
	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, models.OverviewOutput{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set appropriate headers for API request
	httpReq.Header.Set("User-Agent", "Simple-MCP-Server/1.0")
	httpReq.Header.Set("Accept", "application/json")

	// Execute HTTP request using configured client
	res, err := os.httpClient.Do(httpReq)
	if err != nil {
		return nil, models.OverviewOutput{}, fmt.Errorf("failed to fetch stock data for symbol '%s': %w", symbol, err)
	}
	defer res.Body.Close()

	// Check HTTP response status
	if res.StatusCode != http.StatusOK {
		return nil, models.OverviewOutput{}, fmt.Errorf("Alpha Vantage API returned status %d for symbol '%s': this may indicate an invalid symbol, API rate limit exceeded, or service unavailability", res.StatusCode, symbol)
	}

	// Parse JSON response
	var data models.OverviewOutput
	decoder := sonic.ConfigDefault.NewDecoder(res.Body)

	if err := decoder.Decode(&data); err != nil {
		return nil, models.OverviewOutput{}, fmt.Errorf("failed to parse stock data response for symbol '%s': %w", symbol, err)
	}

	// Validate that we received actual stock data
	if data.Symbol == "" && data.Name == "" {
		return nil, models.OverviewOutput{}, fmt.Errorf("no stock data found for symbol '%s': symbol may not exist or may not be supported by Alpha Vantage", symbol)
	}

	// Return successful result
	// First return value is nil as per MCP SDK convention - actual data is in second return value
	return nil, data, nil
}
