package tools

import (
	"context"
	"testing"

	"github.com/yeferson59/finance-mcp/internal/config"
	"github.com/yeferson59/finance-mcp/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestOverviewStock(t *testing.T) {
	cfg := config.NewConfig()
	overviewStock := NewOverviewStock(cfg.APIURL, cfg.APIKey)
	ctx := context.Background()
	input := models.SymbolInput{
		Symbol: "AAPL",
	}

	_, res, err := overviewStock.Get(ctx, nil, input)

	tx := assert.New(t)

	tx.NoError(err)
	tx.NotNil(res)
	tx.Equal(input.Symbol, res.Symbol)
}
