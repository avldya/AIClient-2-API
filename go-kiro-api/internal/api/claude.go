package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/avldya/AIClient-2-API/go-kiro-api/pkg/models"
)

// HandleClaudeMessages handles the /v1/messages endpoint (Claude format)
// 处理 /v1/messages 端点（Claude 格式）
func (h *Handler) HandleClaudeMessages(w http.ResponseWriter, r *http.Request) {
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
		h.handleClaudeStreamingRequest(w, r, &req)
		return
	}

	// Handle non-streaming
	h.handleClaudeNonStreamingRequest(w, r, &req)
}

// handleClaudeNonStreamingRequest handles non-streaming Claude requests
// 处理非流式 Claude 请求
func (h *Handler) handleClaudeNonStreamingRequest(w http.ResponseWriter, r *http.Request, req *models.ChatCompletionRequest) {
	// Make request to Kiro API
	response, err := h.client.GenerateContent(req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert response to Claude format
	// The response should already be in the correct format from the client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleClaudeStreamingRequest handles streaming Claude requests
// 处理流式 Claude 请求
func (h *Handler) handleClaudeStreamingRequest(w http.ResponseWriter, r *http.Request, req *models.ChatCompletionRequest) {
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

	// Convert to Claude format and stream
	claudeStream := streamParser.ConvertToClaudeStream(eventChan)

	for data := range claudeStream {
		if err := h.writeSSE(w, data); err != nil {
			return
		}
	}
}
