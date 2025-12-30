package antigravity

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"go-antigravity-api/pkg/models"
)

// StreamAPI makes a streaming API call
func (c *APIClient) StreamAPI(method string, body interface{}) (<-chan *models.GeminiResponse, <-chan error) {
	responseChan := make(chan *models.GeminiResponse, 10)
	errorChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errorChan)

		if err := c.streamAPIWithRetry(method, body, responseChan, false, 0, 0); err != nil {
			errorChan <- err
		}
	}()

	return responseChan, errorChan
}

// streamAPIWithRetry makes a streaming API call with retry logic
func (c *APIClient) streamAPIWithRetry(method string, body interface{}, responseChan chan<- *models.GeminiResponse, isRetry bool, retryCount int, baseURLIndex int) error {
	if baseURLIndex >= len(c.baseURLs) {
		return fmt.Errorf("all Antigravity base URLs failed")
	}

	baseURL := c.baseURLs[baseURLIndex]
	url := fmt.Sprintf("%s/%s:%s?alt=sse", baseURL, APIVersion, method)

	// Get access token
	accessToken, err := c.authManager.GetAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Prepare request body
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/event-stream")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[Antigravity API] Network error during stream on %s: %v", baseURL, err)
		if baseURLIndex+1 < len(c.baseURLs) {
			log.Println("[Antigravity API] Trying next base URL...")
			return c.streamAPIWithRetry(method, body, responseChan, isRetry, retryCount, baseURLIndex+1)
		}
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	// Handle different status codes
	switch resp.StatusCode {
	case http.StatusOK:
		return c.parseSSEStream(resp.Body, responseChan)

	case http.StatusBadRequest, http.StatusUnauthorized:
		if !isRetry {
			log.Println("[Antigravity API] Received 401/400 during stream. Refreshing auth and retrying...")
			if err := c.authManager.Initialize(true); err != nil {
				return fmt.Errorf("failed to refresh auth: %w", err)
			}
			return c.streamAPIWithRetry(method, body, responseChan, true, retryCount, baseURLIndex)
		}
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed: %s", string(body))

	case http.StatusTooManyRequests:
		if baseURLIndex+1 < len(c.baseURLs) {
			log.Printf("[Antigravity API] Rate limited on %s. Trying next base URL...", baseURL)
			return c.streamAPIWithRetry(method, body, responseChan, isRetry, retryCount, baseURLIndex+1)
		}
		if retryCount < c.maxRetries {
			delay := c.baseDelay * time.Duration(1<<uint(retryCount))
			log.Printf("[Antigravity API] Rate limited during stream. Retrying in %v...", delay)
			time.Sleep(delay)
			return c.streamAPIWithRetry(method, body, responseChan, isRetry, retryCount+1, 0)
		}
		return fmt.Errorf("rate limit exceeded after %d retries", c.maxRetries)

	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		if retryCount < c.maxRetries {
			delay := c.baseDelay * time.Duration(1<<uint(retryCount))
			log.Printf("[Antigravity API] Server error %d during stream. Retrying in %v...", resp.StatusCode, delay)
			time.Sleep(delay)
			return c.streamAPIWithRetry(method, body, responseChan, isRetry, retryCount+1, baseURLIndex)
		}
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d after %d retries: %s", resp.StatusCode, c.maxRetries, string(body))

	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}

// parseSSEStream parses Server-Sent Events stream
func (c *APIClient) parseSSEStream(reader io.Reader, responseChan chan<- *models.GeminiResponse) error {
	scanner := bufio.NewScanner(reader)
	var buffer []string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			// Extract data after "data: " prefix
			data := strings.TrimPrefix(line, "data: ")
			buffer = append(buffer, data)
		} else if line == "" && len(buffer) > 0 {
			// Empty line indicates end of an event, parse the buffer
			dataStr := strings.Join(buffer, "\n")
			if err := c.parseAndSendChunk(dataStr, responseChan); err != nil {
				log.Printf("[Antigravity Stream] Failed to parse JSON chunk: %v", err)
			}
			buffer = nil
		}
	}

	// Parse any remaining buffer
	if len(buffer) > 0 {
		dataStr := strings.Join(buffer, "\n")
		if err := c.parseAndSendChunk(dataStr, responseChan); err != nil {
			log.Printf("[Antigravity Stream] Failed to parse final JSON chunk: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// parseAndSendChunk parses a JSON chunk and sends it to the response channel
func (c *APIClient) parseAndSendChunk(data string, responseChan chan<- *models.GeminiResponse) error {
	var antigravityResp models.AntigravityResponse
	if err := json.Unmarshal([]byte(data), &antigravityResp); err != nil {
		return err
	}

	if geminiResp := ToGeminiAPIResponse(&antigravityResp); geminiResp != nil {
		responseChan <- geminiResp
	}

	return nil
}
