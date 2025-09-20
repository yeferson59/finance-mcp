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
