package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv        string
	HTTPPort      int
	DatabaseURL   string
	JWTSecret     string
	JWTIssuer     string
	JWTTTLMinutes int
}

func Load() (Config, error) {
	port, err := strconv.Atoi(getEnv("HTTP_PORT", "8080"))
	if err != nil {
		return Config{}, fmt.Errorf("parse HTTP_PORT: %w", err)
	}

	jwtTTLMinutes, err := strconv.Atoi(getEnv("JWT_TTL_MINUTES", "60"))
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_TTL_MINUTES: %w", err)
	}

	if jwtTTLMinutes <= 0 {
		return Config{}, fmt.Errorf("JWT_TTL_MINUTES must be greater than zero")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	if len(jwtSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	return Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		HTTPPort:      port,
		DatabaseURL:   databaseURL,
		JWTSecret:     jwtSecret,
		JWTIssuer:     getEnv("JWT_ISSUER", "mcp-gateway"),
		JWTTTLMinutes: jwtTTLMinutes,
	}, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
