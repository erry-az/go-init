package server

import (
	"context"
	"errors"
	"fmt"
	apiv1 "github.com/erry-az/go-init/api/v1"
	"github.com/erry-az/go-init/internal/service"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"log"
	"net/http"
	"strings"
)

// HttpServer represents the http server
type HttpServer struct {
	userService *service.UserService
	httpServer  *http.Server
	httpPort    int
}

func NewHttpServer(userService *service.UserService, httpPort int) *HttpServer {
	return &HttpServer{
		userService: userService,
		httpPort:    httpPort,
	}
}

func (h *HttpServer) Start(ctx context.Context) error {
	// Start HTTP gateway
	mux := runtime.NewServeMux()
	err := apiv1.RegisterUserServiceHandlerServer(ctx, mux, h.userService)
	if err != nil {
		return fmt.Errorf("failed to register gateway handler: %w", err)
	}

	// Create an HTTP mux to serve both the gRPC gateway and the Swagger docs
	httpMux := http.NewServeMux()
	httpMux.Handle("/", mux)

	// Add handler for serving Swagger UI
	httpMux.HandleFunc("/docs/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Serve the Swagger JSON file directly if requested
		if strings.HasSuffix(path, "service.swagger.json") {
			w.Header().Set("Content-Type", "application/json")
			http.ServeFile(w, r, "docs/v1/service.swagger.json")
			return
		}

		// Otherwise serve the Swagger UI
		if path == "/docs" || path == "/docs/" {
			renderSwaggerUI(w, "/docs/service.swagger.json")
			return
		}

		// Handle 404 for any other paths under /docs/
		http.NotFound(w, r)
	})

	h.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", h.httpPort),
		Handler: httpMux,
	}

	go func() {
		log.Printf("Starting HTTP gateway on port %d", h.httpPort)
		if err := h.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to serve HTTP gateway: %v", err)
		}
	}()

	return nil
}

// renderSwaggerUI writes the Swagger UI HTML to the response
func renderSwaggerUI(w http.ResponseWriter, specURL string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui.css">
  <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/favicon-32x32.png" sizes="32x32">
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; padding: 0; }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>

  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "` + specURL + `",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout"
      });
      window.ui = ui;
    };
  </script>
</body>
</html>`

	w.Write([]byte(html))
}

func (h *HttpServer) Stop(ctx context.Context) error {
	if h.httpServer == nil {
		return nil
	}

	return h.httpServer.Shutdown(ctx)
}
