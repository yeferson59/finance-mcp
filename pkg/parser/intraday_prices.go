package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	MetaData   MetaData                   `json:"Meta Data"`
	TimeSeries map[string]OHLCV           `json:"-"`
	RawData    map[string]json.RawMessage `json:"-"`
}

func (r *AlphaVantageResponse) UnmarshalJSON(data []byte) error {
	var temp map[string]json.RawMessage
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if metaDataRaw, exists := temp["Meta Data"]; exists {
		if err := json.Unmarshal(metaDataRaw, &r.MetaData); err != nil {
			return fmt.Errorf("error parsing Meta Data: %w", err)
		}
	}

	var timeSeriesKey string
	possibleKeys := []string{
		"Time Series (1min)",
		"Time Series (5min)",
		"Time Series (15min)",
		"Time Series (30min)",
		"Time Series (60min)",
	}

	for _, key := range possibleKeys {
		if _, exists := temp[key]; exists {
			timeSeriesKey = key
			break
		}
	}

	if timeSeriesKey == "" {
		return fmt.Errorf("no time series data found")
	}

	r.TimeSeries = make(map[string]OHLCV)
	if err := json.Unmarshal(temp[timeSeriesKey], &r.TimeSeries); err != nil {
		return fmt.Errorf("error parsing time series: %w", err)
	}

	r.RawData = temp

	return nil
}

func ParseResponse(jsonData string) (*AlphaVantageResponse, error) {
	var response AlphaVantageResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}
	return &response, nil
}

func (r *AlphaVantageResponse) GetTimeSeriesInterval() string {
	return r.MetaData.Interval
}

func (r *AlphaVantageResponse) GetTimeSeriesKey() string {
	possibleKeys := []string{
		"Time Series (1min)",
		"Time Series (5min)",
		"Time Series (15min)",
		"Time Series (30min)",
		"Time Series (60min)",
	}

	for _, key := range possibleKeys {
		if _, exists := r.RawData[key]; exists {
			return key
		}
	}
	return ""
}

func (r *AlphaVantageResponse) ProcessTimeSeries() (*models.IntradayStockOutput, error) {
	processed := &models.IntradayStockOutput{
		MetaData:   models.MetaData(r.MetaData),
		TimeSeries: make([]models.OHLCVFloat, 0, len(r.TimeSeries)),
	}

	for timestampStr, ohlcv := range r.TimeSeries {
		timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing timestamp %s: %w", timestampStr, err)
		}

		open, err := strconv.ParseFloat(ohlcv.Open, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing open price: %w", err)
		}

		high, err := strconv.ParseFloat(ohlcv.High, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing high price: %w", err)
		}

		low, err := strconv.ParseFloat(ohlcv.Low, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing low price: %w", err)
		}

		close, err := strconv.ParseFloat(ohlcv.Close, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing close price: %w", err)
		}

		volume, err := strconv.ParseInt(ohlcv.Volume, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing volume: %w", err)
		}

		processedOHLCV := OHLCVFloat{
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		}

		processed.TimeSeries = append(processed.TimeSeries, models.OHLCVFloat(processedOHLCV))
	}

	return processed, nil
}

func (r *AlphaVantageResponse) GetLastRefreshedTime() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", r.MetaData.LastRefreshed)
}
