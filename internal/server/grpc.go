package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"buf.build/go/protovalidate"
	watermillbase "github.com/ThreeDotsLabs/watermill"
	handlergrpc "github.com/erry-az/go-init/internal/handler/grpc"
	"github.com/erry-az/go-init/internal/repository/sqlc"
	"github.com/erry-az/go-init/pkg/watermill"
	"github.com/erry-az/go-init/proto/api/v1"
	protovalidateMidleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	server         *grpc.Server
	userService    *handlergrpc.UserService
	productService *handlergrpc.ProductService
}

func NewGRPCServer(dbPool *pgxpool.Pool, sqlDB *sql.DB, logger watermillbase.LoggerAdapter) (*GRPCServer, error) {
	// Create Watermill publisher (needs sql.DB)
	publisher, err := watermill.NewPublisher(sqlDB, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	// Create SQLC querier (needs pgx connection)
	querier := sqlc.New(dbPool)

	// Create services
	userService := handlergrpc.NewUserService(querier, publisher)
	productService := handlergrpc.NewProductService(querier, publisher)

	validator, err := protovalidate.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	// Create gRPC server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(protovalidateMidleware.UnaryServerInterceptor(validator)),
	)

	// Register services
	v1.RegisterUserServiceServer(server, userService)
	v1.RegisterProductServiceServer(server, productService)

	// Enable reflection for development
	reflection.Register(server)

	return &GRPCServer{
		server:         server,
		userService:    userService,
		productService: productService,
	}, nil
}

func (s *GRPCServer) Start(ctx context.Context, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	log.Printf("gRPC server starting on port %s", port)

	// Start server in goroutine
	go func() {
		if err := s.server.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	log.Println("Shutting down gRPC server...")
	s.server.GracefulStop()

	return nil
}

func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}
