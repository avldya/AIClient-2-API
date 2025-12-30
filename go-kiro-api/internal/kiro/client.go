package kiro

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/avldya/AIClient-2-API/go-kiro-api/pkg/models"
)

// HTTPClient wraps http.Client with custom configuration
// HTTP 客户端包装器
type HTTPClient struct {
	client      *http.Client
	proxyURL    *url.URL
	timeout     time.Duration
	maxRetries  int
	baseDelay   time.Duration
}

// NewHTTPClient creates a new HTTP client
// 创建新的 HTTP 客户端
func NewHTTPClient(config map[string]interface{}) *HTTPClient {
	timeout := AxiosTimeout
	if val, ok := config["timeout"].(int); ok {
		timeout = val
	}

	maxRetries := DefaultMaxRetries
	if val, ok := config["requestMaxRetries"].(int); ok {
		maxRetries = val
	}

	baseDelay := DefaultBaseDelay
	if val, ok := config["requestBaseDelay"].(int); ok {
		baseDelay = val
	}

	hc := &HTTPClient{
		timeout:    time.Duration(timeout) * time.Millisecond,
		maxRetries: maxRetries,
		baseDelay:  time.Duration(baseDelay) * time.Millisecond,
	}

	// Set up proxy if configured
	if proxyStr, ok := config["proxy"].(string); ok && proxyStr != "" {
		if proxyURL, err := url.Parse(proxyStr); err == nil {
			hc.proxyURL = proxyURL
		}
	}

	// Create HTTP client with custom transport
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	if hc.proxyURL != nil {
		transport.Proxy = http.ProxyURL(hc.proxyURL)
	}

	hc.client = &http.Client{
		Transport: transport,
		Timeout:   hc.timeout,
	}

	return hc
}

// Post performs a POST request
// 执行 POST 请求
func (hc *HTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	return hc.client.Do(req)
}

// Client represents the Kiro API client
// Kiro API 客户端
type Client struct {
	authManager  *AuthManager
	converter    *Converter
	streamParser *StreamParser
	httpClient   *HTTPClient
	baseURL      string
	amazonQURL   string
	config       map[string]interface{}
	initialized  bool
}

// NewClient creates a new Kiro API client
// 创建新的 Kiro API 客户端
func NewClient(config map[string]interface{}) *Client {
	httpClient := NewHTTPClient(config)

	c := &Client{
		converter:    NewConverter(),
		streamParser: NewStreamParser(),
		httpClient:   httpClient,
		config:       config,
	}

	c.authManager = NewAuthManager(config, httpClient)

	return c
}

// Initialize initializes the client
// 初始化客户端
func (c *Client) Initialize() error {
	if c.initialized {
		return nil
	}

	// Initialize authentication
	if err := c.authManager.Initialize(false); err != nil {
		return fmt.Errorf("authentication initialization failed: %w", err)
	}

	// Set up URLs
	region := c.authManager.region
	c.baseURL = ReplaceRegion(BaseURL, region)
	c.amazonQURL = ReplaceRegion(AmazonQURL, region)

	c.initialized = true
	return nil
}

// GenerateContent generates content (non-streaming)
// 生成内容（非流式）
func (c *Client) GenerateContent(req *models.ChatCompletionRequest) (interface{}, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return nil, err
		}
	}

	// Check if token needs refresh
	if c.authManager.IsExpiryNear() {
		if err := c.authManager.Initialize(true); err != nil {
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}
	}

	// Convert request format
	cwReq, err := c.converter.BuildCodeWhispererRequest(req.Messages, req.Model, req.Tools, req.System)
	if err != nil {
		return nil, fmt.Errorf("request conversion failed: %w", err)
	}

	// Make API call with retry
	response, err := c.makeRequestWithRetry(req.Model, cwReq, false)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GenerateContentStream generates content (streaming)
