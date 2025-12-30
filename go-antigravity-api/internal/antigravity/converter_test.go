package antigravity

import (
	"testing"

	"go-antigravity-api/pkg/models"
)

func TestEnsureRolesInContents(t *testing.T) {
	tests := []struct {
		name     string
		request  *models.GeminiRequest
		expected string
	}{
		{
			name: "SystemInstruction without role",
			request: &models.GeminiRequest{
				SystemInstruction: &models.Content{
					Parts: []models.Part{{Text: "You are a helpful assistant"}},
				},
			},
			expected: "user",
		},
		{
			name: "Contents without roles",
			request: &models.GeminiRequest{
				Contents: []models.Content{
					{Parts: []models.Part{{Text: "Hello"}}},
					{Parts: []models.Part{{Text: "Hi there"}}},
				},
			},
			expected: "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ensureRolesInContents(tt.request)

			if tt.request.SystemInstruction != nil && tt.request.SystemInstruction.Role != tt.expected {
				t.Errorf("SystemInstruction role = %s, want %s", tt.request.SystemInstruction.Role, tt.expected)
			}

			for i, content := range tt.request.Contents {
				if content.Role != tt.expected {
					t.Errorf("Content[%d] role = %s, want %s", i, content.Role, tt.expected)
				}
			}
		})
	}
}

func TestGeminiToAntigravity(t *testing.T) {
	projectID := "test-project-123"
	modelName := "gemini-3-pro-high"

	request := &models.GeminiRequest{
		Contents: []models.Content{
			{
				Role:  "user",
				Parts: []models.Part{{Text: "Hello"}},
			},
		},
		SafetySettings: []models.SafetySetting{
			{Category: "HARM_CATEGORY_DANGEROUS", Threshold: "BLOCK_NONE"},
		},
	}

	result := GeminiToAntigravity(modelName, request, projectID)

	// Check basic fields
	if result.Model != modelName {
		t.Errorf("Model = %s, want %s", result.Model, modelName)
	}

	if result.UserAgent != "antigravity" {
		t.Errorf("UserAgent = %s, want 'antigravity'", result.UserAgent)
	}

	if result.Project != projectID {
		t.Errorf("Project = %s, want %s", result.Project, projectID)
	}

	// Check that RequestID is generated
	if result.RequestID == "" || len(result.RequestID) < 10 {
		t.Errorf("RequestID should be generated, got: %s", result.RequestID)
	}

	// Check that SafetySettings are removed
	if result.Request.SafetySettings != nil {
		t.Errorf("SafetySettings should be removed")
	}

	// Check that contents are preserved
	if len(result.Request.Contents) != 1 {
		t.Errorf("Contents length = %d, want 1", len(result.Request.Contents))
	}
}

func TestGeminiToAntigravityWithThinking(t *testing.T) {
	projectID := "test-project"
	modelName := "gemini-2-pro" // Non-Gemini-3 model

	temp := 0.7
	thinkingBudget := 1000
	request := &models.GeminiRequest{
		Contents: []models.Content{
			{Role: "user", Parts: []models.Part{{Text: "Test"}}},
		},
		GenerationConfig: &models.GenerationConfig{
			Temperature: &temp,
			ThinkingConfig: &models.ThinkingConfig{
				ThinkingLevel:  "high",
				ThinkingBudget: thinkingBudget,
			},
		},
	}

	result := GeminiToAntigravity(modelName, request, projectID)

	// For non-Gemini-3 models, thinking level should be cleared and budget set to -1
	if result.Request.GenerationConfig.ThinkingConfig.ThinkingLevel != "" {
		t.Errorf("ThinkingLevel should be empty for non-Gemini-3 models, got: %s",
			result.Request.GenerationConfig.ThinkingConfig.ThinkingLevel)
	}

	if result.Request.GenerationConfig.ThinkingConfig.ThinkingBudget != -1 {
		t.Errorf("ThinkingBudget should be -1 for non-Gemini-3 models, got: %d",
			result.Request.GenerationConfig.ThinkingConfig.ThinkingBudget)
	}
}

func TestToGeminiAPIResponse(t *testing.T) {
	// Test with nil
	if result := ToGeminiAPIResponse(nil); result != nil {
		t.Errorf("ToGeminiAPIResponse(nil) should return nil")
	}

	// Test with empty response
	antigravityResp := &models.AntigravityResponse{}
	if result := ToGeminiAPIResponse(antigravityResp); result != nil {
		t.Errorf("ToGeminiAPIResponse with nil response should return nil")
	}

	// Test with valid response
	geminiResp := &models.GeminiResponse{
		Candidates: []models.Candidate{
			{
				Content: &models.Content{
					Role:  "model",
					Parts: []models.Part{{Text: "Hello!"}},
				},
			},
		},
	}
	antigravityResp = &models.AntigravityResponse{
		Response: geminiResp,
	}

	result := ToGeminiAPIResponse(antigravityResp)
	if result == nil {
		t.Fatal("ToGeminiAPIResponse should not return nil")
	}

	if len(result.Candidates) != 1 {
		t.Errorf("Candidates length = %d, want 1", len(result.Candidates))
	}
}

func TestDeepCopyGeminiRequest(t *testing.T) {
	temp := 0.9
	original := &models.GeminiRequest{
		Contents: []models.Content{
			{Role: "user", Parts: []models.Part{{Text: "Test"}}},
		},
		GenerationConfig: &models.GenerationConfig{
			Temperature: &temp,
		},
	}

	copy := deepCopyGeminiRequest(original)

	// Modify the copy
	*copy.GenerationConfig.Temperature = 0.5
	copy.Contents[0].Role = "system"

	// Original should be unchanged
	if *original.GenerationConfig.Temperature != 0.9 {
		t.Errorf("Original was modified during copy")
	}

	if original.Contents[0].Role != "user" {
		t.Errorf("Original contents were modified during copy")
	}
}
