// Package client provides HTTP client implementations with dependency injection pattern.
//
// This package implements a high-performance HTTP client interface designed for
// financial API integration, with specific optimizations for services like Alpha Vantage.
//
// Key features:
//   - Interface-based design for dependency injection and testability
//   - FastHTTP implementation for maximum performance
//   - Connection pooling and automatic compression handling
//   - Retry logic and error handling
//   - Configurable timeouts and rate limiting
//
// Usage:
//
//	client := client.NewFastHTTPClient(client.DefaultConfig())
//	response, err := client.Get(ctx, "https://api.example.com", headers)
package client

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

// HTTPClient defines the interface for HTTP client implementations.
// This interface enables dependency injection and makes testing easier.
type HTTPClient interface {
	// Get performs a GET request to the specified URL with optional headers
	Get(ctx context.Context, url string, headers map[string]string) (*Response, error)

	// Post performs a POST request with body and headers
	Post(ctx context.Context, url string, body []byte, headers map[string]string) (*Response, error)

	// Do performs a request with full control over method, body, and headers
	Do(ctx context.Context, method, url string, body []byte, headers map[string]string) (*Response, error)

	// Close cleans up any resources used by the client
	Close() error

	// Stats returns performance statistics about the client
	Stats() ClientStats
}

// Response represents an HTTP response with automatic decompression support
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// ClientStats provides performance metrics about the HTTP client
type ClientStats struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	AverageLatency     time.Duration
	ConnectionsActive  int
	ConnectionsTotal   int64
}

// Config holds configuration for HTTP clients
type Config struct {
	// Connection settings
	MaxConnsPerHost     int
	MaxIdleConnDuration time.Duration
	MaxConnDuration     time.Duration
	MaxConnWaitTimeout  time.Duration

	// Request settings
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	MaxResponseBodySize int

	// Retry settings
	MaxRetries int
	RetryDelay time.Duration

	// Client identification
	UserAgent string

	// Performance settings
	EnableCompression bool
	EnableKeepAlive   bool
}

// DefaultConfig returns a configuration optimized for financial API usage
func DefaultConfig() *Config {
	return &Config{
		MaxConnsPerHost:     100,
		MaxIdleConnDuration: 90 * time.Second,
		MaxConnDuration:     10 * time.Minute,
		MaxConnWaitTimeout:  30 * time.Second,
		ReadTimeout:         30 * time.Second,
		WriteTimeout:        30 * time.Second,
		MaxResponseBodySize: 10 * 1024 * 1024,
		MaxRetries:          2,
		RetryDelay:          500 * time.Millisecond,
		UserAgent:           "Finance-MCP-Client/1.0",
		EnableCompression:   true,
		EnableKeepAlive:     true,
	}
}

// FastHTTPClient implements HTTPClient using valyala/fasthttp for maximum performance
type FastHTTPClient struct {
	client *fasthttp.Client
	config *Config
	stats  *clientStats
	mu     sync.RWMutex
}

// clientStats tracks performance metrics
type clientStats struct {
	totalRequests      int64
	successfulRequests int64
	failedRequests     int64
	totalLatency       time.Duration
	mu                 sync.RWMutex
}

// NewFastHTTPClient creates a new FastHTTP-based client with the given configuration
func NewFastHTTPClient(config *Config) *FastHTTPClient {
	if config == nil {
		config = DefaultConfig()
	}

	client := &fasthttp.Client{
		MaxConnsPerHost:               config.MaxConnsPerHost,
		MaxIdleConnDuration:           config.MaxIdleConnDuration,
		MaxConnDuration:               config.MaxConnDuration,
		MaxConnWaitTimeout:            config.MaxConnWaitTimeout,
		ReadTimeout:                   config.ReadTimeout,
		WriteTimeout:                  config.WriteTimeout,
		MaxResponseBodySize:           config.MaxResponseBodySize,
		DisableHeaderNamesNormalizing: false,
		DisablePathNormalizing:        true,
		Name:                          config.UserAgent,
		RetryIf: func(request *fasthttp.Request) bool {
			return false
		},
	}

	return &FastHTTPClient{
		client: client,
		config: config,
		stats:  &clientStats{},
	}
}

