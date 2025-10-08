package client

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Verify default values
	if config.MaxConnsPerHost != 100 {
		t.Errorf("Expected MaxConnsPerHost to be 100, got %d", config.MaxConnsPerHost)
	}

	if config.ReadTimeout != 30*time.Second {
		t.Errorf("Expected ReadTimeout to be 30s, got %v", config.ReadTimeout)
	}

	if config.UserAgent != "Finance-MCP-Client/1.0" {
		t.Errorf("Expected UserAgent to be 'Finance-MCP-Client/1.0', got %s", config.UserAgent)
	}

	if !config.EnableCompression {
		t.Error("Expected EnableCompression to be true")
	}

	if !config.EnableKeepAlive {
		t.Error("Expected EnableKeepAlive to be true")
	}
}

func TestFastHTTPClient_Creation(t *testing.T) {
	// Test with default config
	client := NewFastHTTPClient(nil)
	if client == nil {
		t.Fatal("NewFastHTTPClient(nil) returned nil")
	}

	// Test with custom config
	config := &Config{
		MaxConnsPerHost: 50,
		ReadTimeout:     15 * time.Second,
		UserAgent:       "TestClient/1.0",
	}

	client = NewFastHTTPClient(config)
	if client == nil {
		t.Fatal("NewFastHTTPClient(config) returned nil")
	}

	if client.config.MaxConnsPerHost != 50 {
		t.Errorf("Expected MaxConnsPerHost to be 50, got %d", client.config.MaxConnsPerHost)
	}
}

func TestMockClient_BasicFunctionality(t *testing.T) {
	ctx := context.Background()
	mock := NewMockClient()

	// Test default response
	resp, err := mock.Get(ctx, "https://example.com", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if string(resp.Body) != `{"status": "mock"}` {
		t.Errorf("Unexpected response body: %s", string(resp.Body))
	}

	// Verify call count
	if mock.GetCallCount("https://example.com") != 1 {
		t.Errorf("Expected call count 1, got %d", mock.GetCallCount("https://example.com"))
	}
}

func TestMockClient_CustomResponses(t *testing.T) {
	ctx := context.Background()
	mock := NewMockClient()

	// Set custom response
	customResp := &Response{
		StatusCode: 201,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(`{"message": "created"}`),
	}
	mock.SetResponse("https://api.example.com", customResp)

	// Test custom response
	resp, err := mock.Get(ctx, "https://api.example.com", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Errorf("Expected status code 201, got %d", resp.StatusCode)
	}

	if string(resp.Body) != `{"message": "created"}` {
		t.Errorf("Unexpected response body: %s", string(resp.Body))
	}

	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header, got %v", resp.Headers)
	}
}

