package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yeferson59/finance-mcp/internal/config"
	"github.com/yeferson59/finance-mcp/internal/models"
)

func TestIntradayPrice(t *testing.T) {
	cfg := config.NewConfig()
	intradayPrice := NewIntradayPriceStock(cfg.APIURL, cfg.APIKey)
	input := models.IntradayPriceInput{Symbol: "AAPL", Interval: "60min"}

	_, res, err := intradayPrice.Get(context.Background(), nil, input)

	tx := assert.New(t)

	tx.NoError(err)
	tx.NotEmpty(res)
	tx.Equal(input.Symbol, res.MetaData.Symbol)
	tx.Equal(input.Interval, res.MetaData.Interval)
	tx.NotEmpty(res.TimeSeries)
}
