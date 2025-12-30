package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go-antigravity-api/internal/antigravity"
	"go-antigravity-api/internal/api"
	"go-antigravity-api/internal/config"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	projectID := flag.String("project-id", "", "Google Cloud Project ID")
	port := flag.Int("port", 0, "Server port (overrides config)")
	flag.Parse()

	// Load configuration
	var cfg *config.Config
	var err error

	if *configPath != "" {
		cfg, err = config.Load(*configPath)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		log.Printf("Loaded configuration from: %s", *configPath)
	} else {
		cfg = config.LoadFromEnv()
		log.Println("Using configuration from environment variables")
	}

	// Override with command line flags
	if *projectID != "" {
		cfg.Antigravity.ProjectID = *projectID
	}
	if *port != 0 {
		cfg.Server.Port = *port
	}

	// Create Antigravity service
	service := antigravity.NewService(cfg.ToMap())

	// Initialize service
	log.Println("Initializing Antigravity API service...")
	if err := service.Initialize(); err != nil {
		log.Printf("Warning: Failed to initialize service: %v", err)
		log.Println("Service will attempt to initialize on first request")
	}

	// Create API handler
	handler := api.NewHandler(service, cfg.Server.APIKey)

	// Setup routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", handler.HandleHealth)

	// Model listing
	mux.HandleFunc("/v1beta/models", handler.HandleListModels)

	// Content generation (non-streaming)
	mux.HandleFunc("/v1beta/models/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if contains(r.URL.Path, ":streamGenerateContent") {
			handler.HandleStreamGenerateContent(w, r)
		} else if contains(r.URL.Path, ":generateContent") {
			handler.HandleGenerateContent(w, r)
		} else {
			http.Error(w, "invalid endpoint", http.StatusNotFound)
		}
	})

	// Usage/quota
	mux.HandleFunc("/usage", handler.HandleGetUsage)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("\nShutting down server...")
		if err := server.Close(); err != nil {
			log.Printf("Error closing server: %v", err)
		}
	}()

	// Print startup information
	log.Println("========================================")
	log.Println("  Antigravity API Server Started")
	log.Println("========================================")
	log.Printf("Server listening on: http://%s", addr)
	log.Printf("Project ID: %s", service.GetProjectID())
	log.Println("")
	log.Println("Available endpoints:")
	log.Println("  GET  /health - Health check")
	log.Println("  GET  /v1beta/models - List models")
	log.Println("  POST /v1beta/models/{model}:generateContent - Generate content")
	log.Println("  POST /v1beta/models/{model}:streamGenerateContent - Stream content")
	log.Println("  GET  /usage - Get usage limits")
	log.Println("")
	log.Println("Press Ctrl+C to stop")
	log.Println("========================================")

	// Start listening
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}