func TestMockClient_Errors(t *testing.T) {
	ctx := context.Background()
	mock := NewMockClient()

	// Set error for specific URL
	expectedErr := fmt.Errorf("network error")
	mock.SetError("https://error.example.com", expectedErr)

	// Test error response
	resp, err := mock.Get(ctx, "https://error.example.com", nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp != nil {
		t.Error("Expected nil response on error")
	}

	if err.Error() != "network error" {
		t.Errorf("Expected 'network error', got %v", err)
	}

	// Verify call count is still tracked
	if mock.GetCallCount("https://error.example.com") != 1 {
		t.Errorf("Expected call count 1, got %d", mock.GetCallCount("https://error.example.com"))
	}
}

func TestMockClient_AllMethods(t *testing.T) {
	ctx := context.Background()
	mock := NewMockClient()

	testURL := "https://api.example.com/test"

	// Test GET
	_, err := mock.Get(ctx, testURL, nil)
	if err != nil {
		t.Errorf("GET failed: %v", err)
	}

	// Test POST
	_, err = mock.Post(ctx, testURL, []byte("test body"), nil)
	if err != nil {
		t.Errorf("POST failed: %v", err)
	}

	// Test Do
	_, err = mock.Do(ctx, "PUT", testURL, []byte("put body"), nil)
	if err != nil {
		t.Errorf("Do (PUT) failed: %v", err)
	}

	// Verify all calls were tracked
	if mock.GetCallCount(testURL) != 3 {
		t.Errorf("Expected call count 3, got %d", mock.GetCallCount(testURL))
	}
}

func TestURLBuilder(t *testing.T) {
	// Test basic URL building
	builder := NewURLBuilder("https://api.example.com/query")
	url, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if url != "https://api.example.com/query" {
		t.Errorf("Expected 'https://api.example.com/query', got %s", url)
	}

	// Test with parameters
	builder = NewURLBuilder("https://api.example.com/query")
	url, err = builder.
		AddParam("function", "OVERVIEW").
		AddParam("symbol", "AAPL").
		SetParam("apikey", "demo").
		Build()

	if err != nil {
		t.Fatalf("Build with params failed: %v", err)
	}

	// Check that URL contains all parameters
	if !strings.Contains(url, "function=OVERVIEW") {
		t.Error("URL missing function parameter")
	}
	if !strings.Contains(url, "symbol=AAPL") {
		t.Error("URL missing symbol parameter")
	}
	if !strings.Contains(url, "apikey=demo") {
		t.Error("URL missing apikey parameter")
	}

	// Test invalid base URL
	builder = NewURLBuilder("invalid://url with spaces")
	_, err = builder.Build()
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestURLBuilder_Chaining(t *testing.T) {
	// Test method chaining
	builder := NewURLBuilder("https://api.example.com/query")

	// Add same parameter multiple times
	url, err := builder.
		AddParam("tag", "finance").
		AddParam("tag", "stocks").
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Should contain both tag values
	if !strings.Contains(url, "tag=finance") || !strings.Contains(url, "tag=stocks") {
		t.Errorf("URL should contain both tag values: %s", url)
	}

	// Test SetParam overwrites
	builder = NewURLBuilder("https://api.example.com/query")
	url, err = builder.
		AddParam("key", "old").
		SetParam("key", "new").
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if strings.Contains(url, "key=old") {
		t.Error("SetParam should have overwritten the old value")
	}
	if !strings.Contains(url, "key=new") {
		t.Error("URL should contain the new value")
	}
}

func TestClientStats_MockClient(t *testing.T) {
	ctx := context.Background()
	mock := NewMockClient()

	// Make some requests
	mock.Get(ctx, "https://example1.com", nil)
	mock.Get(ctx, "https://example2.com", nil)
	mock.Post(ctx, "https://example3.com", []byte("body"), nil)

	stats := mock.Stats()

	if stats.TotalRequests != 3 {
		t.Errorf("Expected TotalRequests to be 3, got %d", stats.TotalRequests)
	}

	if stats.SuccessfulRequests != 3 {
		t.Errorf("Expected SuccessfulRequests to be 3, got %d", stats.SuccessfulRequests)
	}

	if stats.FailedRequests != 0 {
		t.Errorf("Expected FailedRequests to be 0, got %d", stats.FailedRequests)
	}
}

func TestDependencyInjection_Interface(t *testing.T) {
	// Test that both implementations satisfy the interface
	var client1 HTTPClient = NewMockClient()
	var client2 HTTPClient = NewFastHTTPClient(DefaultConfig())

	// Test interface methods on both implementations
	clients := []HTTPClient{client1, client2}

	for i, client := range clients {
		t.Run(fmt.Sprintf("Client_%d", i), func(t *testing.T) {
			// Test that all interface methods exist and are callable
			if client == nil {
				t.Fatal("Client is nil")
			}

			// These should not panic
			stats := client.Stats()
			_ = stats

			err := client.Close()
			if err != nil {
				t.Errorf("Close() returned error: %v", err)
			}

			// Note: We don't test actual HTTP calls for FastHTTPClient here
			// since they would require a real server, but we test the interface compliance
		})
	}
}

func TestResponse_Structure(t *testing.T) {
	resp := &Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Server":       "test",
		},
		Body: []byte(`{"test": "data"}`),
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	if resp.Headers["Content-Type"] != "application/json" {
		t.Error("Headers not properly set")
	}

	if string(resp.Body) != `{"test": "data"}` {
		t.Error("Body not properly set")
	}
}

func TestConfig_Validation(t *testing.T) {
	// Test that config with zero values works
	config := &Config{}
	client := NewFastHTTPClient(config)

	if client == nil {
		t.Fatal("Client should be created even with empty config")
	}

	// Test that nil config uses defaults
	client = NewFastHTTPClient(nil)
	if client == nil {
		t.Fatal("Client should be created with nil config")
	}

	if client.config == nil {
		t.Fatal("Client config should not be nil when created with nil config")
	}
}

// Benchmark tests to verify performance characteristics
func BenchmarkMockClient_Get(b *testing.B) {
	ctx := context.Background()
	mock := NewMockClient()
	url := "https://benchmark.example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mock.Get(ctx, url, nil)
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
	}
}

func BenchmarkURLBuilder(b *testing.B) {
	baseURL := "https://api.example.com/query"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := NewURLBuilder(baseURL)
		_, err := builder.
			AddParam("function", "OVERVIEW").
			AddParam("symbol", "AAPL").
			AddParam("apikey", "demo").
			Build()
		if err != nil {
			b.Fatalf("URL build failed: %v", err)
		}
	}
}

// Example test showing how to use dependency injection in practice
func ExampleHTTPClient() {
	// Create a mock client for testing
	mockClient := NewMockClient()

	// Set up expected response
	expectedResponse := &Response{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       []byte(`{"symbol": "AAPL", "price": "150.00"}`),
	}
	mockClient.SetResponse("https://api.example.com/stock/AAPL", expectedResponse)

	// Function that depends on HTTPClient interface
	getStockPrice := func(client HTTPClient, symbol string) (string, error) {
		url := fmt.Sprintf("https://api.example.com/stock/%s", symbol)
		resp, err := client.Get(context.Background(), url, nil)
		if err != nil {
			return "", err
		}

		// In real code, you'd parse the JSON here
		return string(resp.Body), nil
	}

	// Use the function with mock client
	result, err := getStockPrice(mockClient, "AAPL")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", result)
	// Output: Response: {"symbol": "AAPL", "price": "150.00"}
}
