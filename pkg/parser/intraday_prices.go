package parser

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yeferson59/finance-mcp/internal/models"
)

type OHLCV struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
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

type AlphaVantageResponse struct {
	MetaData   MetaData         `json:"Meta Data"`
	TimeSeries map[string]OHLCV `json:"-"`
	rawData    map[string]any
}

func IntradayPrices(jsonData []byte) (*AlphaVantageResponse, error) {
	var response AlphaVantageResponse
	var rawResponse map[string]any

	// First, unmarshal into a generic map to handle dynamic keys
	err := json.Unmarshal(jsonData, &rawResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON into raw map: %w", err)
	}

	// Store raw data for later processing
	response.rawData = rawResponse

	// Unmarshal into the structured response for MetaData
	err = json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON into structured response: %w", err)
	}

	// Check for API error messages
	if errorMsg, exists := rawResponse["Error Message"]; exists {
		return nil, fmt.Errorf("API error: %v", errorMsg)
	}

	if note, exists := rawResponse["Note"]; exists {
		return nil, fmt.Errorf("API note (likely rate limit): %v", note)
	}

	if info, exists := rawResponse["Information"]; exists {
		if infoStr, ok := info.(string); ok {
			if strings.Contains(strings.ToLower(infoStr), "rate limit") || strings.Contains(strings.ToLower(infoStr), "premium") {
				return nil, fmt.Errorf("API rate limit reached: %v", info)
			}
			return nil, fmt.Errorf("API information: %v", info)
		}
	}

	// Find and extract the time series data
	err = response.extractTimeSeries()
	if err != nil {
		return nil, fmt.Errorf("error extracting time series: %w", err)
	}

	return &response, nil
}

// extractTimeSeries finds the time series data in the raw response
// The key format is "Time Series (interval)" where interval can be 1min, 5min, etc.
func (r *AlphaVantageResponse) extractTimeSeries() error {
	if r.rawData == nil {
		return fmt.Errorf("no raw data available")
	}

	var timeSeriesKey string
	var timeSeriesData any

	// Look for time series key in the raw data
	for key, value := range r.rawData {
		if strings.Contains(strings.ToLower(key), "time series") {
			timeSeriesKey = key
			timeSeriesData = value
			break
		}
	}

	if timeSeriesKey == "" {
		return fmt.Errorf("no time series data found in response")
	}

	// Convert the time series data to our expected format
	timeSeriesMap, ok := timeSeriesData.(map[string]any)
	if !ok {
		return fmt.Errorf("time series data is not in expected format")
	}

	r.TimeSeries = make(map[string]OHLCV)

	for timestamp, ohlcvData := range timeSeriesMap {
		ohlcvMap, ok := ohlcvData.(map[string]any)
		if !ok {
			continue // Skip invalid entries
		}

		ohlcv := OHLCV{}

		// Extract OHLCV values safely
		if open, exists := ohlcvMap["1. open"]; exists {
			if openStr, ok := open.(string); ok {
				ohlcv.Open = openStr
			}
		}

		if high, exists := ohlcvMap["2. high"]; exists {
			if highStr, ok := high.(string); ok {
				ohlcv.High = highStr
			}
		}

		if low, exists := ohlcvMap["3. low"]; exists {
			if lowStr, ok := low.(string); ok {
				ohlcv.Low = lowStr
			}
		}

		if close, exists := ohlcvMap["4. close"]; exists {
			if closeStr, ok := close.(string); ok {
				ohlcv.Close = closeStr
			}
		}

		if volume, exists := ohlcvMap["5. volume"]; exists {
			if volumeStr, ok := volume.(string); ok {
				ohlcv.Volume = volumeStr
			}
		}

		r.TimeSeries[timestamp] = ohlcv
	}

	return nil
}

func (r *AlphaVantageResponse) ProcessTimeSeries() (*models.IntradayStockOutput, error) {
	if r.TimeSeries == nil {
		return &models.IntradayStockOutput{
			MetaData:   models.MetaData(r.MetaData),
			TimeSeries: []models.OHLCVFloat{},
		}, nil
	}

	processed := &models.IntradayStockOutput{
		MetaData:   models.MetaData(r.MetaData),
		TimeSeries: make([]models.OHLCVFloat, 0, len(r.TimeSeries)),
	}

	results := make(chan models.OHLCVFloat, len(r.TimeSeries))
	errChan := make(chan error, len(r.TimeSeries))
	var wg sync.WaitGroup

	for timestampStr, ohlcv := range r.TimeSeries {
		wg.Add(1)
		go func(ts string, o OHLCV) {
			defer wg.Done()

			timestamp, err := time.Parse("2006-01-02 15:04:05", ts)
			if err != nil {
				errChan <- fmt.Errorf("error parsing timestamp %s: %w", ts, err)
				return
			}

			open, err := strconv.ParseFloat(o.Open, 64)
			if err != nil {
				errChan <- fmt.Errorf("error parsing open price for %s: %w", ts, err)
				return
			}

			high, err := strconv.ParseFloat(o.High, 64)
			if err != nil {
				errChan <- fmt.Errorf("error parsing high price for %s: %w", ts, err)
				return
			}

			low, err := strconv.ParseFloat(o.Low, 64)
			if err != nil {
				errChan <- fmt.Errorf("error parsing low price for %s: %w", ts, err)
				return
			}

			closePrice, err := strconv.ParseFloat(o.Close, 64)
			if err != nil {
				errChan <- fmt.Errorf("error parsing close price for %s: %w", ts, err)
				return
			}

			volume, err := strconv.ParseInt(o.Volume, 10, 64)
			if err != nil {
				errChan <- fmt.Errorf("error parsing volume for %s: %w", ts, err)
				return
			}

			processedOHLCV := OHLCVFloat{
				Timestamp: timestamp,
				Open:      open,
				High:      high,
				Low:       low,
				Close:     closePrice,
				Volume:    volume,
			}

			results <- models.OHLCVFloat(processedOHLCV)
		}(timestampStr, ohlcv)
	}

	wg.Wait()
	close(results)
	close(errChan)

	// Check for any errors
	if len(errChan) > 0 {
		return nil, <-errChan
	}

	// Collect all results
	for v := range results {
		processed.TimeSeries = append(processed.TimeSeries, v)
	}

	// Sort by timestamp
	sort.Slice(processed.TimeSeries, func(i, j int) bool {
		return processed.TimeSeries[i].Timestamp.Before(processed.TimeSeries[j].Timestamp)
	})

	return processed, nil
}
