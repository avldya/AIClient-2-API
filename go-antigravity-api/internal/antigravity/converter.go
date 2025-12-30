package antigravity

import (
	"encoding/json"
	"strings"

	"go-antigravity-api/pkg/models"
)

// GeminiToAntigravity converts Gemini format request to Antigravity format
func GeminiToAntigravity(modelName string, payload *models.GeminiRequest, projectID string) *models.AntigravityRequest {
	// Deep copy the request to avoid modifying the original
	requestCopy := deepCopyGeminiRequest(payload)

	// Ensure roles are set in contents
	ensureRolesInContents(requestCopy)

	// Create Antigravity request
	antigravityReq := &models.AntigravityRequest{
		Model:     modelName,
		UserAgent: "antigravity",
		Project:   projectID,
		RequestID: GenerateRequestID(),
		Request:   requestCopy,
	}

	// Session ID is generated as part of the request
	// It's included in the request metadata when sending to the API

	// Delete safety settings
	antigravityReq.Request.SafetySettings = nil

	// Set tool config
	if antigravityReq.Request.ToolConfig != nil {
		if antigravityReq.Request.ToolConfig.FunctionCallingConfig == nil {
			antigravityReq.Request.ToolConfig.FunctionCallingConfig = &models.FunctionCallingConfig{}
		}
		antigravityReq.Request.ToolConfig.FunctionCallingConfig.Mode = "VALIDATED"
	}

	// Delete maxOutputTokens
	if antigravityReq.Request.GenerationConfig != nil {
		antigravityReq.Request.GenerationConfig.MaxOutputTokens = nil
	}

	// Handle Thinking config
	if !strings.HasPrefix(modelName, "gemini-3-") {
		if antigravityReq.Request.GenerationConfig != nil &&
			antigravityReq.Request.GenerationConfig.ThinkingConfig != nil &&
			antigravityReq.Request.GenerationConfig.ThinkingConfig.ThinkingLevel != "" {
			antigravityReq.Request.GenerationConfig.ThinkingConfig.ThinkingLevel = ""
			antigravityReq.Request.GenerationConfig.ThinkingConfig.ThinkingBudget = -1
		}
	}

	// Handle Claude model tool declarations
	if strings.HasPrefix(modelName, "claude-sonnet-") || strings.HasPrefix(modelName, "claude-opus-") {
		if antigravityReq.Request.Tools != nil {
			for i := range antigravityReq.Request.Tools {
				if antigravityReq.Request.Tools[i].FunctionDeclarations != nil {
					for j := range antigravityReq.Request.Tools[i].FunctionDeclarations {
						funcDecl := &antigravityReq.Request.Tools[i].FunctionDeclarations[j]
						if funcDecl.ParametersJSONSchema != nil && len(funcDecl.ParametersJSONSchema) > 0 {
							funcDecl.Parameters = funcDecl.ParametersJSONSchema
							// Remove $schema field if it exists
							delete(funcDecl.Parameters, "$schema")
							funcDecl.ParametersJSONSchema = nil
						}
					}
				}
			}
		}
	}

	return antigravityReq
}

// ToGeminiAPIResponse converts Antigravity response to Gemini format
func ToGeminiAPIResponse(antigravityResponse *models.AntigravityResponse) *models.GeminiResponse {
	if antigravityResponse == nil || antigravityResponse.Response == nil {
		return nil
	}

	return antigravityResponse.Response
}

// ensureRolesInContents ensures all content parts have roles
func ensureRolesInContents(requestBody *models.GeminiRequest) {
	// Set default role for systemInstruction
	if requestBody.SystemInstruction != nil && requestBody.SystemInstruction.Role == "" {
		requestBody.SystemInstruction.Role = "user"
	}

	// Set default role for contents
	if requestBody.Contents != nil {
		for i := range requestBody.Contents {
			if requestBody.Contents[i].Role == "" {
				requestBody.Contents[i].Role = "user"
			}
		}
	}
}

// deepCopyGeminiRequest creates a deep copy of GeminiRequest
func deepCopyGeminiRequest(req *models.GeminiRequest) *models.GeminiRequest {
	// Use JSON marshal/unmarshal for deep copy
	data, err := json.Marshal(req)
	if err != nil {
		// Fallback to shallow copy if marshaling fails
		copy := *req
		return &copy
	}

	var newReq models.GeminiRequest
	if err := json.Unmarshal(data, &newReq); err != nil {
		// Fallback to shallow copy if unmarshaling fails
		copy := *req
		return &copy
	}

	return &newReq
}
