// Package request provides HTTP client implementations for financial APIs.
//
// This package implements Alpha Vantage API client using dependency injection
// pattern with the HTTPClient interface. This design enables better testability,
// maintainability, and separation of concerns.
//
// Key features:
//   - Dependency injection using HTTPClient interface
//   - URL construction with validation
//   - Alpha Vantage specific error handling
//   - Configurable retry and timeout settings
//   - Thread-safe operations
package request

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/yeferson59/finance-mcp/pkg/client"
	"github.com/yeferson59/finance-mcp/pkg/errors"
)

// Query represents a URL query parameter with name and value
type Query struct {
	Name  string
	Value string
}

// NewQuery creates a new Query with the given name and value
func NewQuery(name string, value string) Query {
	return Query{
		Name:  name,
		Value: value,
	}
}

// AlphaVantageConfig holds configuration specific to Alpha Vantage API
type AlphaVantageConfig struct {
	BaseURL   string
	APIKey    string
	UserAgent string
	Timeout   time.Duration
}

// DefaultAlphaVantageConfig returns default configuration for Alpha Vantage API
func DefaultAlphaVantageConfig() *AlphaVantageConfig {
	return &AlphaVantageConfig{
		BaseURL:   "https://www.alphavantage.co/query",
		UserAgent: "Finance-MCP-Server/1.0",
		Timeout:   30 * time.Second,
	}
}

// AlphaVantageClient provides a high-level interface for Alpha Vantage API requests
type AlphaVantageClient struct {
	httpClient client.HTTPClient
	config     *AlphaVantageConfig
}

// NewAlphaVantageClient creates a new Alpha Vantage client with dependency injection
func NewAlphaVantageClient(httpClient client.HTTPClient, config *AlphaVantageConfig) *AlphaVantageClient {
	if config == nil {
		config = DefaultAlphaVantageConfig()
	}

	return &AlphaVantageClient{
		httpClient: httpClient,
		config:     config,
	}
}

// NewDefaultAlphaVantageClient creates a client with FastHTTP implementation and default config
func NewDefaultAlphaVantageClient(apiKey string) *AlphaVantageClient {
	config := DefaultAlphaVantageConfig()
	config.APIKey = apiKey

	httpConfig := client.DefaultConfig()
	httpConfig.UserAgent = config.UserAgent
	httpConfig.ReadTimeout = config.Timeout
	httpConfig.WriteTimeout = config.Timeout

	httpClient := client.NewFastHTTPClient(httpConfig)
	return NewAlphaVantageClient(httpClient, config)
}

// RequestAlpha represents a request to the Alpha Vantage API with modern design patterns
type RequestAlpha struct {
	client  *AlphaVantageClient
	symbol  string
	queries []Query
}

// NewAlpha creates a new Alpha Vantage request instance using the client
// This maintains compatibility with the existing API while using the new client internally
func NewAlpha(baseURL string, apiKey string, symbol string, queries []Query) *RequestAlpha {
	config := &AlphaVantageConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: 30 * time.Second,
	}

	httpConfig := client.DefaultConfig()
	httpClient := client.NewFastHTTPClient(httpConfig)
	alphaClient := NewAlphaVantageClient(httpClient, config)

	return &RequestAlpha{
		client:  alphaClient,
		symbol:  symbol,
		queries: queries,
	}
}

// NewAlphaWithClient creates a new request with a specific Alpha Vantage client
// This is the preferred way when using dependency injection
func NewAlphaWithClient(alphaClient *AlphaVantageClient, symbol string, queries []Query) *RequestAlpha {
	return &RequestAlpha{
		client:  alphaClient,
		symbol:  symbol,
		queries: queries,
	}
}

// validate checks if all required fields are present
func (ra *RequestAlpha) validate() error {
	if strings.TrimSpace(ra.symbol) == "" {
		return errors.ErrSymbolRequired
	}

	if ra.client.config.APIKey == "" {
		return errors.ErrAPIKeyRequired
	}

	if ra.client.config.BaseURL == "" {
		return errors.ErrBaseURLRequired
	}

	return nil
}

// buildURL constructs the complete API URL with all parameters using URLBuilder
func (ra *RequestAlpha) buildURL() (string, error) {
	symbol := strings.ToUpper(strings.TrimSpace(ra.symbol))

	if err := ra.validate(); err != nil {
		return "", err
	}

	builder := client.NewURLBuilder(ra.client.config.BaseURL)

	// Add custom queries
	for _, query := range ra.queries {
		value := query.Value
		// Convert function parameter to uppercase for Alpha Vantage API
		if query.Name == "function" {
			value = strings.ToUpper(value)
		}
		builder.AddParam(query.Name, value)
	}

	builder.AddParam("symbol", symbol)
	builder.AddParam("apikey", ra.client.config.APIKey)

	return builder.Build()
}

// Get performs the HTTP GET request to Alpha Vantage API
func (ra *RequestAlpha) Get() ([]byte, error) {
	return ra.GetWithContext(context.Background())
}

