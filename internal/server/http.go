package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/erry-az/go-sample/proto/api/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type HTTPServer struct {
	server       *http.Server
	mux          *runtime.ServeMux
	swaggerSpecs map[string]string
}

type SwaggerSpec struct {
	Name string `json:"name"`
	Path string `json:"path"`
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

	// Load swagger specifications
	swaggerSpecs, err := loadSwaggerSpecs()
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger specs: %w", err)
	}

	return &HTTPServer{
		mux:          mux,
		swaggerSpecs: swaggerSpecs,
	}, nil
}

func (s *HTTPServer) Start(ctx context.Context, port string) error {
	// Create main HTTP mux to combine gRPC gateway and swagger
	mainMux := http.NewServeMux()

	// Mount gRPC gateway
	mainMux.Handle("/", s.mux)

	// Mount swagger endpoints
	s.setupSwaggerRoutes(mainMux)

	// Create HTTP server
	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: mainMux,
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

func loadSwaggerSpecs() (map[string]string, error) {
	specs := make(map[string]string)
	docsDir := "docs"

	err := filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".swagger.json") {
			relPath := strings.TrimPrefix(path, docsDir+"/")
			name := strings.TrimSuffix(filepath.Base(path), ".swagger.json")

			// Create a more descriptive name based on path
			parts := strings.Split(relPath, "/")
			if len(parts) > 1 {
				name = strings.Join(parts[:len(parts)-1], "/") + "/" + name
			}

			specs[name] = path
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return specs, nil
}

func (s *HTTPServer) setupSwaggerRoutes(mux *http.ServeMux) {
	// Swagger UI endpoint
	mux.HandleFunc("/swagger/", s.serveSwaggerUI)

	// Swagger specs list endpoint
	mux.HandleFunc("/swagger/specs", s.serveSwaggerSpecs)

	// Individual swagger spec endpoints
	for name, path := range s.swaggerSpecs {
		specPath := "/swagger/spec/" + name
		mux.HandleFunc(specPath, s.serveSwaggerSpec(path))
	}
}

func (s *HTTPServer) serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	swaggerHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" />
    <style>
        .swagger-ui .topbar { display: none; }
        .spec-selector {
            margin: 20px 0;
            padding: 10px;
            background: #f8f9fa;
            border-radius: 5px;
        }
        .spec-selector select {
            padding: 8px 12px;
            font-size: 14px;
            border: 1px solid #ccc;
            border-radius: 4px;
            background: white;
            min-width: 300px;
        }
    </style>
</head>
<body>
    <div id="swagger-ui">
        <div class="spec-selector">
            <label for="spec-select">Select API Specification: </label>
            <select id="spec-select" onchange="loadSpec()">
                <option value="">Choose a specification...</option>
            </select>
        </div>
    </div>
    
    <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
    <script>
        let ui;
        
        async function loadSpecs() {
            try {
                const response = await fetch('/swagger/specs');
                const specs = await response.json();
                const select = document.getElementById('spec-select');
                
                specs.forEach(spec => {
                    const option = document.createElement('option');
                    option.value = spec.path;
                    option.textContent = spec.name;
                    select.appendChild(option);
                });
                
                // Load first spec by default if available
                if (specs.length > 0) {
                    select.value = specs[0].path;
                    loadSpec();
                }
            } catch (error) {
                console.error('Failed to load swagger specs:', error);
            }
        }
        
        function loadSpec() {
            const select = document.getElementById('spec-select');
            const specPath = select.value;
            
            if (!specPath) return;
            
            // Create a container for SwaggerUI that preserves the selector
            const swaggerContainer = document.getElementById('swagger-ui');
            let uiContainer = document.getElementById('swagger-ui-container');
            
            if (!uiContainer) {
                uiContainer = document.createElement('div');
                uiContainer.id = 'swagger-ui-container';
                swaggerContainer.appendChild(uiContainer);
            } else {
                uiContainer.innerHTML = '';
            }
            
            ui = SwaggerUIBundle({
                url: specPath,
                dom_id: '#swagger-ui-container',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.presets.standalone
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ]
            });
        }
        
        // Load specs on page load
        loadSpecs();
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(swaggerHTML))
}

func (s *HTTPServer) serveSwaggerSpecs(w http.ResponseWriter, r *http.Request) {
	var specs []SwaggerSpec
	for name := range s.swaggerSpecs {
		specs = append(specs, SwaggerSpec{
			Name: name,
			Path: "/swagger/spec/" + name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(specs)
}

func (s *HTTPServer) serveSwaggerSpec(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, filePath)
	}
}
