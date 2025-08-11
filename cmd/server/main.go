package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/erry-az/go-init/config"
	"github.com/erry-az/go-init/internal/server"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create pgxpool connection for SQLC
	dbPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to create pgx pool: %v", err)
	}
	defer dbPool.Close()

	// Test pgxpool connection
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database pool: %v", err)
	}

	// Create standard SQL connection for Watermill
	sqlDB, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to SQL database: %v", err)
	}
	defer sqlDB.Close()

	// Test SQL database connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping SQL database: %v", err)
	}

	// Create logger
	logger := watermill.NewStdLogger(false, false)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and start gRPC server
	grpcServer, err := server.NewGRPCServer(dbPool, sqlDB, logger)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	go func() {
		if err := grpcServer.Start(ctx, "9090"); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Create and start HTTP server (gRPC Gateway)
	httpServer, err := server.NewHTTPServer("9090")
	if err != nil {
		log.Fatalf("Failed to create HTTP server: %v", err)
	}

	go func() {
		if err := httpServer.Start(ctx, "8080"); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Println("Server started:")
	log.Println("  gRPC server: :9090")
	log.Println("  HTTP server: :8080")
	log.Println("  API docs: http://localhost:8080/docs")
	log.Println("Press Ctrl+C to exit...")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down servers...")
	cancel()

	grpcServer.Stop()
	httpServer.Stop()

	log.Println("Servers shut down successfully")
}