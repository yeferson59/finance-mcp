package requests

import (
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/yeferson59/finance-mcp/pkg/errors"
)

type Query struct {
	name  string
	value string
}

func NewQuery(name string, value string) Query {
	return Query{
		name:  name,
		value: value,
	}
}

type RequestAlpha struct {
	symbol  string
	apiKey  string
	baseURL string
	querys  []Query
}

func NewAlpha(baseURL string, apiKey string, symbol string, querys []Query) *RequestAlpha {
	return &RequestAlpha{
		apiKey:  apiKey,
		baseURL: baseURL,
		symbol:  symbol,
		querys:  querys,
	}
}

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

func (ra *RequestAlpha) buildURL() (string, error) {
	ra.symbol = strings.ToUpper(strings.TrimSpace(ra.symbol))

	if err := ra.validate(); err != nil {
		return "", err
	}

	var url string = ra.baseURL

	for _, query := range ra.querys {
		if query.name == "function" {
			url += fmt.Sprintf("%s=%s&", query.name, strings.ToUpper(query.value))
		} else {
			url += fmt.Sprintf("%s=%s&", query.name, query.value)
		}
	}

	url += fmt.Sprintf("symbol=%s", ra.symbol)
	url += fmt.Sprintf("&apikey=%s", ra.apiKey)
	return url, nil
}

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
	req.Header.Set("User-Agent", "Simple-MCP-Server/1.0")
	req.Header.Set("Accept", "application/json")

	if err := fasthttp.Do(req, res); err != nil {
		return nil, err
	}

	if res.StatusCode() != fasthttp.StatusOK {
		return nil, errors.ErrUnexpectedStatusCode
	}

	return res.Body(), nil
}
