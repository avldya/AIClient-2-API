package antigravity

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGenerateRequestID(t *testing.T) {
	id := GenerateRequestID()

	// Should start with "agent-"
	if !strings.HasPrefix(id, "agent-") {
		t.Errorf("Request ID should start with 'agent-', got: %s", id)
	}

	// Should be followed by a valid UUID
	uuidPart := strings.TrimPrefix(id, "agent-")
	if _, err := uuid.Parse(uuidPart); err != nil {
		t.Errorf("Request ID should contain valid UUID, got: %s", uuidPart)
	}
}

func TestGenerateSessionID(t *testing.T) {
	id := GenerateSessionID()

	// Should start with "-"
	if !strings.HasPrefix(id, "-") {
		t.Errorf("Session ID should start with '-', got: %s", id)
	}

	// Should be followed by a number
	numPart := strings.TrimPrefix(id, "-")
	if len(numPart) < 1 {
		t.Errorf("Session ID should have numeric part, got: %s", id)
	}
}

func TestGenerateProjectID(t *testing.T) {
	id := GenerateProjectID()

	// Should have format {adj}-{noun}-{random5}
	parts := strings.Split(id, "-")
	if len(parts) != 3 {
		t.Errorf("Project ID should have 3 parts separated by '-', got: %s", id)
	}

	// Last part should be 5 characters
	if len(parts[2]) != 5 {
		t.Errorf("Project ID should have 5-character random suffix, got: %s", parts[2])
	}
}

func TestAlias2ModelName(t *testing.T) {
	tests := []struct {
		alias    string
		expected string
	}{
		{"gemini-3-pro-preview", "gemini-3-pro-high"},
		{"gemini-3-flash-preview", "gemini-3-flash"},
		{"gemini-claude-sonnet-4-5", "claude-sonnet-4-5"},
		{"unknown-model", "unknown-model"},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			result := Alias2ModelName(tt.alias)
			if result != tt.expected {
				t.Errorf("Alias2ModelName(%s) = %s, want %s", tt.alias, result, tt.expected)
			}
		})
	}
}

func TestModelName2Alias(t *testing.T) {
	tests := []struct {
		modelName string
		expected  string
	}{
		{"gemini-3-pro-high", "gemini-3-pro-preview"},
		{"gemini-3-flash", "gemini-3-flash-preview"},
		{"claude-sonnet-4-5", "gemini-claude-sonnet-4-5"},
		{"unknown-model", "unknown-model"},
	}

	for _, tt := range tests {
		t.Run(tt.modelName, func(t *testing.T) {
			result := ModelName2Alias(tt.modelName)
			if result != tt.expected {
				t.Errorf("ModelName2Alias(%s) = %s, want %s", tt.modelName, result, tt.expected)
			}
		})
	}
}

func TestBidirectionalMapping(t *testing.T) {
	// Test that alias -> model -> alias works correctly
	aliases := []string{
		"gemini-3-pro-preview",
		"gemini-3-flash-preview",
		"gemini-claude-sonnet-4-5",
	}

	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			modelName := Alias2ModelName(alias)
			backToAlias := ModelName2Alias(modelName)
			if backToAlias != alias {
				t.Errorf("Bidirectional mapping failed: %s -> %s -> %s", alias, modelName, backToAlias)
			}
		})
	}
}
