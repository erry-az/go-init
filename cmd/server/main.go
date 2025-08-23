package main

import (
	"log/slog"
	"os"

	"github.com/erry-az/go-init/config"
	"github.com/erry-az/go-init/internal/app"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.New()
	if err != nil {
		slog.Error("Error loading config:", slog.Any("error", err))
		return
	}

	// Create and initialize application
	application, err := app.NewEndpoint(cfg)
	if err != nil {
		slog.Error("Failed to initialize application", slog.Any("error", err))
		os.Exit(1)
	}
	defer application.Close()

	// Start application servers and wait for shutdown signal
	if err := application.Start(); err != nil {
		slog.Error("Application startup failed", slog.Any("error", err))
		os.Exit(1)
	}
}
