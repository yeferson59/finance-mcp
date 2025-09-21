package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/yeferson59/finance-mcp/pkg/file"
)

type Env struct{}

func NewEnv() *Env {
	return &Env{}
}

func (Env) loadEnv() error {
	envPath := file.GetPathFile(".env")
	err := godotenv.Load(envPath)

	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	return nil
}

func (e *Env) GetEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	log.Println("[ENV] Environment variable not found:", key)
	return defaultValue
}
