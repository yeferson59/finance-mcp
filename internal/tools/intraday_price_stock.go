package tools

import (
	"context"

	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yeferson59/finance-mcp/internal/models"
	"github.com/yeferson59/finance-mcp/pkg/parser"
)

type IntradayPriceStock struct {
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

func NewIntradayPriceStock(apiURL, apiKey string) *IntradayPriceStock {
	return &IntradayPriceStock{
		APIURL: apiURL,
		APIKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *IntradayPriceStock) Get(ctx context.Context, req *mcp.CallToolRequest, input models.IntradayPriceInput) (*mcp.CallToolResult, models.IntradayStockOutput, error) {
	baseURL := fmt.Sprintf("%s/query?function=TIME_SERIES_INTRADAY&symbol=%s&interval=%s&apikey=%s", s.APIURL, strings.ToUpper(input.Symbol), input.Interval, s.APIKey)

	if input.Adjusted != nil {
		baseURL += fmt.Sprintf("&adjusted=%t", *input.Adjusted)
	}
	if input.ExtendedHours != nil {
		baseURL += fmt.Sprintf("&extended_hours=%t", *input.ExtendedHours)
	}
	if input.Month != nil {
		baseURL += fmt.Sprintf("&month=%s", *input.Month)
	}
	if input.OutputSize != nil {
		baseURL += fmt.Sprintf("&outputsize=%s", *input.OutputSize)
	}

	res, err := s.httpClient.Get(baseURL)

	if err != nil {
		return nil, models.IntradayStockOutput{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, models.IntradayStockOutput{}, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, models.IntradayStockOutput{}, err
	}

	rawData, err := parser.IntradayPrices(bodyBytes)
	if err != nil {
		return nil, models.IntradayStockOutput{}, err
	}

	data, err := rawData.ProcessTimeSeries()
	if err != nil {
		return nil, models.IntradayStockOutput{}, err
	}

	return nil, *data, nil
}
