// Package models defines the data structures used for input and output
// in the MCP financial data server.
package models

import (
	"time"
)

// OverviewOutput represents comprehensive stock and company information
// returned by the get-stock MCP tool.
//
// This struct contains all financial data fields typically provided by
// Alpha Vantage's OVERVIEW function, including company fundamentals,
// market metrics, and financial ratios.
//
// All fields are optional as different stocks may have different data availability.
// Monetary values are typically in USD unless otherwise specified.
// Field names match exactly with Alpha Vantage API response format.
type OverviewOutput struct {
	// Basic Company Information
	Symbol      string `json:"Symbol,omitempty"`      // Stock ticker symbol (e.g., "AAPL")
	Name        string `json:"Name,omitempty"`        // Company name (e.g., "Apple Inc")
	Description string `json:"Description,omitempty"` // Business description
	Country     string `json:"Country,omitempty"`     // Country of incorporation
	Sector      string `json:"Sector,omitempty"`      // Business sector (e.g., "Technology")
	Industry    string `json:"Industry,omitempty"`    // Industry classification
	Address     string `json:"Address,omitempty"`     // Company headquarters address
	Currency    string `json:"Currency,omitempty"`    // Trading currency (usually "USD")
	Exchange    string `json:"Exchange,omitempty"`    // Stock exchange (e.g., "NASDAQ")

	// Market Data
	MarketCapitalization       string `json:"MarketCapitalization,omitempty"`       // Total market value
	SharesOutstanding          string `json:"SharesOutstanding,omitempty"`          // Total shares outstanding
	BookValue                  string `json:"BookValue,omitempty"`                  // Book value per share
	DividendPerShare           string `json:"DividendPerShare,omitempty"`           // Annual dividend per share
	DividendYield              string `json:"DividendYield,omitempty"`              // Dividend yield percentage
	EPS                        string `json:"EPS,omitempty"`                        // Earnings per share (TTM)
	RevenuePerShareTTM         string `json:"RevenuePerShareTTM,omitempty"`         // Revenue per share (trailing 12 months)
	ProfitMargin               string `json:"ProfitMargin,omitempty"`               // Net profit margin
	OperatingMarginTTM         string `json:"OperatingMarginTTM,omitempty"`         // Operating margin (TTM)
	ReturnOnAssetsTTM          string `json:"ReturnOnAssetsTTM,omitempty"`          // Return on assets (TTM)
	ReturnOnEquityTTM          string `json:"ReturnOnEquityTTM,omitempty"`          // Return on equity (TTM)
	RevenueTTM                 string `json:"RevenueTTM,omitempty"`                 // Total revenue (TTM)
	GrossProfitTTM             string `json:"GrossProfitTTM,omitempty"`             // Gross profit (TTM)
	DilutedEPSTTM              string `json:"DilutedEPSTTM,omitempty"`              // Diluted earnings per share (TTM)
	QuarterlyEarningsGrowthYOY string `json:"QuarterlyEarningsGrowthYOY,omitempty"` // Quarterly earnings growth
	QuarterlyRevenueGrowthYOY  string `json:"QuarterlyRevenueGrowthYOY,omitempty"`  // Quarterly revenue growth

	// Financial Ratios
	PERatio              string `json:"PERatio,omitempty"`              // Price-to-earnings ratio
	PEGRatio             string `json:"PEGRatio,omitempty"`             // Price/earnings to growth ratio
	PriceToBookRatio     string `json:"PriceToBookRatio,omitempty"`     // Price-to-book ratio
	PriceToSalesRatioTTM string `json:"PriceToSalesRatioTTM,omitempty"` // Price-to-sales ratio (TTM)
	EVToRevenue          string `json:"EVToRevenue,omitempty"`          // Enterprise value to revenue
	EVToEBITDA           string `json:"EVToEBITDA,omitempty"`           // Enterprise value to EBITDA
	Beta                 string `json:"Beta,omitempty"`                 // Stock volatility measure
	ForwardPE            string `json:"ForwardPE,omitempty"`            // Forward price-to-earnings ratio
	AnalystTargetPrice   string `json:"AnalystTargetPrice,omitempty"`   // Analyst consensus target price

	// Trading Data
	Week52High          string `json:"52WeekHigh,omitempty"`          // 52-week high price
	Week52Low           string `json:"52WeekLow,omitempty"`           // 52-week low price
	Day50MovingAverage  string `json:"50DayMovingAverage,omitempty"`  // 50-day moving average
	Day200MovingAverage string `json:"200DayMovingAverage,omitempty"` // 200-day moving average
	DividendDate        string `json:"DividendDate,omitempty"`        // Last dividend payment date
	ExDividendDate      string `json:"ExDividendDate,omitempty"`      // Ex-dividend date

	// Additional Company Data
	FiscalYearEnd string `json:"FiscalYearEnd,omitempty"` // Fiscal year end month
	LatestQuarter string `json:"LatestQuarter,omitempty"` // Most recent quarter reported
	EBITDA        string `json:"EBITDA,omitempty"`        // Earnings before interest, taxes, depreciation, and amortization
	AssetType     string `json:"AssetType,omitempty"`     // Type of asset (usually "Common Stock")
	CIK           string `json:"CIK,omitempty"`           // Central Index Key (SEC identifier)
}

type OHLCVFloat struct {
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
}

type MetaData struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	Interval      string `json:"4. Interval"`
	OutputSize    string `json:"5. Output Size"`
	TimeZone      string `json:"6. Time Zone"`
}

type IntradayStockOutput struct {
	MetaData   MetaData     `json:"meta_data"`
	TimeSeries []OHLCVFloat `json:"time_series"`
}
