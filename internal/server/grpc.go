package server

import (
	"context"
	"fmt"
	apiv1 "github.com/erry-az/go-init/api/v1"
	"github.com/erry-az/go-init/internal/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

// GRPCServer represents the gRPC server
type GRPCServer struct {
	userService *service.UserService
	grpcServer  *grpc.Server
	grpcPort    int
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(userService *service.UserService, grpcPort int) *GRPCServer {
	return &GRPCServer{
		userService: userService,
		grpcPort:    grpcPort,
	}
}

// Start starts the gRPC and HTTP servers
func (s *GRPCServer) Start(ctx context.Context) error {
	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.grpcServer = grpc.NewServer()
	apiv1.RegisterUserServiceServer(s.grpcServer, s.userService)

	go func() {
		log.Printf("Starting gRPC server on port %d", s.grpcPort)
		if err := s.grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	return nil
}

// Stop stops the gRPC and HTTP servers
func (s *GRPCServer) Stop(ctx context.Context) error {
	// Stop gRPC server
	s.grpcServer.GracefulStop()

	return nil
}
