package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/iashu2k/mcp-gateway/backend/internal/config"
	"github.com/iashu2k/mcp-gateway/backend/internal/httpapi"
	"github.com/iashu2k/mcp-gateway/backend/internal/observability"
	"github.com/iashu2k/mcp-gateway/backend/internal/platform/database"
)

func main() {
	// Load .env file if present (ignore error in production)
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("starting MCP Gateway API",
		"environment", cfg.AppEnv,
		"port", cfg.HTTPPort,
	)

	// Initialize OpenTelemetry (optional for local dev)
	tp, err := observability.InitTracer("mcp-gateway")
	if err != nil {
		logger.Warn("failed to initialize tracer", "error", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tp.Shutdown(ctx); err != nil {
				logger.Error("failed to shutdown tracer", "error", err)
			}
		}()
		logger.Info("OpenTelemetry tracing initialized")
	}

	// Connect to database
	db, err := database.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection established")

	// Update database connection metrics periodically
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			stats := db.Stat()
			observability.DatabaseConnectionsOpen.Set(float64(stats.TotalConns()))
		}
	}()

	// Create HTTP router
	router := httpapi.NewRouter(logger, db, cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("HTTP server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}
