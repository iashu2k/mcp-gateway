package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/iashu2k/mcp-gateway/backend/internal/config"
	"github.com/iashu2k/mcp-gateway/backend/internal/httpapi"
	"github.com/iashu2k/mcp-gateway/backend/internal/platform/database"
)

func main() {
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}

	db, err := database.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           httpapi.NewRouter(logger, db),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting MCP Gateway API",
			"environment", cfg.AppEnv,
			"port", cfg.HTTPPort,
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignal

	logger.Info("shutting down MCP Gateway API")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}