// GetWithContext performs the HTTP GET request with context support
func (ra *RequestAlpha) GetWithContext(ctx context.Context) ([]byte, error) {
	url, err := ra.buildURL()
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	if ctx == context.Background() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ra.client.config.Timeout)
		defer cancel()
	}

	headers := map[string]string{
		"Cache-Control": "no-cache",
		"Accept":        "application/json",
	}

	response, err := ra.client.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	if response.StatusCode != fasthttp.StatusOK {
		switch response.StatusCode {
		case fasthttp.StatusTooManyRequests:
			return nil, fmt.Errorf("API rate limit exceeded (status %d)", response.StatusCode)
		case fasthttp.StatusUnauthorized:
			return nil, fmt.Errorf("invalid API key (status %d)", response.StatusCode)
		case fasthttp.StatusForbidden:
			return nil, fmt.Errorf("access forbidden - check API permissions (status %d)", response.StatusCode)
		default:
			return nil, fmt.Errorf("%w: received status %d", errors.ErrUnexpectedStatusCode, response.StatusCode)
		}
	}

	if err := ra.checkAPIError(response.Body); err != nil {
		return nil, err
	}

	return response.Body, nil
}

// checkAPIError checks if the Alpha Vantage response contains an error message
// Uses bytes.Contains for better performance by avoiding string allocation
func (ra *RequestAlpha) checkAPIError(body []byte) error {
	errorPatterns := []struct {
		pattern []byte
		message string
	}{
		{[]byte("Invalid API call"), "Invalid API function or parameters"},
		{[]byte("the parameter apikey is invalid"), "Invalid API key"},
		{[]byte("higher API call frequency"), "API call frequency limit reached"},
		{[]byte("Thank you for using Alpha Vantage"), "API limit reached - premium key required"},
		{[]byte("Error Message"), "API returned an error"},
	}

	for _, errorPattern := range errorPatterns {
		if bytes.Contains(body, errorPattern.pattern) {
			return fmt.Errorf("API error: %s", errorPattern.message)
		}
	}

	return nil
}

// SetTimeout configures the request timeout
func (ra *RequestAlpha) SetTimeout(timeout time.Duration) *RequestAlpha {
	ra.client.config.Timeout = timeout
	return ra
}

// GetStats returns HTTP client statistics
func (ra *RequestAlpha) GetStats() client.ClientStats {
	return ra.client.httpClient.Stats()
}

// Close cleans up resources
func (ra *RequestAlpha) Close() error {
	return ra.client.httpClient.Close()
}

// AlphaVantageClientPool manages a pool of Alpha Vantage clients for different API keys
type AlphaVantageClientPool struct {
	clients map[string]*AlphaVantageClient
	config  *AlphaVantageConfig
}

// NewAlphaVantageClientPool creates a new client pool
func NewAlphaVantageClientPool(config *AlphaVantageConfig) *AlphaVantageClientPool {
	if config == nil {
		config = DefaultAlphaVantageConfig()
	}

	return &AlphaVantageClientPool{
		clients: make(map[string]*AlphaVantageClient),
		config:  config,
	}
}

// GetClient returns a client for the specified API key, creating it if necessary
func (pool *AlphaVantageClientPool) GetClient(apiKey string) *AlphaVantageClient {
	if client, exists := pool.clients[apiKey]; exists {
		return client
	}

	config := *pool.config
	config.APIKey = apiKey

	httpConfig := client.DefaultConfig()
	httpConfig.UserAgent = config.UserAgent
	httpConfig.ReadTimeout = config.Timeout
	httpConfig.WriteTimeout = config.Timeout

	httpClient := client.NewFastHTTPClient(httpConfig)
	alphaClient := NewAlphaVantageClient(httpClient, &config)

	pool.clients[apiKey] = alphaClient
	return alphaClient
}

// Close closes all clients in the pool
func (pool *AlphaVantageClientPool) Close() error {
	for _, client := range pool.clients {
		if err := client.httpClient.Close(); err != nil {
			return err
		}
	}
	return nil
}

// GetPoolStats returns aggregated statistics for all clients in the pool
func (pool *AlphaVantageClientPool) GetPoolStats() map[string]client.ClientStats {
	stats := make(map[string]client.ClientStats)
	for apiKey, client := range pool.clients {
		stats[apiKey] = client.httpClient.Stats()
	}
	return stats
}

// GetStats returns HTTP client statistics for the Alpha Vantage client
func (ac *AlphaVantageClient) GetStats() client.ClientStats {
	return ac.httpClient.Stats()
}

// Close cleans up resources used by the Alpha Vantage client
func (ac *AlphaVantageClient) Close() error {
	return ac.httpClient.Close()
}

// SetTimeout configures the request timeout for the Alpha Vantage client
func (ac *AlphaVantageClient) SetTimeout(timeout time.Duration) {
	ac.config.Timeout = timeout
}
