package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntradayPrices_Success(t *testing.T) {
	mockResponse := `{
		"Meta Data": {
			"1. Information": "Intraday (5min) open, high, low, close prices and volume",
			"2. Symbol": "AAPL",
			"3. Last Refreshed": "2024-01-15 20:00:00",
			"4. Interval": "5min",
			"5. Output Size": "Compact",
			"6. Time Zone": "US/Eastern"
		},
		"Time Series (5min)": {
			"2024-01-15 20:00:00": {
				"1. open": "185.50",
				"2. high": "185.75",
				"3. low": "185.25",
				"4. close": "185.60",
				"5. volume": "125000"
			},
			"2024-01-15 19:55:00": {
				"1. open": "185.20",
				"2. high": "185.55",
				"3. low": "185.15",
				"4. close": "185.50",
				"5. volume": "98000"
			},
			"2024-01-15 19:50:00": {
				"1. open": "184.80",
				"2. high": "185.25",
				"3. low": "184.75",
				"4. close": "185.20",
				"5. volume": "87500"
			}
		}
	}`

	response, err := IntradayPrices([]byte(mockResponse))
	require.NoError(t, err)
	require.NotNil(t, response)

	// Test metadata
	assert.Equal(t, "AAPL", response.MetaData.Symbol)
	assert.Equal(t, "5min", response.MetaData.Interval)
	assert.Equal(t, "2024-01-15 20:00:00", response.MetaData.LastRefreshed)
	assert.Equal(t, "Compact", response.MetaData.OutputSize)

	// Test time series data extraction
	assert.NotNil(t, response.TimeSeries)
	assert.Len(t, response.TimeSeries, 3)

	// Test specific data point
	ohlcv, exists := response.TimeSeries["2024-01-15 20:00:00"]
	assert.True(t, exists)
	assert.Equal(t, "185.50", ohlcv.Open)
	assert.Equal(t, "185.75", ohlcv.High)
	assert.Equal(t, "185.25", ohlcv.Low)
	assert.Equal(t, "185.60", ohlcv.Close)
	assert.Equal(t, "125000", ohlcv.Volume)
}

func TestIntradayPrices_ProcessTimeSeries(t *testing.T) {
	mockResponse := `{
		"Meta Data": {
			"1. Information": "Intraday (1min) open, high, low, close prices and volume",
			"2. Symbol": "MSFT",
			"3. Last Refreshed": "2024-01-15 16:00:00",
			"4. Interval": "1min",
			"5. Output Size": "Compact",
			"6. Time Zone": "US/Eastern"
		},
		"Time Series (1min)": {
			"2024-01-15 16:00:00": {
				"1. open": "380.50",
				"2. high": "380.75",
				"3. low": "380.25",
				"4. close": "380.60",
				"5. volume": "75000"
			},
			"2024-01-15 15:59:00": {
				"1. open": "380.20",
				"2. high": "380.55",
				"3. low": "380.15",
				"4. close": "380.50",
				"5. volume": "68000"
			}
		}
	}`

	response, err := IntradayPrices([]byte(mockResponse))
	require.NoError(t, err)

	processed, err := response.ProcessTimeSeries()
	require.NoError(t, err)
	require.NotNil(t, processed)

	// Test processed metadata
	assert.Equal(t, "MSFT", processed.MetaData.Symbol)
	assert.Equal(t, "1min", processed.MetaData.Interval)

	// Test processed time series
	assert.Len(t, processed.TimeSeries, 2)

	// Test data conversion and sorting (should be sorted by timestamp)
	assert.True(t, processed.TimeSeries[0].Timestamp.Before(processed.TimeSeries[1].Timestamp))

	// Test first data point (earlier timestamp)
	firstPoint := processed.TimeSeries[0]
	expectedTime, _ := time.Parse("2006-01-02 15:04:05", "2024-01-15 15:59:00")
	assert.Equal(t, expectedTime, firstPoint.Timestamp)
	assert.Equal(t, 380.20, firstPoint.Open)
	assert.Equal(t, 380.55, firstPoint.High)
	assert.Equal(t, 380.15, firstPoint.Low)
	assert.Equal(t, 380.50, firstPoint.Close)
	assert.Equal(t, int64(68000), firstPoint.Volume)

	// Test second data point (later timestamp)
	secondPoint := processed.TimeSeries[1]
	expectedTime2, _ := time.Parse("2006-01-02 15:04:05", "2024-01-15 16:00:00")
	assert.Equal(t, expectedTime2, secondPoint.Timestamp)
	assert.Equal(t, 380.50, secondPoint.Open)
	assert.Equal(t, 380.75, secondPoint.High)
	assert.Equal(t, 380.25, secondPoint.Low)
	assert.Equal(t, 380.60, secondPoint.Close)
	assert.Equal(t, int64(75000), secondPoint.Volume)
}

