package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("SERVER_API_KEY", "test-key")
	os.Setenv("ANTIGRAVITY_PROJECT_ID", "test-project")
	
	defer func() {
		// Cleanup
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_API_KEY")
		os.Unsetenv("ANTIGRAVITY_PROJECT_ID")
	}()

	cfg := LoadFromEnv()

	// Test server config
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %s, want 127.0.0.1", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}

	if cfg.Server.APIKey != "test-key" {
		t.Errorf("Server.APIKey = %s, want test-key", cfg.Server.APIKey)
	}

	// Test antigravity config
	if cfg.Antigravity.ProjectID != "test-project" {
		t.Errorf("Antigravity.ProjectID = %s, want test-project", cfg.Antigravity.ProjectID)
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg := LoadFromEnv()

	// Test default values
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Default Server.Host = %s, want 0.0.0.0", cfg.Server.Host)
	}

	if cfg.Server.Port != 3000 {
		t.Errorf("Default Server.Port = %d, want 3000", cfg.Server.Port)
	}

	if cfg.Antigravity.RequestMaxRetries != 3 {
		t.Errorf("Default RequestMaxRetries = %d, want 3", cfg.Antigravity.RequestMaxRetries)
	}

	if cfg.Antigravity.RequestBaseDelay != 1000 {
		t.Errorf("Default RequestBaseDelay = %d, want 1000", cfg.Antigravity.RequestBaseDelay)
	}

	if cfg.Antigravity.UserAgent != "antigravity/1.11.5 windows/amd64" {
		t.Errorf("Default UserAgent = %s, want 'antigravity/1.11.5 windows/amd64'", cfg.Antigravity.UserAgent)
	}
}

func TestToMap(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host:   "localhost",
			Port:   3000,
			APIKey: "test",
		},
		Antigravity: AntigravityConfig{
			ProjectID:         "my-project",
			RequestMaxRetries: 5,
			RequestBaseDelay:  2000,
			UserAgent:         "test-agent",
		},
	}

	m := cfg.ToMap()

	// Check map contents
	if m["projectId"] != "my-project" {
		t.Errorf("Map projectId = %v, want 'my-project'", m["projectId"])
	}

	if m["maxRetries"] != float64(5) {
		t.Errorf("Map maxRetries = %v, want 5", m["maxRetries"])
	}

	if m["baseDelay"] != float64(2000) {
		t.Errorf("Map baseDelay = %v, want 2000", m["baseDelay"])
	}

	if m["userAgent"] != "test-agent" {
		t.Errorf("Map userAgent = %v, want 'test-agent'", m["userAgent"])
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "Environment variable set",
			key:          "TEST_VAR_1",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "Environment variable not set",
			key:          "TEST_VAR_2",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%s, %s) = %s, want %s", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "Valid integer",
			key:          "TEST_INT_1",
			defaultValue: 100,
			envValue:     "200",
			expected:     200,
		},
		{
			name:         "Invalid integer",
			key:          "TEST_INT_2",
			defaultValue: 100,
			envValue:     "not-a-number",
			expected:     100,
		},
		{
			name:         "Not set",
			key:          "TEST_INT_3",
			defaultValue: 100,
			envValue:     "",
			expected:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvInt(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvInt(%s, %d) = %d, want %d", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
