package grpc

import (
	"fmt"
	"net"

	"buf.build/go/protovalidate"
	protovalidateMidleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	cfg *Config
	*grpc.Server
}

func NewServer(config *Config) (*Server, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	// Create gRPC endpoint
	server := grpc.NewServer(
		grpc.UnaryInterceptor(protovalidateMidleware.UnaryServerInterceptor(validator)),
	)

	reflection.Register(server)

	return &Server{
		cfg:    config,
		Server: server,
	}, nil
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.cfg.Port, err)
	}

	return s.Serve(lis)
}