func TestIntradayPrices_APIError(t *testing.T) {
	mockErrorResponse := `{
		"Error Message": "Invalid API call. Please retry or visit the documentation (https://www.alphavantage.co/documentation/) for TIME_SERIES_INTRADAY."
	}`

	_, err := IntradayPrices([]byte(mockErrorResponse))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
	assert.Contains(t, err.Error(), "Invalid API call")
}

func TestIntradayPrices_RateLimitNote(t *testing.T) {
	mockRateLimitResponse := `{
		"Note": "Thank you for using Alpha Vantage! Our standard API call frequency is 5 calls per minute and 100 calls per day. Please subscribe to any of the premium plans at https://www.alphavantage.co/premium/ to instantly remove all daily rate limits."
	}`

	_, err := IntradayPrices([]byte(mockRateLimitResponse))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API note")
	assert.Contains(t, err.Error(), "rate limit")
}

func TestIntradayPrices_InformationRateLimit(t *testing.T) {
	mockInfoResponse := `{
		"Information": "We have detected your API key and our standard API rate limit is 25 requests per day. Please subscribe to any of the premium plans at https://www.alphavantage.co/premium/ to instantly remove all daily rate limits."
	}`

	_, err := IntradayPrices([]byte(mockInfoResponse))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API rate limit reached")
	assert.Contains(t, err.Error(), "25 requests per day")
}

func TestIntradayPrices_InformationGeneral(t *testing.T) {
	mockInfoResponse := `{
		"Information": "This is some general information about the API."
	}`

	_, err := IntradayPrices([]byte(mockInfoResponse))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API information")
}

func TestIntradayPrices_DifferentIntervals(t *testing.T) {
	testCases := []struct {
		name     string
		interval string
		tsKey    string
	}{
		{"1min", "1min", "Time Series (1min)"},
		{"5min", "5min", "Time Series (5min)"},
		{"15min", "15min", "Time Series (15min)"},
		{"30min", "30min", "Time Series (30min)"},
		{"60min", "60min", "Time Series (60min)"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockResponse := `{
				"Meta Data": {
					"1. Information": "Intraday (` + tc.interval + `) open, high, low, close prices and volume",
					"2. Symbol": "GOOGL",
					"3. Last Refreshed": "2024-01-15 16:00:00",
					"4. Interval": "` + tc.interval + `",
					"5. Output Size": "Compact",
					"6. Time Zone": "US/Eastern"
				},
				"` + tc.tsKey + `": {
					"2024-01-15 16:00:00": {
						"1. open": "150.50",
						"2. high": "150.75",
						"3. low": "150.25",
						"4. close": "150.60",
						"5. volume": "45000"
					}
				}
			}`

			response, err := IntradayPrices([]byte(mockResponse))
			require.NoError(t, err)
			require.NotNil(t, response)

			assert.Equal(t, "GOOGL", response.MetaData.Symbol)
			assert.Equal(t, tc.interval, response.MetaData.Interval)
			assert.Len(t, response.TimeSeries, 1)
		})
	}
}

func TestIntradayPrices_EmptyTimeSeries(t *testing.T) {
	mockResponse := `{
		"Meta Data": {
			"1. Information": "Intraday (5min) open, high, low, close prices and volume",
			"2. Symbol": "AAPL",
			"3. Last Refreshed": "2024-01-15 20:00:00",
			"4. Interval": "5min",
			"5. Output Size": "Compact",
			"6. Time Zone": "US/Eastern"
		},
		"Time Series (5min)": {}
	}`

	response, err := IntradayPrices([]byte(mockResponse))
	require.NoError(t, err)

	processed, err := response.ProcessTimeSeries()
	require.NoError(t, err)
	assert.Empty(t, processed.TimeSeries)
	assert.Equal(t, "AAPL", processed.MetaData.Symbol)
}

