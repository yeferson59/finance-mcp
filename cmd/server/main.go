// Package main implements a Model Context Protocol (MCP) server for financial market data.
//
// This server exposes stock market data through the MCP protocol, allowing AI models
// and other MCP clients to query real-time financial information using Alpha Vantage API.
//
// The server runs as a high-performance fasthttp server, accepting JSON-RPC requests and returning
// structured financial data responses with optimal performance characteristics.
//
// Architecture:
//   - MCP Server: Handles protocol communication and tool registration
//   - FastHTTP Server: High-performance HTTP server with adapter for MCP handlers
//   - Alpha Vantage Client: Fetches real-time market data
//   - Tools: Implements specific financial data retrieval functions
//
// Usage:
//
//	The server listens on port 8080 and can be queried by MCP clients
//	for real-time financial market data with optimized performance.
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/yeferson59/finance-mcp/internal/config"
	"github.com/yeferson59/finance-mcp/internal/tools"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FastHTTPAdapter adapts net/http handlers to work with fasthttp
// This allows us to use MCP's HTTP handlers with fasthttp for better performance
type FastHTTPAdapter struct {
	httpHandler http.Handler
}

// NewFastHTTPAdapter creates a new adapter for converting net/http handlers to fasthttp
func NewFastHTTPAdapter(handler http.Handler) *FastHTTPAdapter {
	return &FastHTTPAdapter{
		httpHandler: handler,
	}
}

// Handler converts fasthttp.RequestCtx to net/http Request/Response and delegates to the wrapped handler
func (a *FastHTTPAdapter) Handler(ctx *fasthttp.RequestCtx) {
	// Convert fasthttp request to net/http request
	req, err := a.convertToHTTPRequest(ctx)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(fmt.Sprintf("Error converting request: %v", err))
		return
	}

	// Create a response writer that captures the response
	rw := &responseWriter{
		header: make(http.Header),
		body:   &bytes.Buffer{},
	}

	// Call the original handler
	a.httpHandler.ServeHTTP(rw, req)

	// Convert response back to fasthttp
	a.convertToFastHTTPResponse(ctx, rw)
}

// convertToHTTPRequest converts fasthttp.RequestCtx to *http.Request
func (a *FastHTTPAdapter) convertToHTTPRequest(ctx *fasthttp.RequestCtx) (*http.Request, error) {
	var body io.Reader
	if len(ctx.PostBody()) > 0 {
		body = bytes.NewReader(ctx.PostBody())
	}

	req, err := http.NewRequest(string(ctx.Method()), ctx.URI().String(), body)
	if err != nil {
		return nil, err
	}

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		req.Header.Add(string(key), string(value))
	})

	if len(ctx.PostBody()) > 0 {
		req.ContentLength = int64(len(ctx.PostBody()))
	}

	req.RemoteAddr = ctx.RemoteAddr().String()

	return req, nil
}

// convertToFastHTTPResponse converts the captured HTTP response to fasthttp response
func (a *FastHTTPAdapter) convertToFastHTTPResponse(ctx *fasthttp.RequestCtx, rw *responseWriter) {

	ctx.SetStatusCode(rw.statusCode)

	for key, values := range rw.header {
		for _, value := range values {
			ctx.Response.Header.Add(key, value)
		}
	}

	ctx.SetBody(rw.body.Bytes())
}

// responseWriter implements http.ResponseWriter to capture the response
type responseWriter struct {
	header     http.Header
	body       *bytes.Buffer
	statusCode int
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.body.Write(data)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

// Additional methods to satisfy http.ResponseWriter interface completely
func (rw *responseWriter) WriteString(s string) (int, error) {
	return rw.Write([]byte(s))
}

// setupCORS configures CORS headers for the response
func setupCORS(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	ctx.Response.Header.Set("Access-Control-Max-Age", "86400")
}

// healthCheckHandler provides a simple health check endpoint
func healthCheckHandler(ctx *fasthttp.RequestCtx) {
	if string(ctx.Path()) == "/health" {
		setupCORS(ctx)
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		ctx.SetBodyString(`{"status":"ok","service":"finance-mcp-server"}`)
		return
	}
}

func main() {
	log.Println("Starting Finance MCP Server with FastHTTP...")

	cfg := config.NewConfig()
	impl := cfg.Implementation
	server := mcp.NewServer(impl, nil)

	stockOverviewTool := tools.NewOverviewStock(cfg.APIURL, cfg.APIKey)
	stockIntradayPriceTool := tools.NewIntradayPriceStock(cfg.APIURL, cfg.APIKey)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_overview_stock",
		Description: "Get comprehensive stock market data for a specific company using its stock symbol (e.g., AAPL, GOOGL, MSFT). Returns detailed financial metrics, company information, and market data.",
	}, stockOverviewTool.Get)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_intraday_price_stock",
		Description: "Get intraday stock price data for a specific company using its stock symbol (e.g., AAPL, GOOGL, MSFT). Returns price, volume, and other financial metrics for the specified time interval.",
	}, stockIntradayPriceTool.Get)

	mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)

	adapter := NewFastHTTPAdapter(mcpHandler)

	mainHandler := func(ctx *fasthttp.RequestCtx) {

		setupCORS(ctx)

		if string(ctx.Method()) == "OPTIONS" {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		}

		if string(ctx.Path()) == "/health" {
			healthCheckHandler(ctx)
			return
		}

		if strings.HasPrefix(string(ctx.Path()), "/mcp") || string(ctx.Path()) == "/" {
			adapter.Handler(ctx)
			return
		}

		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		ctx.SetBodyString(`{"error":"endpoint not found"}`)
	}

	server_config := &fasthttp.Server{
		Handler:                       mainHandler,
		DisableKeepalive:              false,
		ReadTimeout:                   30000, // 30 seconds
		WriteTimeout:                  30000, // 30 seconds
		IdleTimeout:                   60000, // 60 seconds
		MaxConnsPerIP:                 1000,
		MaxRequestsPerConn:            1000,
		MaxRequestBodySize:            10 * 1024 * 1024, // 10MB
		ReduceMemoryUsage:             true,
		TCPKeepalive:                  true,
		NoDefaultServerHeader:         true,
		NoDefaultContentType:          true,
		DisableHeaderNamesNormalizing: false,
	}

	server_config.Name = "Finance-MCP-Server/1.0"

	port := ":8080"
	log.Printf("FastHTTP server starting on port %s", port)
	log.Printf("Health check available at: http://localhost%s/health", port)
	log.Printf("MCP endpoint available at: http://localhost%s/", port)

	if err := server_config.ListenAndServe(port); err != nil {
		log.Fatalf("FastHTTP server failed to start: %v", err)
	}
}
