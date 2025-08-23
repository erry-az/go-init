package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/erry-az/go-init/config"
	handlergrpc "github.com/erry-az/go-init/internal/handler/grpc"
	"github.com/erry-az/go-init/internal/repository/sqlc"
	"github.com/erry-az/go-init/internal/server"
	"github.com/erry-az/go-init/internal/server/http"
	"github.com/erry-az/go-init/internal/usecase"
	"github.com/erry-az/go-init/pkg/watmil"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App represents the application with all dependencies
type App struct {
	// Business logic components
	UserUsecase    usecase.UserUsecase
	ProductUsecase usecase.ProductUsecase
	UserService    *handlergrpc.UserService
	ProductService *handlergrpc.ProductService
	Publisher      *cqrs.EventBus

	// Infrastructure components
	config     *config.Config
	dbPool     *pgxpool.Pool
	logger     watermill.LoggerAdapter
	grpcServer *server.GRPCServer
	httpServer *http.HTTPServer
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewEndpoint creates a new application with all dependencies wired
func NewEndpoint(cfg *config.Config) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize database connection
	if err := app.initDatabase(); err != nil {
		cancel()
		return nil, err
	}

	// Initialize logger
	app.initLogger()

	// Initialize business logic components
	if err := app.initBusinessLogic(); err != nil {
		cancel()
		return nil, err
	}

	// Initialize servers
	if err := app.initServers(); err != nil {
		cancel()
		return nil, err
	}

	return app, nil
}

// initDatabase initializes the database connection pool
func (a *App) initDatabase() error {
	dbPool, err := pgxpool.New(a.ctx, a.config.Databases.DbDsn)
	if err != nil {
		slog.Error("Failed to create pgx pool", slog.Any("error", err))
		return err
	}

	// Test database connection
	if err := dbPool.Ping(a.ctx); err != nil {
		slog.Error("Failed to ping database", slog.Any("error", err))
		dbPool.Close()
		return err
	}

	a.dbPool = dbPool
	slog.Info("Database connection established")
	return nil
}

// initLogger initializes the watermill logger
func (a *App) initLogger() {
	a.logger = watermill.NewSlogLogger(slog.Default())
}

// initBusinessLogic initializes business logic components
func (a *App) initBusinessLogic() error {
	// Create Watermill publisher
	publisher, err := watmil.NewPublisher(a.dbPool, a.logger)
	if err != nil {
		return err
	}

	// Create SQLC querier
	querier := sqlc.New(a.dbPool)

	// Create usecases
	a.UserUsecase = usecase.NewUserUsecase(querier, publisher)
	a.ProductUsecase = usecase.NewProductUsecase(querier, publisher)

	// Create services
	a.UserService = handlergrpc.NewUserService(a.UserUsecase)
	a.ProductService = handlergrpc.NewProductService(a.ProductUsecase)
	a.Publisher = publisher

	slog.Info("Business logic components initialized")
	return nil
}

// initServers initializes gRPC and HTTP servers
func (a *App) initServers() error {
	// Create gRPC endpoint with services
	grpcServer, err := server.NewGRPCServer(server.GRPCServices{
		UserService:    a.UserService,
		ProductService: a.ProductService,
	})
	if err != nil {
		slog.Error("Failed to create gRPC endpoint", slog.Any("error", err))
		return err
	}
	a.grpcServer = grpcServer

	// Create HTTP endpoint (gRPC Gateway)
	httpServer, err := http.NewHTTPServer(a.config.Servers.GrpcPort)
	if err != nil {
		slog.Error("Failed to create HTTP endpoint", slog.Any("error", err))
		return err
	}
	a.httpServer = httpServer

	slog.Info("Servers initialized")
	return nil
}

// Start starts the application servers and handles graceful shutdown
func (a *App) Start() error {
	// Start gRPC endpoint
	go func() {
		if err := a.grpcServer.Start(a.ctx, a.config.Servers.GrpcPort); err != nil {
			slog.Error("gRPC endpoint error", slog.Any("error", err))
		}
	}()

	// Start HTTP endpoint
	go func() {
		if err := a.httpServer.Start(a.ctx, a.config.Servers.HttpPort); err != nil {
			slog.Error("HTTP endpoint error", slog.Any("error", err))
		}
	}()

	slog.Info("🚀 Application started successfully")
	slog.Info("📡 gRPC endpoint listening", "port", a.config.Servers.GrpcPort)
	slog.Info("🌐 HTTP endpoint listening", "port", a.config.Servers.HttpPort)
	slog.Info("👋 Press Ctrl+C to gracefully shutdown...")

	// Wait for interrupt signal
	return a.waitForShutdown()
}

// waitForShutdown waits for shutdown signals and handles graceful shutdown
func (a *App) waitForShutdown() error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("🛑 Shutdown signal received, starting graceful shutdown...")

	// Cancel context to signal shutdown to all components
	a.cancel()

	// Stop servers
	if a.grpcServer != nil {
		a.grpcServer.Stop()
		slog.Info("✅ gRPC endpoint stopped")
	}

	if a.httpServer != nil {
		a.httpServer.Stop()
		slog.Info("✅ HTTP endpoint stopped")
	}

	// Close database connection
	if a.dbPool != nil {
		a.dbPool.Close()
		slog.Info("✅ Database connection closed")
	}

	slog.Info("🎉 Application shutdown completed successfully")
	return nil
}

// Close performs cleanup of application resources
func (a *App) Close() error {
	if a.cancel != nil {
		a.cancel()
	}

	if a.dbPool != nil {
		a.dbPool.Close()
	}

	return nil
}