func TestIntradayPrices_NoTimeSeries(t *testing.T) {
	mockResponse := `{
		"Meta Data": {
			"1. Information": "Intraday (5min) open, high, low, close prices and volume",
			"2. Symbol": "AAPL",
			"3. Last Refreshed": "2024-01-15 20:00:00",
			"4. Interval": "5min",
			"5. Output Size": "Compact",
			"6. Time Zone": "US/Eastern"
		}
	}`

	_, err := IntradayPrices([]byte(mockResponse))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no time series data found")
}

func TestIntradayPrices_InvalidJSON(t *testing.T) {
	invalidJSON := `{"invalid": json"}`

	_, err := IntradayPrices([]byte(invalidJSON))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing JSON")
}

func TestIntradayPrices_MalformedTimeSeriesData(t *testing.T) {
	mockResponse := `{
		"Meta Data": {
			"1. Information": "Intraday (5min) open, high, low, close prices and volume",
			"2. Symbol": "AAPL",
			"3. Last Refreshed": "2024-01-15 20:00:00",
			"4. Interval": "5min",
			"5. Output Size": "Compact",
			"6. Time Zone": "US/Eastern"
		},
		"Time Series (5min)": {
			"2024-01-15 20:00:00": {
				"1. open": "not-a-number",
				"2. high": "185.75",
				"3. low": "185.25",
				"4. close": "185.60",
				"5. volume": "125000"
			}
		}
	}`

	response, err := IntradayPrices([]byte(mockResponse))
	require.NoError(t, err)

	_, err = response.ProcessTimeSeries()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing")
}

func TestIntradayPrices_InvalidTimestamp(t *testing.T) {
	mockResponse := `{
		"Meta Data": {
			"1. Information": "Intraday (5min) open, high, low, close prices and volume",
			"2. Symbol": "AAPL",
			"3. Last Refreshed": "2024-01-15 20:00:00",
			"4. Interval": "5min",
			"5. Output Size": "Compact",
			"6. Time Zone": "US/Eastern"
		},
		"Time Series (5min)": {
			"invalid-timestamp": {
				"1. open": "185.50",
				"2. high": "185.75",
				"3. low": "185.25",
				"4. close": "185.60",
				"5. volume": "125000"
			}
		}
	}`

	response, err := IntradayPrices([]byte(mockResponse))
	require.NoError(t, err)

	_, err = response.ProcessTimeSeries()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing timestamp")
}

func TestProcessTimeSeries_SortingOrder(t *testing.T) {
	mockResponse := `{
		"Meta Data": {
			"1. Information": "Intraday (5min) open, high, low, close prices and volume",
			"2. Symbol": "AAPL",
			"3. Last Refreshed": "2024-01-15 20:00:00",
			"4. Interval": "5min",
			"5. Output Size": "Compact",
			"6. Time Zone": "US/Eastern"
		},
		"Time Series (5min)": {
			"2024-01-15 20:00:00": {
				"1. open": "185.50",
				"2. high": "185.75",
				"3. low": "185.25",
				"4. close": "185.60",
				"5. volume": "125000"
			},
			"2024-01-15 19:50:00": {
				"1. open": "184.80",
				"2. high": "185.25",
				"3. low": "184.75",
				"4. close": "185.20",
				"5. volume": "87500"
			},
			"2024-01-15 19:55:00": {
				"1. open": "185.20",
				"2. high": "185.55",
				"3. low": "185.15",
				"4. close": "185.50",
				"5. volume": "98000"
			}
		}
	}`

	response, err := IntradayPrices([]byte(mockResponse))
	require.NoError(t, err)

	processed, err := response.ProcessTimeSeries()
	require.NoError(t, err)

	// Verify data is sorted chronologically (earliest first)
	assert.Len(t, processed.TimeSeries, 3)

	// Should be sorted: 19:50, 19:55, 20:00
	assert.True(t, processed.TimeSeries[0].Timestamp.Before(processed.TimeSeries[1].Timestamp))
	assert.True(t, processed.TimeSeries[1].Timestamp.Before(processed.TimeSeries[2].Timestamp))

	// Verify the specific order
	expectedTimes := []string{
		"2024-01-15 19:50:00",
		"2024-01-15 19:55:00",
		"2024-01-15 20:00:00",
	}

	for i, expectedTime := range expectedTimes {
		expected, _ := time.Parse("2006-01-02 15:04:05", expectedTime)
		assert.Equal(t, expected, processed.TimeSeries[i].Timestamp)
	}
}
