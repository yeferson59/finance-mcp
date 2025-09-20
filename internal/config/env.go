package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

type Env struct{}

func NewEnv() *Env {
	return &Env{}
}

func (Env) loadEnv() error {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	envPath := filepath.Join(projectRoot, ".env")
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
