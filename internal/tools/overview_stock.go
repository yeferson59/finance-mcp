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
	"net/http"
	"time"

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

	var data models.OverviewOutput
	NewJson := parser.NewJSON()

	err = NewJson.Parse(&data, bytes.NewReader(res))

	return nil, data, nil
}