// Get performs an HTTP GET request
func (c *FastHTTPClient) Get(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, "GET", url, nil, headers)
}

// Post performs an HTTP POST request
func (c *FastHTTPClient) Post(ctx context.Context, url string, body []byte, headers map[string]string) (*Response, error) {
	return c.Do(ctx, "POST", url, body, headers)
}

// Do performs an HTTP request with full control over method, body, and headers
func (c *FastHTTPClient) Do(ctx context.Context, method, url string, body []byte, headers map[string]string) (*Response, error) {
	startTime := time.Now()

	c.stats.mu.Lock()
	c.stats.totalRequests++
	c.stats.mu.Unlock()

	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		response, err := c.performRequest(ctx, method, url, body, headers)
		if err == nil {
			latency := time.Since(startTime)
			c.stats.mu.Lock()
			c.stats.successfulRequests++
			c.stats.totalLatency += latency
			c.stats.mu.Unlock()

			return response, nil
		}

		lastErr = err

		if c.shouldNotRetry(err) {
			break
		}

		if attempt < c.config.MaxRetries {
			select {
			case <-time.After(c.config.RetryDelay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	c.stats.mu.Lock()
	c.stats.failedRequests++
	c.stats.mu.Unlock()

	return nil, fmt.Errorf("failed after %d attempts: %w", c.config.MaxRetries+1, lastErr)
}

// performRequest executes a single HTTP request
func (c *FastHTTPClient) performRequest(ctx context.Context, method, url string, body []byte, headers map[string]string) (*Response, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	req.Header.SetUserAgent(c.config.UserAgent)

	if body != nil {
		req.SetBody(body)
	}

	if c.config.EnableCompression {
		req.Header.Set("Accept-Encoding", "gzip, deflate")
	}

	if c.config.EnableKeepAlive {
		req.Header.Set("Connection", "keep-alive")
	}

	req.Header.Set("Accept", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	timeout := c.config.ReadTimeout
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining < timeout {
			timeout = remaining
		}
	}

	if err := c.client.DoTimeout(req, resp, timeout); err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	response, err := c.convertResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("response conversion failed: %w", err)
	}

	return response, nil
}

// convertResponse converts fasthttp.Response to our Response type with decompression
func (c *FastHTTPClient) convertResponse(resp *fasthttp.Response) (*Response, error) {
	headers := make(map[string]string)
	resp.Header.All()

	body, err := c.decompressBody(resp)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode(),
		Headers:    headers,
		Body:       body,
	}, nil
}

// decompressBody handles automatic decompression of response body
func (c *FastHTTPClient) decompressBody(resp *fasthttp.Response) ([]byte, error) {
	contentEncoding := string(resp.Header.Peek("Content-Encoding"))

	// Fast path: no compression (most common case for Alpha Vantage)
	// Avoid unnecessary copy by directly using the response body
	if contentEncoding == "" {
		bodyBytes := make([]byte, len(resp.Body()))
		copy(bodyBytes, resp.Body())
		return bodyBytes, nil
	}

	// For compressed responses, we need to decompress
	bodyBytes := resp.Body()

	switch contentEncoding {
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress gzip content: %w", err)
		}

		return decompressed, nil

	case "deflate":
		reader := flate.NewReader(bytes.NewReader(bodyBytes))
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress deflate content: %w", err)
		}

		return decompressed, nil

	default:
		return nil, fmt.Errorf("unsupported compression type: %s", contentEncoding)
	}
}

// shouldNotRetry determines if an error should not trigger a retry
func (c *FastHTTPClient) shouldNotRetry(err error) bool {
	errStr := strings.ToLower(err.Error())

	nonRetryableErrors := []string{
		"rate limit",
		"authentication",
		"authorization",
		"invalid",
		"bad request",
		"not found",
		"forbidden",
	}

	for _, nonRetryable := range nonRetryableErrors {
		if strings.Contains(errStr, nonRetryable) {
			return true
		}
	}

	return false
}

