package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/erry-az/go-init/proto/api/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type HTTPServer struct {
	server *http.Server
	mux    *runtime.ServeMux
}

func NewHTTPServer(grpcPort string) (*HTTPServer, error) {
	// Create gRPC connection for gateway
	conn, err := grpc.NewClient("localhost:"+grpcPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	// Create HTTP gateway mux
	mux := runtime.NewServeMux()

	// Register gRPC-Gateway handlers
	err = v1.RegisterUserServiceHandler(context.Background(), mux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register user service handler: %w", err)
	}

	err = v1.RegisterProductServiceHandler(context.Background(), mux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register product service handler: %w", err)
	}

	return &HTTPServer{
		mux: mux,
	}, nil
}

func (s *HTTPServer) Start(ctx context.Context, port string) error {
	// Create HTTP server
	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: s.mux,
	}

	log.Printf("HTTP server starting on port %s", port)

	// Start server in goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	log.Println("Shutting down HTTP server...")
	return s.server.Shutdown(context.Background())
}

func (s *HTTPServer) Stop() error {
	if s.server != nil {
		return s.server.Shutdown(context.Background())
	}
	return nil
}
