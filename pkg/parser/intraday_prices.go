package parser

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"

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
	err := sonic.Unmarshal(jsonData, &rawResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON into raw map: %w", err)
	}

	// Store raw data for later processing
	response.rawData = rawResponse

	// Unmarshal into the structured response for MetaData
	err = sonic.Unmarshal(jsonData, &response)
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

	// For small to medium datasets (< 1000 entries), sequential processing is faster
	// than goroutine overhead. For larger datasets, we use a worker pool.
	if len(r.TimeSeries) < 1000 {
		// Sequential processing for better performance on small datasets
		for timestampStr, ohlcv := range r.TimeSeries {
			processedEntry, err := r.processEntry(timestampStr, ohlcv)
			if err != nil {
				return nil, err
			}
			processed.TimeSeries = append(processed.TimeSeries, processedEntry)
		}
	} else {
		// Use worker pool for large datasets to limit goroutine count
		const numWorkers = 8
		type job struct {
			timestamp string
			ohlcv     OHLCV
		}

		jobs := make(chan job, len(r.TimeSeries))
		results := make(chan models.OHLCVFloat, len(r.TimeSeries))
		errChan := make(chan error, 1)
		var wg sync.WaitGroup

		// Start workers
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := range jobs {
					processedEntry, err := r.processEntry(j.timestamp, j.ohlcv)
					if err != nil {
						select {
						case errChan <- err:
						default:
						}
						return
					}
					results <- processedEntry
				}
			}()
		}

		// Send jobs
		go func() {
			for timestampStr, ohlcv := range r.TimeSeries {
				jobs <- job{timestampStr, ohlcv}
			}
			close(jobs)
		}()

		// Wait and close results
		wg.Wait()
		close(results)
		close(errChan)

		// Check for errors
		if len(errChan) > 0 {
			return nil, <-errChan
		}

		// Collect results
		for v := range results {
			processed.TimeSeries = append(processed.TimeSeries, v)
		}
	}

	// Sort by timestamp
	sort.Slice(processed.TimeSeries, func(i, j int) bool {
		return processed.TimeSeries[i].Timestamp.Before(processed.TimeSeries[j].Timestamp)
	})

	return processed, nil
}

// processEntry processes a single time series entry
func (r *AlphaVantageResponse) processEntry(timestampStr string, ohlcv OHLCV) (models.OHLCVFloat, error) {
	timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		return models.OHLCVFloat{}, fmt.Errorf("error parsing timestamp %s: %w", timestampStr, err)
	}

	open, err := strconv.ParseFloat(ohlcv.Open, 64)
	if err != nil {
		return models.OHLCVFloat{}, fmt.Errorf("error parsing open price for %s: %w", timestampStr, err)
	}

	high, err := strconv.ParseFloat(ohlcv.High, 64)
	if err != nil {
		return models.OHLCVFloat{}, fmt.Errorf("error parsing high price for %s: %w", timestampStr, err)
	}

	low, err := strconv.ParseFloat(ohlcv.Low, 64)
	if err != nil {
		return models.OHLCVFloat{}, fmt.Errorf("error parsing low price for %s: %w", timestampStr, err)
	}

	closePrice, err := strconv.ParseFloat(ohlcv.Close, 64)
	if err != nil {
		return models.OHLCVFloat{}, fmt.Errorf("error parsing close price for %s: %w", timestampStr, err)
	}

	volume, err := strconv.ParseInt(ohlcv.Volume, 10, 64)
	if err != nil {
		return models.OHLCVFloat{}, fmt.Errorf("error parsing volume for %s: %w", timestampStr, err)
	}

	return models.OHLCVFloat{
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     closePrice,
		Volume:    volume,
	}, nil
}