// 生成内容（流式）
func (c *Client) GenerateContentStream(req *models.ChatCompletionRequest) (<-chan interface{}, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return nil, err
		}
	}

	// Check if token needs refresh
	if c.authManager.IsExpiryNear() {
		if err := c.authManager.Initialize(true); err != nil {
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}
	}

	// Convert request format
	cwReq, err := c.converter.BuildCodeWhispererRequest(req.Messages, req.Model, req.Tools, req.System)
	if err != nil {
		return nil, fmt.Errorf("request conversion failed: %w", err)
	}

	// Make streaming API call
	return c.makeStreamingRequest(req.Model, cwReq)
}

// makeRequestWithRetry makes an API request with retry logic
// 执行带重试逻辑的 API 请求
func (c *Client) makeRequestWithRetry(model string, cwReq *models.CodeWhispererRequest, isRetry bool) (interface{}, error) {
	var lastErr error

	for attempt := 0; attempt <= c.httpClient.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * c.httpClient.baseDelay
			time.Sleep(delay)
		}

		response, err := c.makeRequest(model, cwReq)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Check if we should retry
		if !c.shouldRetry(err) {
			break
		}

		// If 403, try refreshing token once
		if attempt == 0 && c.is403Error(err) {
			if refreshErr := c.authManager.Initialize(true); refreshErr == nil {
				continue
			}
		}
	}

	return nil, lastErr
}

// makeRequest makes a single API request
// 执行单个 API 请求
func (c *Client) makeRequest(model string, cwReq *models.CodeWhispererRequest) (interface{}, error) {
	requestURL := c.baseURL
	if len(model) > 7 && model[:7] == "amazonq" {
		requestURL = c.amazonQURL
	}

	jsonData, err := json.Marshal(cwReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set headers
	c.setRequestHeaders(req)

	resp, err := c.httpClient.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to appropriate format
	return result, nil
}

// makeStreamingRequest makes a streaming API request
// 执行流式 API 请求
func (c *Client) makeStreamingRequest(model string, cwReq *models.CodeWhispererRequest) (<-chan interface{}, error) {
	requestURL := c.baseURL
	if len(model) > 7 && model[:7] == "amazonq" {
		requestURL = c.amazonQURL
	}

	jsonData, err := json.Marshal(cwReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set headers
	c.setRequestHeaders(req)

	resp, err := c.httpClient.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	// Parse stream using stream parser
	return c.streamParser.ParseAWSEventStream(resp.Body, model), nil
}

// setRequestHeaders sets common request headers
// 设置请求头
func (c *Client) setRequestHeaders(req *http.Request) {
	token := c.authManager.GetAccessToken()
	sysInfo := GetSystemRuntimeInfo()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", ContentTypeJSON)
	req.Header.Set("Accept", AcceptJSON)
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s %s", UserAgent, KiroVersion, sysInfo["osName"]))
	req.Header.Set("amz-sdk-invocation-id", GenerateUUID())
}

// shouldRetry determines if a request should be retried
// 判断是否应该重试
func (c *Client) shouldRetry(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		// Retry on 429 (rate limit) and 5xx errors
		return apiErr.StatusCode == 429 || apiErr.StatusCode >= 500
	}
	return false
}

// is403Error checks if error is a 403 error
// 检查是否为 403 错误
func (c *Client) is403Error(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 403
	}
	return false
}

// GetUsageLimits retrieves usage limits information
// 获取用量限制信息
func (c *Client) GetUsageLimits() (map[string]interface{}, error) {
	if !c.initialized {
		if err := c.Initialize(); err != nil {
			return nil, err
		}
	}

	usageLimitsURL := ReplaceRegion(UsageLimitsURL, c.authManager.region)

	req, err := http.NewRequest("POST", usageLimitsURL, nil)
	if err != nil {
		return nil, err
	}

	c.setRequestHeaders(req)

	resp, err := c.httpClient.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("usage limits request failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// APIError represents an API error
// API 错误
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// ListModels returns the list of supported models
// 返回支持的模型列表
func (c *Client) ListModels() []string {
	return SupportedModels
}

// GetStreamParser returns the stream parser
// 获取流解析器
func (c *Client) GetStreamParser() *StreamParser {
	return c.streamParser
}
