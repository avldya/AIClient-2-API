package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/avldya/AIClient-2-API/go-kiro-api/internal/api"
	"github.com/avldya/AIClient-2-API/go-kiro-api/internal/config"
	"github.com/avldya/AIClient-2-API/go-kiro-api/internal/kiro"
)

func main() {
	// Parse command line arguments
	configFile := flag.String("config", "", "Path to configuration file")
	host := flag.String("host", "", "Server host address")
	port := flag.Int("port", 0, "Server port")
	apiKey := flag.String("api-key", "", "API key for authentication")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override with command line arguments
	if *host != "" {
		cfg.Server.Host = *host
	}
	if *port != 0 {
		cfg.Server.Port = *port
	}
	if *apiKey != "" {
		cfg.Server.APIKey = *apiKey
	}

	// Create Kiro client
	kiroClient := kiro.NewClient(cfg.ToMap())

	// Initialize client
	if err := kiroClient.Initialize(); err != nil {
		log.Fatalf("Failed to initialize Kiro client: %v", err)
	}

	log.Println("Kiro client initialized successfully")

	// Create API handler
	handler := api.NewHandler(kiroClient, cfg.Server.APIKey)

	// Set up routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", handler.HandleHealthCheck)

	// OpenAI compatible endpoints
	mux.HandleFunc("/v1/chat/completions", 
		handler.LoggingMiddleware(
			handler.CORSMiddleware(
				handler.AuthMiddleware(handler.HandleOpenAIChatCompletions))))

	// Claude compatible endpoints
	mux.HandleFunc("/v1/messages", 
		handler.LoggingMiddleware(
			handler.CORSMiddleware(
				handler.AuthMiddleware(handler.HandleClaudeMessages))))

	// Model listing
	mux.HandleFunc("/v1/models", 
		handler.LoggingMiddleware(
			handler.CORSMiddleware(
				handler.AuthMiddleware(handler.HandleListModels))))

	// Token counting
	mux.HandleFunc("/v1/messages/count_tokens", 
		handler.LoggingMiddleware(
			handler.CORSMiddleware(
				handler.AuthMiddleware(handler.HandleCountTokens))))

	// Usage limits
	mux.HandleFunc("/getUsageLimits", 
		handler.LoggingMiddleware(
			handler.CORSMiddleware(
				handler.AuthMiddleware(handler.HandleGetUsageLimits))))

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("API endpoints:")
	log.Printf("  - OpenAI: http://%s/v1/chat/completions", addr)
	log.Printf("  - Claude: http://%s/v1/messages", addr)
	log.Printf("  - Models: http://%s/v1/models", addr)
	log.Printf("  - Health: http://%s/health", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
