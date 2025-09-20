package tools

import (
	"context"
	"fmt"
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

	assert.NoError(t, err)
	assert.NotNil(t, res)

	fmt.Println(res)
}
