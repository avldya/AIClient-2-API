package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/avldya/AIClient-2-API/go-kiro-api/pkg/models"
)

// HandleOpenAIChatCompletions handles the /v1/chat/completions endpoint (OpenAI format)
// 处理 /v1/chat/completions 端点（OpenAI 格式）
func (h *Handler) HandleOpenAIChatCompletions(w http.ResponseWriter, r *http.Request) {
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

	var req models.ChatCompletionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.handleError(w, err)
		return
	}

	// Validate request
	if req.Model == "" {
		req.Model = "claude-opus-4-5"
	}
	if len(req.Messages) == 0 {
		h.handleError(w, &ValidationError{Message: "messages field is required"})
		return
	}

	// Handle streaming
	if req.Stream {
		h.handleOpenAIStreamingRequest(w, r, &req)
		return
	}

	// Handle non-streaming
	h.handleOpenAINonStreamingRequest(w, r, &req)
}

// handleOpenAINonStreamingRequest handles non-streaming OpenAI requests
// 处理非流式 OpenAI 请求
func (h *Handler) handleOpenAINonStreamingRequest(w http.ResponseWriter, r *http.Request, req *models.ChatCompletionRequest) {
	// Make request to Kiro API
	response, err := h.client.GenerateContent(req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleOpenAIStreamingRequest handles streaming OpenAI requests
// 处理流式 OpenAI 请求
func (h *Handler) handleOpenAIStreamingRequest(w http.ResponseWriter, r *http.Request, req *models.ChatCompletionRequest) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Get streaming response
	eventChan, err := h.client.GenerateContentStream(req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Create stream parser
	streamParser := h.client.GetStreamParser()

	// Convert to OpenAI format and stream
	openaiStream := streamParser.ConvertToOpenAIStream(eventChan, req.Model)

	for data := range openaiStream {
		if err := h.writeSSE(w, data); err != nil {
			return
		}
	}
}

// ValidationError represents a validation error
// 验证错误
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
