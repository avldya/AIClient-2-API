package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// Config represents the application configuration
// 应用配置
type Config struct {
	Server ServerConfig `json:"server"`
	Kiro   KiroConfig   `json:"kiro"`
}

// ServerConfig represents server configuration
// 服务器配置
type ServerConfig struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	APIKey string `json:"apiKey"`
}

// KiroConfig represents Kiro-specific configuration
// Kiro 配置
type KiroConfig struct {
	CredPath          string `json:"credPath"`
	CredsBase64       string `json:"credsBase64"`
	CredsFilePath     string `json:"credsFilePath"`
	Region            string `json:"region"`
	UseSystemProxy    bool   `json:"useSystemProxy"`
	Proxy             string `json:"proxy"`
	RequestMaxRetries int    `json:"requestMaxRetries"`
	RequestBaseDelay  int    `json:"requestBaseDelay"`
	CronNearMinutes   int    `json:"cronNearMinutes"`
	Timeout           int    `json:"timeout"`
}

// LoadConfig loads configuration from file and environment variables
// 从文件和环境变量加载配置
func LoadConfig(configFile string) (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host:   "0.0.0.0",
			Port:   3000,
			APIKey: "123456",
		},
		Kiro: KiroConfig{
			Region:            "us-east-1",
			RequestMaxRetries: 3,
			RequestBaseDelay:  1000,
			CronNearMinutes:   10,
			Timeout:           300000,
		},
	}

	// Load from config file if provided
	if configFile != "" {
		if err := loadFromFile(configFile, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	return config, nil
}

// loadFromFile loads configuration from a JSON file
// 从 JSON 文件加载配置
func loadFromFile(filename string, config *Config) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, config)
}

// loadFromEnv loads configuration from environment variables
// 从环境变量加载配置
func loadFromEnv(config *Config) {
	// Server configuration
	if val := os.Getenv("SERVER_HOST"); val != "" {
		config.Server.Host = val
	}
	if val := os.Getenv("SERVER_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.Server.Port = port
		}
	}
	if val := os.Getenv("SERVER_API_KEY"); val != "" {
		config.Server.APIKey = val
	}
	if val := os.Getenv("API_KEY"); val != "" {
		config.Server.APIKey = val
	}

	// Kiro configuration
	if val := os.Getenv("KIRO_CRED_PATH"); val != "" {
		config.Kiro.CredPath = val
	}
	if val := os.Getenv("KIRO_CREDS_BASE64"); val != "" {
		config.Kiro.CredsBase64 = val
	}
	if val := os.Getenv("KIRO_CREDS_FILE_PATH"); val != "" {
		config.Kiro.CredsFilePath = val
	}
	if val := os.Getenv("KIRO_REGION"); val != "" {
		config.Kiro.Region = val
	}
	if val := os.Getenv("KIRO_PROXY"); val != "" {
		config.Kiro.Proxy = val
	}
	if val := os.Getenv("KIRO_REQUEST_MAX_RETRIES"); val != "" {
		if retries, err := strconv.Atoi(val); err == nil {
			config.Kiro.RequestMaxRetries = retries
		}
	}
	if val := os.Getenv("KIRO_REQUEST_BASE_DELAY"); val != "" {
		if delay, err := strconv.Atoi(val); err == nil {
			config.Kiro.RequestBaseDelay = delay
		}
	}
	if val := os.Getenv("KIRO_CRON_NEAR_MINUTES"); val != "" {
		if minutes, err := strconv.Atoi(val); err == nil {
			config.Kiro.CronNearMinutes = minutes
		}
	}
}

// ToMap converts Config to a map for use with internal packages
// 将配置转换为 map 以供内部包使用
func (c *Config) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"credPath":          c.Kiro.CredPath,
		"credsBase64":       c.Kiro.CredsBase64,
		"credsFilePath":     c.Kiro.CredsFilePath,
		"region":            c.Kiro.Region,
		"proxy":             c.Kiro.Proxy,
		"requestMaxRetries": c.Kiro.RequestMaxRetries,
		"requestBaseDelay":  c.Kiro.RequestBaseDelay,
		"cronNearMinutes":   c.Kiro.CronNearMinutes,
		"timeout":           c.Kiro.Timeout,
	}
}
