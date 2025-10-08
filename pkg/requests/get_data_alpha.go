package requests

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/yeferson59/finance-mcp/pkg/errors"
)

// Query represents a URL query parameter with name and value
type Query struct {
	name  string
	value string
}

// NewQuery creates a new Query with the given name and value
func NewQuery(name string, value string) Query {
	return Query{
		name:  name,
		value: value,
	}
}

// RequestAlpha represents a request to the Alpha Vantage API
type RequestAlpha struct {
	symbol  string
	apiKey  string
	baseURL string
	queries []Query
}

// NewAlpha creates a new Alpha Vantage request instance
func NewAlpha(baseURL string, apiKey string, symbol string, queries []Query) *RequestAlpha {
	return &RequestAlpha{
		apiKey:  apiKey,
		baseURL: baseURL,
		symbol:  symbol,
		queries: queries,
	}
}

// validate checks if all required fields are present
func (ra *RequestAlpha) validate() error {
	if ra.symbol == "" {
		return errors.ErrSymbolRequired
	}

	if ra.apiKey == "" {
		return errors.ErrAPIKeyRequired
	}

	if ra.baseURL == "" {
		return errors.ErrBaseURLRequired
	}

	return nil
}

// buildURL constructs the complete API URL with all parameters
func (ra *RequestAlpha) buildURL() (string, error) {
	ra.symbol = strings.ToUpper(strings.TrimSpace(ra.symbol))

	if err := ra.validate(); err != nil {
		return "", err
	}

	baseURL, err := url.Parse(ra.baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}

	for _, query := range ra.queries {
		value := query.value

		if query.name == "function" {
			value = strings.ToUpper(value)
		}

		params.Add(query.name, value)
	}

	params.Add("symbol", ra.symbol)
	params.Add("apikey", ra.apiKey)

	baseURL.RawQuery = params.Encode()

	return baseURL.String(), nil
}

// Get performs the HTTP GET request to Alpha Vantage API
func (ra *RequestAlpha) Get() ([]byte, error) {
	url, err := ra.buildURL()
	if err != nil {
		return nil, err
	}

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set("User-Agent", "Finance-MCP-Server/1.0")
	req.Header.Set("Accept", "application/json")

	if err := fasthttp.Do(req, res); err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	if res.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("%w: received status %d", errors.ErrUnexpectedStatusCode, res.StatusCode())
	}

	body := make([]byte, len(res.Body()))
	copy(body, res.Body())

	return body, nil
}
