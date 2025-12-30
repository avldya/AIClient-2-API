package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/avldya/AIClient-2-API/go-kiro-api/internal/kiro"
	"github.com/avldya/AIClient-2-API/go-kiro-api/pkg/models"
)

// Handler represents the API handler
// API 处理器
type Handler struct {
	client *kiro.Client
	apiKey string
}

// NewHandler creates a new API handler
// 创建新的 API 处理器
func NewHandler(client *kiro.Client, apiKey string) *Handler {
	return &Handler{
		client: client,
		apiKey: apiKey,
	}
}

// AuthMiddleware validates API key
// API 密钥验证中间件
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if no API key is configured
		if h.apiKey == "" {
			next(w, r)
			return
		}

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" && parts[1] == h.apiKey {
				next(w, r)
				return
			}
		}

		// Check x-api-key header
		if r.Header.Get("x-api-key") == h.apiKey {
			next(w, r)
			return
		}

		// Check query parameter
		if r.URL.Query().Get("key") == h.apiKey {
			next(w, r)
			return
		}

		// Unauthorized
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"message": "Invalid API key",
				"type":    "invalid_request_error",
			},
		})
	}
}

// CORSMiddleware adds CORS headers
// CORS 中间件
func (h *Handler) CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-api-key")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// LoggingMiddleware logs requests
// 日志中间件
func (h *Handler) LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next(w, r)
	}
}

// HandleListModels handles the /v1/models endpoint
// 处理 /v1/models 端点
func (h *Handler) HandleListModels(w http.ResponseWriter, r *http.Request) {
	models := h.client.ListModels()

	response := map[string]interface{}{
		"object": "list",
		"data":   make([]map[string]interface{}, len(models)),
	}

	for i, model := range models {
		response["data"].([]map[string]interface{})[i] = map[string]interface{}{
			"id":       model,
			"object":   "model",
			"created":  1234567890,
			"owned_by": "kiro",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetUsageLimits handles the /getUsageLimits endpoint
// 处理 /getUsageLimits 端点
func (h *Handler) HandleGetUsageLimits(w http.ResponseWriter, r *http.Request) {
	usage, err := h.client.GetUsageLimits()
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

// HandleCountTokens handles the /v1/messages/count_tokens endpoint
// 处理 /v1/messages/count_tokens 端点
func (h *Handler) HandleCountTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.handleError(w, err)
		return
	}
	defer r.Body.Close()

	var req struct {
		Messages []models.Message `json:"messages"`
		System   string           `json:"system,omitempty"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		h.handleError(w, err)
		return
	}

	// Estimate token count
	totalTokens := 0
	if req.System != "" {
		totalTokens += kiro.CountTokens(req.System)
	}

	for _, msg := range req.Messages {
		content := kiro.GetContentText(msg.Content)
		totalTokens += kiro.CountTokens(content)
	}

	response := map[string]interface{}{
		"input_tokens": totalTokens,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleHealthCheck handles the /health endpoint
// 处理健康检查端点
func (h *Handler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// handleError handles errors and sends appropriate response
// 处理错误并发送适当的响应
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	log.Printf("Error: %v", err)

	statusCode := http.StatusInternalServerError
	errorType := "internal_error"
	message := err.Error()

	// Check for API errors
	if apiErr, ok := err.(*kiro.APIError); ok {
		statusCode = apiErr.StatusCode
		if statusCode == 429 {
			errorType = "rate_limit_error"
		} else if statusCode >= 400 && statusCode < 500 {
			errorType = "invalid_request_error"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"message": message,
			"type":    errorType,
		},
	})
}

// writeSSE writes Server-Sent Events
// 写入 SSE 流
func (h *Handler) writeSSE(w http.ResponseWriter, data string) error {
	_, err := fmt.Fprint(w, data)
	if err != nil {
		return err
	}

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}
