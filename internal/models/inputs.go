// Package models defines the data structures used for input and output
// in the MCP financial data server.
//
// This package contains JSON-serializable structs with JSON Schema annotations
// that define the interface between MCP clients (like AI models) and the server.
// All structures include validation tags to ensure proper data types and formats.
package models

// SymbolInput represents the input parameters for stock-related MCP tools.
//
// This struct is used by MCP clients to specify which stock they want to query.
// It includes JSON Schema annotations for automatic validation and documentation
// generation in MCP protocol exchanges.
//
// Example usage in MCP tool calls:
//
//	{"symbol": "AAPL"}     - Apple Inc.
//	{"symbol": "GOOGL"}    - Alphabet Inc. (Google)
//	{"symbol": "MSFT"}     - Microsoft Corporation
//	{"symbol": "TSLA"}     - Tesla, Inc.
//
// The symbol should be a valid stock ticker symbol as recognized by
// major stock exchanges (NYSE, NASDAQ, etc.).
type SymbolInput struct {
	// Symbol is the stock ticker symbol for the company to query.
	//
	// This field accepts standard stock ticker symbols used in major exchanges:
	// - US stocks: "AAPL", "GOOGL", "MSFT", "AMZN", etc.
	// - Should be uppercase (will be converted automatically)
	// - Must be a valid symbol recognized by Alpha Vantage API
	//
	// JSON Schema validation ensures this field is provided and is a string.
	// The description helps AI models understand what kind of input is expected.
	Symbol string `json:"symbol" jsonschema:"the symbol of the stock to get"`
}

type IntradayPriceInput struct {
	Symbol        string  `json:"symbol" jsonschema:"the symbol of the stock to get"`
	Interval      string  `json:"interval" jsonschema:"the interval of the intraday price data e.g. '1min', '5min', '15min', '30min', '60min'"`
	Adjusted      *bool   `json:"adjusted" jsonschema:"By default, adjusted=true and the output time series is adjusted by historical split and dividend events. Set adjusted=false to query raw (as-traded) intraday values."`
	ExtendedHours *bool   `json:"extendedHours" jsonschema:"By default, extended_hours=true and the output time series will include both the regular trading hours and the extended (pre-market and post-market) trading hours (4:00am to 8:00pm Eastern Time for the US market). Set extended_hours=false to query regular trading hours (9:30am to 4:00pm US Eastern Time) only."`
	Month         *string `json:"month" jsonschema:"By default, this parameter is not set and the API will return intraday data for the most recent days of trading. You can use the month parameter (in YYYY-MM format) to query a specific month in history. For example, month=2009-01. Any month in the last 20+ years since 2000-01 (January 2000) is supported."`
	OutputSize    *string `json:"outputSize" jsonschema:"By default, output_size=compact and the API will return a compact set of data points. You can use the output_size parameter to query a full set of data points. For example, output_size=full. Any month in the last 20+ years since 2000-01 (January 2000) is supported."`
}
