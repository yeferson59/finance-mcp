package config

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Config struct {
	APIURL         string              `json:"apiURL"`
	APIKey         string              `json:"apiKey"`
	Implementation *mcp.Implementation `json:"implementation"`
}

func NewConfig() *Config {
	env := NewEnv()
	_ = env.loadEnv()

	apiURL := env.GetEnv("API_URL", "https://www.alphavantage.co")
	apiKey := env.GetEnv("API_KEY", "demo")

	return &Config{
		APIURL: apiURL,
		APIKey: apiKey,
		Implementation: &mcp.Implementation{
			Title:   env.GetEnv("TITLE", "finance-mcp"),
			Name:    env.GetEnv("NAME", "Market-mcp"),
			Version: env.GetEnv("VERSION", "v1.0.0"),
		},
	}
}
