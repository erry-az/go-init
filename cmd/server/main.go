package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/erry-az/go-sample/config"
	"github.com/erry-az/go-sample/internal/app"
	"github.com/erry-az/go-sample/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	sloger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(sloger)

	// Load configuration
	cfg, err := config.New()
	if err != nil {
		slog.Error("Failed to load config: ", slog.Any("error", err))
		return
	}

	// Create pgxpool connection for SQLC
	dbPool, err := pgxpool.New(context.Background(), cfg.Databases.DbDsn)
	if err != nil {
		slog.Error("Failed to create pgx pool ", slog.Any("error", err))
		return
	}
	defer dbPool.Close()

	// Test pgxpool connection
	if err := dbPool.Ping(context.Background()); err != nil {
		slog.Error("Failed to ping database pool ", slog.Any("error", err))
		return
	}

	// Create standard SQL connection for Watermill
	sqlDB, err := sql.Open("pgx", cfg.Databases.PgMqUrl)
	if err != nil {
		slog.Error("Failed to connect to SQL databas v", slog.Any("error", err))
		return
	}
	defer sqlDB.Close()

	// Test SQL database connection
	if err := sqlDB.Ping(); err != nil {
		slog.Error("Failed to ping SQL database ", slog.Any("error", err))
		return
	}

	// Create logger
	logger := watermill.NewSlogLogger(slog.Default())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create application with dependencies
	application, err := app.New(app.Dependencies{
		DBPool: dbPool,
		SqlDB:  sqlDB,
		Logger: logger,
		Config: cfg,
	})
	if err != nil {
		slog.Error("Failed to create application ", slog.Any("error", err))
		return
	}

	// Create and start gRPC server
	grpcServer, err := server.NewGRPCServer(application)
	if err != nil {
		slog.Error("Failed to create gRPC server ", slog.Any("error", err))
		return
	}

	go func() {
		if err := grpcServer.Start(ctx, cfg.Servers.GrpcPort); err != nil {
			log.Printf("gRPC server error: %v", slog.Any("error", err))
			return
		}
	}()

	// Create and start HTTP server (gRPC Gateway)
	httpServer, err := server.NewHTTPServer(cfg.Servers.GrpcPort)
	if err != nil {
		slog.Error("Failed to create HTTP server ", slog.Any("error", err))
		return
	}

	go func() {
		if err := httpServer.Start(ctx, cfg.Servers.HttpPort); err != nil {
			log.Printf("HTTP server error: %v", slog.Any("error", err))
			return
		}
	}()

	slog.Info("Server started:")
	slog.Info("  gRPC server: ", "port", cfg.Servers.GrpcPort)
	slog.Info("  HTTP server: ", "port", cfg.Servers.HttpPort)
	slog.Info("Press Ctrl+C to exit...")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("Shutting down servers...")
	cancel()

	grpcServer.Stop()
	httpServer.Stop()

	slog.Info("Servers shut down successfully")
}
