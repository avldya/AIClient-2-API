package kiro

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/avldya/AIClient-2-API/go-kiro-api/pkg/models"
)

// AuthManager handles authentication and token management
// 认证管理器
type AuthManager struct {
	credentials     *models.Credentials
	base64Creds     string
	credsFilePath   string
	credPath        string
	refreshURL      string
	refreshIDCURL   string
	region          string
	cronNearMinutes int
	httpClient      *HTTPClient
}

// NewAuthManager creates a new authentication manager
// 创建新的认证管理器
func NewAuthManager(config map[string]interface{}, httpClient *HTTPClient) *AuthManager {
	am := &AuthManager{
		credentials:     &models.Credentials{},
		cronNearMinutes: DefaultCronNearMinutes,
		httpClient:      httpClient,
	}

	// Load configuration
	if val, ok := config["credsBase64"].(string); ok {
		am.base64Creds = val
	}
	if val, ok := config["credsFilePath"].(string); ok {
		am.credsFilePath = val
	}
	if val, ok := config["credPath"].(string); ok {
		am.credPath = val
	} else {
		// Default to user's home directory
		if home, err := os.UserHomeDir(); err == nil {
			am.credPath = filepath.Join(home, ".aws", "sso", "cache")
		}
	}
	if val, ok := config["region"].(string); ok {
		am.region = val
	} else {
		am.region = DefaultRegion
	}
	if val, ok := config["cronNearMinutes"].(int); ok {
		am.cronNearMinutes = val
	}

	return am
}

// Initialize loads and validates credentials
// 初始化并加载凭据
func (am *AuthManager) Initialize(forceRefresh bool) error {
	if am.credentials.AccessToken != "" && !forceRefresh {
		return nil
	}

	mergedCreds := make(map[string]interface{})

	// Priority 1: Load from Base64 credentials
	if am.base64Creds != "" {
		decoded, err := base64.StdEncoding.DecodeString(am.base64Creds)
		if err != nil {
			return fmt.Errorf("failed to decode base64 credentials: %w", err)
		}
		var creds map[string]interface{}
		if err := json.Unmarshal(decoded, &creds); err != nil {
			return fmt.Errorf("failed to parse base64 credentials: %w", err)
		}
		for k, v := range creds {
			mergedCreds[k] = v
		}
		am.base64Creds = "" // Clear after use
	}

	// Priority 2 & 3: Load from file path or directory
	targetFilePath := am.credsFilePath
	if targetFilePath == "" {
		targetFilePath = filepath.Join(am.credPath, KiroAuthTokenFile)
	}

	dirPath := filepath.Dir(targetFilePath)
	targetFileName := filepath.Base(targetFilePath)

	// Try to read target file first
	if creds, err := am.loadCredentialsFromFile(targetFilePath); err == nil && creds != nil {
		for k, v := range creds {
			mergedCreds[k] = v
		}
	}

	// Then read other JSON files in the directory (excluding target file)
	if entries, err := os.ReadDir(dirPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") && entry.Name() != targetFileName {
				filePath := filepath.Join(dirPath, entry.Name())
				if creds, err := am.loadCredentialsFromFile(filePath); err == nil && creds != nil {
					// Preserve existing expiresAt
					if expiresAt, ok := mergedCreds["expiresAt"]; ok {
						creds["expiresAt"] = expiresAt
					}
					for k, v := range creds {
						mergedCreds[k] = v
					}
				}
			}
		}
	}

	// Apply merged credentials to the struct
	am.applyCredentials(mergedCreds)

	// Ensure region is set
	if am.credentials.Region != "" {
		am.region = am.credentials.Region
	}
	if am.region == "" {
		am.region = DefaultRegion
	}

	// Set URLs with region
	am.refreshURL = ReplaceRegion(RefreshURL, am.region)
	am.refreshIDCURL = ReplaceRegion(RefreshIDCURL, am.region)

	// Refresh token if forced or if access token is missing but refresh token is available
	if forceRefresh || (am.credentials.AccessToken == "" && am.credentials.RefreshToken != "") {
		if err := am.refreshToken(); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	if am.credentials.AccessToken == "" {
		return fmt.Errorf("no access token available after initialization")
	}

	return nil
}

// loadCredentialsFromFile loads credentials from a JSON file
// 从 JSON 文件加载凭据
func (am *AuthManager) loadCredentialsFromFile(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var creds map[string]interface{}
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	return creds, nil
}

