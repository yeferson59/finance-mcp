// Package tools implements MCP tools for financial data retrieval.
//
// This package provides implementations of MCP (Model Context Protocol) tools
// that can be called by AI models and other MCP clients to fetch real-time
// financial market data from external APIs like Alpha Vantage.
package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/yeferson59/finance-mcp/internal/models"
	"github.com/yeferson59/finance-mcp/internal/validation"
	"github.com/yeferson59/finance-mcp/pkg/client"
	"github.com/yeferson59/finance-mcp/pkg/parser"
	"github.com/yeferson59/finance-mcp/pkg/request"

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
	// alphaClient is the injected Alpha Vantage client
	alphaClient *request.AlphaVantageClient

	// parser is a reusable JSON parser instance to avoid allocation overhead
	// Note: sonic parser is already thread-safe, no mutex needed
	parser *parser.JSON
}

// NewOverviewStock creates a new OverviewStock tool instance with the provided
// Alpha Vantage API configuration using dependency injection.
//
// Parameters:
//   - apiURL: Base URL for Alpha Vantage API (e.g., "https://www.alphavantage.co")
//   - apiKey: Valid Alpha Vantage API key for authentication
//
// Returns:
//   - Configured OverviewStock instance ready for use as MCP tool
//
// The returned instance includes a preconfigured JSON parser and HTTP client
// that are reused across requests for better performance.
func NewOverviewStock(apiURL, apiKey string) *OverviewStock {
	config := &request.AlphaVantageConfig{
		BaseURL: apiURL,
		APIKey:  apiKey,
		Timeout: 30 * time.Second,
	}

	httpConfig := client.DefaultConfig()
	httpConfig.UserAgent = "Finance-MCP-Server/1.0"
	httpClient := client.NewFastHTTPClient(httpConfig)
	alphaClient := request.NewAlphaVantageClient(httpClient, config)

	return &OverviewStock{
		alphaClient: alphaClient,
		parser:      parser.NewJSON(),
	}
}

// validateInput performs input validation on the symbol input
func (os *OverviewStock) validateInput(input models.SymbolInput) error {
	return validation.ValidateSymbol(input.Symbol)
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

	requestClient := request.NewAlphaWithClient(
		os.alphaClient,
		input.Symbol,
		[]request.Query{
			request.NewQuery("function", "OVERVIEW"),
		},
	)

	// Make API request with context support
	res, err := requestClient.GetWithContext(ctx)
	if err != nil {
		return nil, models.OverviewOutput{}, fmt.Errorf("failed to fetch stock data for symbol '%s': %w", input.Symbol, err)
	}

	select {
	case <-ctx.Done():
		return nil, models.OverviewOutput{}, ctx.Err()
	default:
	}

	var data models.OverviewOutput

	// sonic parser is already thread-safe, no lock needed
	err = os.parser.ParseBytes(&data, res)
	if err != nil {
		return nil, models.OverviewOutput{}, fmt.Errorf("failed to parse stock data for symbol '%s': %w", input.Symbol, err)
	}

	if err := os.validateResponse(data, input.Symbol); err != nil {
		return nil, models.OverviewOutput{}, err
	}

	return nil, data, nil
}
