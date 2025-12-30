package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the application configuration
type Config struct {
	Server      ServerConfig      `json:"server"`
	Antigravity AntigravityConfig `json:"antigravity"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	APIKey string `json:"apiKey"`
}

// AntigravityConfig represents Antigravity service configuration
type AntigravityConfig struct {
	OAuthCredsFilePath string `json:"oauthCredsFilePath"`
	ProjectID          string `json:"projectId"`
	BaseURLDaily       string `json:"baseUrlDaily"`
	BaseURLAutopush    string `json:"baseUrlAutopush"`
	UserAgent          string `json:"userAgent"`
	RequestMaxRetries  int    `json:"requestMaxRetries"`
	RequestBaseDelay   int    `json:"requestBaseDelay"`
	CronNearMinutes    int    `json:"cronNearMinutes"`
}

// Load loads configuration from a file
func Load(configPath string) (*Config, error) {
	// Set defaults
	config := &Config{
		Server: ServerConfig{
			Host:   "0.0.0.0",
			Port:   3000,
			APIKey: "",
		},
		Antigravity: AntigravityConfig{
			OAuthCredsFilePath: "",
			ProjectID:          "",
			BaseURLDaily:       "https://daily-cloudcode-pa.sandbox.googleapis.com",
			BaseURLAutopush:    "https://autopush-cloudcode-pa.sandbox.googleapis.com",
			UserAgent:          "antigravity/1.11.5 windows/amd64",
			RequestMaxRetries:  3,
			RequestBaseDelay:   1000,
			CronNearMinutes:    50,
		},
	}

	// If no config file specified, return defaults
	if configPath == "" {
		return config, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	config := &Config{
		Server: ServerConfig{
			Host:   getEnv("SERVER_HOST", "0.0.0.0"),
			Port:   getEnvInt("SERVER_PORT", 3000),
			APIKey: getEnv("SERVER_API_KEY", ""),
		},
		Antigravity: AntigravityConfig{
			OAuthCredsFilePath: getEnv("ANTIGRAVITY_CREDS_FILE", ""),
			ProjectID:          getEnv("ANTIGRAVITY_PROJECT_ID", ""),
			BaseURLDaily:       getEnv("ANTIGRAVITY_BASE_URL_DAILY", "https://daily-cloudcode-pa.sandbox.googleapis.com"),
			BaseURLAutopush:    getEnv("ANTIGRAVITY_BASE_URL_AUTOPUSH", "https://autopush-cloudcode-pa.sandbox.googleapis.com"),
			UserAgent:          getEnv("ANTIGRAVITY_USER_AGENT", "antigravity/1.11.5 windows/amd64"),
			RequestMaxRetries:  getEnvInt("ANTIGRAVITY_MAX_RETRIES", 3),
			RequestBaseDelay:   getEnvInt("ANTIGRAVITY_BASE_DELAY", 1000),
			CronNearMinutes:    getEnvInt("ANTIGRAVITY_CRON_NEAR_MINUTES", 50),
		},
	}

	return config
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// ToMap converts config to a map for use with the antigravity package
func (c *Config) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"oauthCredsFilePath": c.Antigravity.OAuthCredsFilePath,
		"projectId":          c.Antigravity.ProjectID,
		"baseUrlDaily":       c.Antigravity.BaseURLDaily,
		"baseUrlAutopush":    c.Antigravity.BaseURLAutopush,
		"userAgent":          c.Antigravity.UserAgent,
		"maxRetries":         float64(c.Antigravity.RequestMaxRetries),
		"baseDelay":          float64(c.Antigravity.RequestBaseDelay),
	}
}
