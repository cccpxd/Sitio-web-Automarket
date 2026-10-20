package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"automarket/internal/app"
	"automarket/internal/config"
	"automarket/internal/platform/logging"
	"automarket/internal/platform/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.New(cfg)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer func() {
		if err := application.Close(); err != nil {
			logger.Error("database_close_failed", "error", err)
		}
	}()

	logger.Info("application_started", "port", cfg.Port)
	return server.Run(ctx, application.Handler, ":"+cfg.Port, logger)
}
