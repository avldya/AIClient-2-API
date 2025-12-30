package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"go-antigravity-api/internal/antigravity"
	"go-antigravity-api/pkg/models"
)

// Handler handles HTTP requests
type Handler struct {
	service *antigravity.Service
	apiKey  string
}

// NewHandler creates a new Handler
func NewHandler(service *antigravity.Service, apiKey string) *Handler {
	return &Handler{
		service: service,
		apiKey:  apiKey,
	}
}

// authenticateRequest checks API key if configured
func (h *Handler) authenticateRequest(r *http.Request) error {
	if h.apiKey == "" {
		return nil // No authentication required
	}

	// Check Authorization header
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return fmt.Errorf("missing authorization header")
	}

	// Check Bearer token
	if !strings.HasPrefix(auth, "Bearer ") {
		return fmt.Errorf("invalid authorization header format")
	}

	token := strings.TrimPrefix(auth, "Bearer ")
	if token != h.apiKey {
		return fmt.Errorf("invalid API key")
	}

	return nil
}

// HandleListModels handles GET /v1beta/models
func (h *Handler) HandleListModels(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticateRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	models, err := h.service.ListModels()
	if err != nil {
		log.Printf("[API] Error listing models: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

// HandleGenerateContent handles POST /v1beta/models/{model}:generateContent
func (h *Handler) HandleGenerateContent(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticateRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Extract model from path
	model := extractModelFromPath(r.URL.Path, ":generateContent")
	if model == "" {
		http.Error(w, "invalid model path", http.StatusBadRequest)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request models.GeminiRequest
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Generate content
	response, err := h.service.GenerateContent(model, &request)
	if err != nil {
		log.Printf("[API] Error generating content: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleStreamGenerateContent handles POST /v1beta/models/{model}:streamGenerateContent
func (h *Handler) HandleStreamGenerateContent(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticateRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Extract model from path
	model := extractModelFromPath(r.URL.Path, ":streamGenerateContent")
	if model == "" {
		http.Error(w, "invalid model path", http.StatusBadRequest)
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var request models.GeminiRequest
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Set headers for streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Stream content
	responseChan, errorChan := h.service.StreamGenerateContent(model, &request)

	for {
		select {
		case response, ok := <-responseChan:
			if !ok {
				return
			}

			// Write SSE event
			data, err := json.Marshal(response)
			if err != nil {
				log.Printf("[API] Error marshaling response: %v", err)
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

		case err, ok := <-errorChan:
			if ok && err != nil {
				log.Printf("[API] Error streaming content: %v", err)
				// Write error as SSE event
				fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
				flusher.Flush()
			}
			return
		}
	}
}

// HandleGetUsage handles GET /usage
func (h *Handler) HandleGetUsage(w http.ResponseWriter, r *http.Request) {
	if err := h.authenticateRequest(r); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	usage, err := h.service.GetUsageLimits()
	if err != nil {
		log.Printf("[API] Error getting usage: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// HandleHealth handles GET /health
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"projectId": h.service.GetProjectID(),
	})
}

// extractModelFromPath extracts the model name from the URL path
func extractModelFromPath(path string, suffix string) string {
	// Path format: /v1beta/models/{model}:action
	path = strings.TrimSuffix(path, suffix)
	parts := strings.Split(path, "/")
	if len(parts) >= 4 && parts[1] == "v1beta" && parts[2] == "models" {
		return parts[3]
	}
	return ""
}
