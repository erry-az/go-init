package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"buf.build/go/protovalidate"
	"github.com/erry-az/go-sample/internal/app"
	"github.com/erry-az/go-sample/proto/api/v1"
	protovalidateMidleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	server *grpc.Server
	app    *app.App
}

func NewGRPCServer(application *app.App) (*GRPCServer, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	// Create gRPC server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(protovalidateMidleware.UnaryServerInterceptor(validator)),
	)

	// Register services
	v1.RegisterUserServiceServer(server, application.UserService)
	v1.RegisterProductServiceServer(server, application.ProductService)

	// Enable reflection for development
	reflection.Register(server)

	return &GRPCServer{
		server: server,
		app:    application,
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
