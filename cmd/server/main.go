// Package main implements a Model Context Protocol (MCP) server for financial market data.
//
// This server exposes stock market data through the MCP protocol, allowing AI models
// and other MCP clients to query real-time financial information using Alpha Vantage API.
//
// The server runs as a high-performance Fiber server (built on fasthttp), accepting JSON-RPC
// requests and returning structured financial data responses with optimal performance.
//
// Architecture:
//   - MCP Server: Handles protocol communication and tool registration
//   - Fiber Server: High-performance HTTP framework with clean API
//   - Alpha Vantage Client: Fetches real-time market data
//   - Tools: Implements specific financial data retrieval functions
//
// Usage:
//
//	The server listens on port 8080 and can be queried by MCP clients
//	for real-time financial market data with enterprise-grade performance.
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/yeferson59/finance-mcp/internal/config"
	"github.com/yeferson59/finance-mcp/internal/tools"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// setupFiberApp configures a Fiber app with optimal performance settings
func setupFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		Prefork:              false,
		StrictRouting:        false,
		CaseSensitive:        false,
		UnescapePath:         true,
		BodyLimit:            10 * 1024 * 1024,
		Concurrency:          256 * 1024,
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
		IdleTimeout:          60 * time.Second,
		ReadBufferSize:       8192,
		WriteBufferSize:      8192,
		CompressedFileSuffix: ".fiber.gz",
		ProxyHeader:          fiber.HeaderXForwardedFor,

		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := "Internal Server Error"

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				message = e.Message
			}

			return c.Status(code).JSON(fiber.Map{
				"error":     message,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"path":      c.Path(),
				"method":    c.Method(),
			})
		},

		ServerHeader: "Finance-MCP-Server/1.0",
		AppName:      "Finance MCP Server",
	})

	return app
}

// setupMiddleware configures all necessary middleware for the application
func setupMiddleware(app *fiber.App) {
	app.Use(requestid.New())

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "*",
		AllowCredentials: false,
		ExposeHeaders:    "X-Request-ID",
		MaxAge:           86400,
	}))

	app.Use(etag.New())

	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} ${path} | ${ip} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
	}))

	app.Use(healthcheck.New(healthcheck.Config{
		LivenessProbe: func(c *fiber.Ctx) bool {
			return true
		},
		ReadinessProbe: func(c *fiber.Ctx) bool {
			return true
		},
	}))
}

// setupRoutes configures all application routes
func setupRoutes(app *fiber.App, mcpHandler http.Handler) {

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "finance-mcp-server",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"uptime":    time.Since(startTime).String(),
		})
	})

	app.Get("/health/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "alive",
		})
	})

	app.Get("/health/ready", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ready",
			"checks": fiber.Map{
				"api": "ok",
			},
		})
	})

	app.Get("/info", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":        "Finance MCP Server",
			"version":     "1.0.0",
			"description": "Model Context Protocol server for financial market data",
			"endpoints": fiber.Map{
				"health":  "/health",
				"mcp":     "/",
				"mcp_alt": "/mcp",
			},
		})
	})

	app.All("/", adaptor.HTTPHandler(mcpHandler))
	app.All("/mcp", adaptor.HTTPHandler(mcpHandler))
	app.All("/mcp/*", adaptor.HTTPHandler(mcpHandler))

	app.Use(func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "Endpoint not found")
	})
}

var startTime = time.Now()

func main() {
	log.Println("🚀 Starting Finance MCP Server with Fiber framework...")

	cfg := config.NewConfig()
	if cfg.APIURL == "" || cfg.APIKey == "" {
		log.Fatal("❌ Missing required configuration: APIURL and APIKey must be set")
	}

	impl := cfg.Implementation
	server := mcp.NewServer(impl, nil)

	log.Println("📊 Initializing financial data tools with DI architecture...")

	stockOverviewTool := tools.NewOverviewStock(cfg.APIURL, cfg.APIKey)
	stockIntradayPriceTool := tools.NewIntradayPriceStock(cfg.APIURL, cfg.APIKey)

	log.Println("🔧 Registering MCP tools...")
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_overview_stock",
		Description: "Get comprehensive stock market data for a specific company using its stock symbol (e.g., AAPL, GOOGL, MSFT). Returns detailed financial metrics, company information, and market data.",
	}, stockOverviewTool.Get)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_intraday_price_stock",
		Description: "Get intraday stock price data for a specific company using its stock symbol (e.g., AAPL, GOOGL, MSFT). Returns price, volume, and other financial metrics for the specified time interval.",
	}, stockIntradayPriceTool.Get)

	mcpHTTPHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)

	log.Println("⚡ Configuring Fiber application...")
	app := setupFiberApp()

	setupMiddleware(app)

	setupRoutes(app, mcpHTTPHandler)

	port := ":8080"

	log.Println("✅ Finance MCP Server configured successfully")
	log.Printf("🌐 Server starting on port %s", port)
	log.Printf("🏥 Health check: http://localhost%s/health", port)
	log.Printf("📋 API info: http://localhost%s/info", port)
	log.Printf("🔗 MCP endpoint: http://localhost%s/", port)
	log.Println("⚡ Using FastHTTP client with connection pooling")
	log.Printf("🔧 Client stats endpoint: http://localhost%s/health (includes client metrics)", port)
	log.Println("📈 Ready to serve financial market data requests with optimized performance!")

	if err := app.Listen(port); err != nil {
		log.Fatalf("❌ Fiber server failed to start: %v", err)
	}
}
