// Package main implements a Model Context Protocol (MCP) server for financial market data.
//
// This server exposes stock market data through the MCP protocol, allowing AI models
// and other MCP clients to query real-time financial information using Alpha Vantage API.
//
// The server runs as a stdio-based MCP server, accepting JSON-RPC requests and returning
// structured financial data responses.
//
// Architecture:
//   - MCP Server: Handles protocol communication and tool registration
//   - Alpha Vantage Client: Fetches real-time market data
//   - Tools: Implements specific financial data retrieval functions
//
// Usage:
//
//	The server is designed to be launched by MCP clients (like AI models) and
//	communicates via stdin/stdout using the MCP protocol.
package main

import (
	"context"
	"log"

	"github.com/yeferson59/finance-mcp/internal/config"
	"github.com/yeferson59/finance-mcp/internal/tools"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// main initializes and runs the MCP server for financial market data.
//
// The function performs the following steps:
// 1. Loads configuration (API keys, URLs) from environment variables
// 2. Creates an MCP server instance with stdio transport
// 3. Initializes financial data tools (Alpha Vantage integration)
// 4. Registers available tools with the MCP server
// 5. Starts the server and handles incoming MCP requests
//
// The server will run indefinitely, processing MCP tool calls until terminated.
func main() {
	// Step 1: Load configuration from environment variables
	// This includes API URL and API key for Alpha Vantage service
	cfg := config.NewConfig()

	// Step 2: Extract MCP server implementation details from config
	// This contains server metadata and capabilities
	impl := cfg.Implementation

	// Step 3: Create background context for server operations
	// This context will be used for all MCP protocol operations
	ctx := context.Background()

	// Step 4: Initialize MCP server with implementation details
	// The server handles MCP protocol communication and tool routing
	server := mcp.NewServer(impl, nil)

	// Step 5: Create Alpha Vantage stock overview tool
	// This tool handles HTTP requests to Alpha Vantage API for stock data
	// Note: Both parameters should be cfg.APIURL and cfg.APIKey respectively
	stockOverviewTool := tools.NewOverviewStock(cfg.APIURL, cfg.APIKey)

	// Step 6: Register the "get-stock" tool with the MCP server
	// Tool specification:
	//   - Name: "get-stock" (identifier used by MCP clients)
	//   - Description: Human-readable description for AI models
	//   - Handler: stockOverviewTool.Get (function that processes requests)
	//
	// This tool allows MCP clients to query stock information by providing a symbol
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get-stock",
		Description: "Get comprehensive stock market data for a specific company using its stock symbol (e.g., AAPL, GOOGL, MSFT). Returns detailed financial metrics, company information, and market data.",
	}, stockOverviewTool.Get)

	// Step 7: Start the MCP server with stdio transport
	// The server communicates via stdin/stdout using JSON-RPC over stdio
	// This allows the server to be easily integrated with MCP clients
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatal("Failed to start MCP server:", err)
	}
}
