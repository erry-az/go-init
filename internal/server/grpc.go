package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"buf.build/go/protovalidate"
	handlergrpc "github.com/erry-az/go-init/internal/handler/grpc"
	protovalidateMidleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	server *grpc.Server
}

type GRPCServices struct {
	UserService    *handlergrpc.UserService
	ProductService *handlergrpc.ProductService
}

func NewGRPCServer(services GRPCServices) (*GRPCServer, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	// Create gRPC endpoint
	server := grpc.NewServer(
		grpc.UnaryInterceptor(protovalidateMidleware.UnaryServerInterceptor(validator)),
	)

	return &GRPCServer{
		server: server,
	}, nil
}

func (s *GRPCServer) Start(ctx context.Context, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	log.Printf("gRPC endpoint starting on port %s", port)

	reflection.Register(s.server)

	return s.server.Serve(lis)
}

func (s *GRPCServer) RegisterService(services ...func(s *grpc.Server)) {
	for _, service := range services {
		service(s.server)
	}
}

func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}
