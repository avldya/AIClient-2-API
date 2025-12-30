package antigravity

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"go-antigravity-api/pkg/models"
)

// ModelsManager manages available models
type ModelsManager struct {
	availableModels []string
}

// NewModelsManager creates a new ModelsManager
func NewModelsManager() *ModelsManager {
	return &ModelsManager{
		availableModels: []string{},
	}
}

// SetAvailableModels sets the available models
func (mm *ModelsManager) SetAvailableModels(models []string) {
	mm.availableModels = models
}

// GetAvailableModels returns the available models
func (mm *ModelsManager) GetAvailableModels() []string {
	return mm.availableModels
}

// ListModels returns formatted model information
func (mm *ModelsManager) ListModels() map[string]interface{} {
	now := time.Now().Unix()
	formattedModels := make([]models.ModelInfo, 0, len(mm.availableModels))

	for _, modelID := range mm.availableModels {
		// Create display name from model ID
		words := strings.Split(modelID, "-")
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
		displayName := strings.Join(words, " ")

		modelInfo := models.ModelInfo{
			Name:                       fmt.Sprintf("models/%s", modelID),
			Version:                    "1.0.0",
			DisplayName:                displayName,
			Description:                fmt.Sprintf("Antigravity model: %s", modelID),
			InputTokenLimit:            1024000,
			OutputTokenLimit:           65535,
			SupportedGenerationMethods: []string{"generateContent", "streamGenerateContent"},
			Object:                     "model",
			Created:                    now,
			OwnedBy:                    "antigravity",
			Type:                       "antigravity",
		}

		// Add thinking capabilities for thinking models
		if strings.Contains(modelID, "-thinking") {
			modelInfo.Thinking = &models.ThinkingCapabilities{
				Min:            1024,
				Max:            100000,
				ZeroAllowed:    false,
				DynamicAllowed: true,
			}
		}

		formattedModels = append(formattedModels, modelInfo)
	}

	return map[string]interface{}{
		"models": formattedModels,
	}
}

// ValidateModel checks if a model is available
func (mm *ModelsManager) ValidateModel(modelName string) (string, error) {
	// Check if model exists in available models
	for _, available := range mm.availableModels {
		if available == modelName {
			return modelName, nil
		}
	}

	// If not found and we have models, use the first one
	if len(mm.availableModels) > 0 {
		return mm.availableModels[0], fmt.Errorf("model '%s' not found, using default: %s", modelName, mm.availableModels[0])
	}

	return "", fmt.Errorf("no models available")
}

// ParseModelsFromResponse parses models from fetchAvailableModels response
func ParseModelsFromResponse(response map[string]interface{}) []string {
	modelsMap, ok := response["models"].(map[string]interface{})
	if !ok {
		return []string{}
	}

	models := make([]string, 0, len(modelsMap))
	for modelName := range modelsMap {
		alias := ModelName2Alias(modelName)
		if alias != "" {
			models = append(models, alias)
		}
	}

	// Sort models alphabetically
	sort.Strings(models)
	return models
}