// applyCredentials applies loaded credentials to the credentials struct
// 应用加载的凭据到结构体
func (am *AuthManager) applyCredentials(creds map[string]interface{}) {
	if val, ok := creds["accessToken"].(string); ok && am.credentials.AccessToken == "" {
		am.credentials.AccessToken = val
	}
	if val, ok := creds["refreshToken"].(string); ok && am.credentials.RefreshToken == "" {
		am.credentials.RefreshToken = val
	}
	if val, ok := creds["clientId"].(string); ok && am.credentials.ClientID == "" {
		am.credentials.ClientID = val
	}
	if val, ok := creds["clientSecret"].(string); ok && am.credentials.ClientSecret == "" {
		am.credentials.ClientSecret = val
	}
	if val, ok := creds["authMethod"].(string); ok && am.credentials.AuthMethod == "" {
		am.credentials.AuthMethod = val
	}
	if val, ok := creds["expiresAt"].(string); ok && am.credentials.ExpiresAt == "" {
		am.credentials.ExpiresAt = val
	}
	if val, ok := creds["profileArn"].(string); ok && am.credentials.ProfileArn == "" {
		am.credentials.ProfileArn = val
	}
	if val, ok := creds["region"].(string); ok && am.credentials.Region == "" {
		am.credentials.Region = val
	}
	if val, ok := creds["uuid"].(string); ok && am.credentials.UUID == "" {
		am.credentials.UUID = val
	}
}

// refreshToken refreshes the access token using refresh token
// 使用 refresh token 刷新 access token
func (am *AuthManager) refreshToken() error {
	if am.credentials.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	requestBody := map[string]string{
		"refreshToken": am.credentials.RefreshToken,
	}

	refreshURL := am.refreshURL
	if am.credentials.AuthMethod != AuthMethodSocial {
		refreshURL = am.refreshIDCURL
		requestBody["clientId"] = am.credentials.ClientID
		requestBody["clientSecret"] = am.credentials.ClientSecret
		requestBody["grantType"] = "refresh_token"
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := am.httpClient.Post(refreshURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	// Update credentials
	if accessToken, ok := result["accessToken"].(string); ok {
		am.credentials.AccessToken = accessToken
	} else {
		return fmt.Errorf("invalid refresh response: missing accessToken")
	}

	if refreshToken, ok := result["refreshToken"].(string); ok {
		am.credentials.RefreshToken = refreshToken
	}

	if profileArn, ok := result["profileArn"].(string); ok {
		am.credentials.ProfileArn = profileArn
	}

	if expiresIn, ok := result["expiresIn"].(float64); ok {
		expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
		am.credentials.ExpiresAt = expiresAt.Format(time.RFC3339)
	}

	// Save updated credentials to file
	if err := am.saveCredentials(); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return nil
}

// saveCredentials saves updated credentials to file
// 保存更新的凭据到文件
func (am *AuthManager) saveCredentials() error {
	targetFilePath := am.credsFilePath
	if targetFilePath == "" {
		targetFilePath = filepath.Join(am.credPath, KiroAuthTokenFile)
	}

	// Read existing file if it exists
	existingData := make(map[string]interface{})
	if data, err := os.ReadFile(targetFilePath); err == nil {
		json.Unmarshal(data, &existingData)
	}

	// Update with new token data
	existingData["accessToken"] = am.credentials.AccessToken
	existingData["refreshToken"] = am.credentials.RefreshToken
	existingData["expiresAt"] = am.credentials.ExpiresAt
	if am.credentials.ProfileArn != "" {
		existingData["profileArn"] = am.credentials.ProfileArn
	}

	// Write back to file
	jsonData, err := json.MarshalIndent(existingData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(targetFilePath, jsonData, 0600)
}

// IsExpiryNear checks if the token is near expiry
// 检查 token 是否即将过期
func (am *AuthManager) IsExpiryNear() bool {
	if am.credentials.ExpiresAt == "" {
		return true
	}

	expiresAt, err := time.Parse(time.RFC3339, am.credentials.ExpiresAt)
	if err != nil {
		return true
	}

	nearTime := time.Now().Add(time.Duration(am.cronNearMinutes) * time.Minute)
	return expiresAt.Before(nearTime)
}

// GetAccessToken returns the current access token
// 获取当前的 access token
func (am *AuthManager) GetAccessToken() string {
	return am.credentials.AccessToken
}

// GetCredentials returns the credentials
// 获取凭据
func (am *AuthManager) GetCredentials() *models.Credentials {
	return am.credentials
}
