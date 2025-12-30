package antigravity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// APIClient handles API requests to Antigravity endpoints
type APIClient struct {
	authManager *AuthManager
	baseURLs    []string
	userAgent   string
	maxRetries  int
	baseDelay   time.Duration
	httpClient  *http.Client
}

// NewAPIClient creates a new APIClient
func NewAPIClient(authManager *AuthManager, config map[string]interface{}) *APIClient {
	baseURLDaily := BaseURLDaily
	baseURLAutopush := BaseURLAutopush
	userAgent := DefaultUserAgent
	maxRetries := DefaultMaxRetries
	baseDelay := DefaultBaseDelay

	if val, ok := config["baseUrlDaily"].(string); ok && val != "" {
		baseURLDaily = val
	}
	if val, ok := config["baseUrlAutopush"].(string); ok && val != "" {
		baseURLAutopush = val
	}
	if val, ok := config["userAgent"].(string); ok && val != "" {
		userAgent = val
	}
	if val, ok := config["maxRetries"].(float64); ok {
		maxRetries = int(val)
	}
	if val, ok := config["baseDelay"].(float64); ok {
		baseDelay = int(val)
	}

	return &APIClient{
		authManager: authManager,
		baseURLs:    []string{baseURLDaily, baseURLAutopush},
		userAgent:   userAgent,
		maxRetries:  maxRetries,
		baseDelay:   time.Duration(baseDelay) * time.Millisecond,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// CallAPI makes an API call to Antigravity
func (c *APIClient) CallAPI(method string, body interface{}, response interface{}) error {
	return c.callAPIWithRetry(method, body, response, false, 0, 0)
}

// callAPIWithRetry makes an API call with retry logic
func (c *APIClient) callAPIWithRetry(method string, body interface{}, response interface{}, isRetry bool, retryCount int, baseURLIndex int) error {
	if baseURLIndex >= len(c.baseURLs) {
		return fmt.Errorf("all Antigravity base URLs failed")
	}

	baseURL := c.baseURLs[baseURLIndex]
	url := fmt.Sprintf("%s/%s:%s", baseURL, APIVersion, method)

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

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[Antigravity API] Network error on %s: %v", baseURL, err)
		// Try next base URL on network error
		if baseURLIndex+1 < len(c.baseURLs) {
			log.Printf("[Antigravity API] Trying next base URL...")
			return c.callAPIWithRetry(method, body, response, isRetry, retryCount, baseURLIndex+1)
		}
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle different status codes
	switch resp.StatusCode {
	case http.StatusOK:
		if err := json.Unmarshal(respBody, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return nil

	case http.StatusBadRequest, http.StatusUnauthorized:
		if !isRetry {
			log.Println("[Antigravity API] Received 401/400. Refreshing auth and retrying...")
			if err := c.authManager.Initialize(true); err != nil {
				return fmt.Errorf("failed to refresh auth: %w", err)
			}
			return c.callAPIWithRetry(method, body, response, true, retryCount, baseURLIndex)
		}
		return fmt.Errorf("authentication failed: %s", string(respBody))

	case http.StatusTooManyRequests:
		// Try next base URL
		if baseURLIndex+1 < len(c.baseURLs) {
			log.Printf("[Antigravity API] Rate limited on %s. Trying next base URL...", baseURL)
			return c.callAPIWithRetry(method, body, response, isRetry, retryCount, baseURLIndex+1)
		}
		// Exponential backoff
		if retryCount < c.maxRetries {
			delay := c.baseDelay * time.Duration(1<<uint(retryCount))
			log.Printf("[Antigravity API] Rate limited. Retrying in %v...", delay)
			time.Sleep(delay)
			return c.callAPIWithRetry(method, body, response, isRetry, retryCount+1, 0)
		}
		return fmt.Errorf("rate limit exceeded after %d retries", c.maxRetries)

	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		if retryCount < c.maxRetries {
			delay := c.baseDelay * time.Duration(1<<uint(retryCount))
			log.Printf("[Antigravity API] Server error %d. Retrying in %v...", resp.StatusCode, delay)
			time.Sleep(delay)
			return c.callAPIWithRetry(method, body, response, isRetry, retryCount+1, baseURLIndex)
		}
		return fmt.Errorf("server error %d after %d retries: %s", resp.StatusCode, c.maxRetries, string(respBody))

	default:
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}
}

// FetchAvailableModels fetches available models from the API
func (c *APIClient) FetchAvailableModels() ([]string, error) {
	for _, baseURL := range c.baseURLs {
		url := fmt.Sprintf("%s/%s:fetchAvailableModels", baseURL, APIVersion)

		accessToken, err := c.authManager.GetAccessToken()
		if err != nil {
			continue
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader([]byte("{}")))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		req.Header.Set("User-Agent", c.userAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Printf("[Antigravity] Failed to fetch models from %s: %v", baseURL, err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			models := ParseModelsFromResponse(result)
			log.Printf("[Antigravity] Available models: [%d models found]", len(models))
			return models, nil
		}
		resp.Body.Close()
	}

	return nil, fmt.Errorf("failed to fetch models from all endpoints")
}
