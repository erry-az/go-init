package main

import (
	"context"
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

	// Create consumer application
	consumerApp, err := app.NewConsumerApp(cfg)
	if err != nil {
		slog.Error("Error creating consumer app:", slog.Any("error", err))
		return
	}

	err = consumerApp.Run(context.Background())
	if err != nil {
		slog.Error("Error loading consumer:", slog.Any("error", err))
	}
}
