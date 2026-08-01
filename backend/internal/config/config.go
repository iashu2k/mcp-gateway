package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv      string
	HTTPPort    int
	DatabaseURL string
}

func Load() (Config, error) {
	port, err := strconv.Atoi(getEnv("HTTP_PORT", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("parse HTTP_PORT: %w", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		HTTPPort:    port,
		DatabaseURL: databaseURL,
	}, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