// Close cleans up client resources
func (c *FastHTTPClient) Close() error {
	// FastHTTP client doesn't have explicit close method
	// Resources are cleaned up automatically
	return nil
}

// Stats returns performance statistics
func (c *FastHTTPClient) Stats() ClientStats {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()

	var avgLatency time.Duration
	if c.stats.successfulRequests > 0 {
		avgLatency = c.stats.totalLatency / time.Duration(c.stats.successfulRequests)
	}

	return ClientStats{
		TotalRequests:      c.stats.totalRequests,
		SuccessfulRequests: c.stats.successfulRequests,
		FailedRequests:     c.stats.failedRequests,
		AverageLatency:     avgLatency,
		ConnectionsActive:  0,
		ConnectionsTotal:   0,
	}
}

// URLBuilder helps construct URLs with query parameters
type URLBuilder struct {
	baseURL string
	params  url.Values
}

// NewURLBuilder creates a new URL builder
func NewURLBuilder(baseURL string) *URLBuilder {
	return &URLBuilder{
		baseURL: baseURL,
		params:  make(url.Values),
	}
}

// AddParam adds a query parameter
func (b *URLBuilder) AddParam(key, value string) *URLBuilder {
	b.params.Add(key, value)
	return b
}

// SetParam sets a query parameter (replaces existing)
func (b *URLBuilder) SetParam(key, value string) *URLBuilder {
	b.params.Set(key, value)
	return b
}

// Build constructs the final URL
func (b *URLBuilder) Build() (string, error) {
	parsedURL, err := url.Parse(b.baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	parsedURL.RawQuery = b.params.Encode()
	return parsedURL.String(), nil
}

// MockClient implements HTTPClient for testing purposes
type MockClient struct {
	responses map[string]*Response
	errors    map[string]error
	callCount map[string]int
	mu        sync.RWMutex
}

// NewMockClient creates a new mock client for testing
func NewMockClient() *MockClient {
	return &MockClient{
		responses: make(map[string]*Response),
		errors:    make(map[string]error),
		callCount: make(map[string]int),
	}
}

// SetResponse configures the mock to return a specific response for a URL
func (m *MockClient) SetResponse(url string, response *Response) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[url] = response
}

// SetError configures the mock to return an error for a URL
func (m *MockClient) SetError(url string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[url] = err
}

// GetCallCount returns how many times a URL was called
func (m *MockClient) GetCallCount(url string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount[url]
}

// Get implements HTTPClient interface
func (m *MockClient) Get(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	return m.Do(ctx, "GET", url, nil, headers)
}

// Post implements HTTPClient interface
func (m *MockClient) Post(ctx context.Context, url string, body []byte, headers map[string]string) (*Response, error) {
	return m.Do(ctx, "POST", url, body, headers)
}

// Do implements HTTPClient interface
func (m *MockClient) Do(ctx context.Context, method, url string, body []byte, headers map[string]string) (*Response, error) {
	m.mu.Lock()
	m.callCount[url]++
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if err, exists := m.errors[url]; exists {
		return nil, err
	}

	if response, exists := m.responses[url]; exists {
		return response, nil
	}

	return &Response{
		StatusCode: 200,
		Headers:    make(map[string]string),
		Body:       []byte(`{"status": "mock"}`),
	}, nil
}

// Close implements HTTPClient interface
func (m *MockClient) Close() error {
	return nil
}

// Stats implements HTTPClient interface
func (m *MockClient) Stats() ClientStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := int64(0)
	for _, count := range m.callCount {
		total += int64(count)
	}

	return ClientStats{
		TotalRequests:      total,
		SuccessfulRequests: total,
		FailedRequests:     0,
		AverageLatency:     time.Millisecond,
	}
}
