package antigravity

import (
	"fmt"
	"log"

	"go-antigravity-api/pkg/models"
)

// Service is the main Antigravity API service
type Service struct {
	authManager      *AuthManager
	apiClient        *APIClient
	modelsManager    *ModelsManager
	discoveryManager *DiscoveryManager
	quotaManager     *QuotaManager
	projectID        string
	isInitialized    bool
}

// NewService creates a new Antigravity service
func NewService(config map[string]interface{}) *Service {
	// Get credentials path
	credsPath := ""
	if val, ok := config["oauthCredsFilePath"].(string); ok {
		credsPath = val
	}

	// Get project ID
	projectID := ""
	if val, ok := config["projectId"].(string); ok {
		projectID = val
	}

	// Create auth manager
	authManager := NewAuthManager(credsPath)

	// Create API client
	apiClient := NewAPIClient(authManager, config)

	// Create managers
	modelsManager := NewModelsManager()
	discoveryManager := NewDiscoveryManager(apiClient)
	quotaManager := NewQuotaManager(apiClient)

	// Set project ID if provided
	if projectID != "" {
		discoveryManager.SetProjectID(projectID)
	}

	return &Service{
		authManager:      authManager,
		apiClient:        apiClient,
		modelsManager:    modelsManager,
		discoveryManager: discoveryManager,
		quotaManager:     quotaManager,
		projectID:        projectID,
		isInitialized:    false,
	}
}

// Initialize initializes the service
func (s *Service) Initialize() error {
	if s.isInitialized {
		return nil
	}

	log.Println("[Antigravity] Initializing Antigravity API Service...")

	// Initialize authentication
	if err := s.authManager.Initialize(false); err != nil {
		return fmt.Errorf("failed to initialize auth: %w", err)
	}

	// Discover or use provided project ID
	if s.projectID == "" {
		projectID, err := s.discoveryManager.DiscoverProject()
		if err != nil {
			log.Printf("[Antigravity] Warning: %v", err)
		}
		s.projectID = projectID
	} else {
		log.Printf("[Antigravity] Using provided Project ID: %s", s.projectID)
	}

	// Fetch available models
	models, err := s.apiClient.FetchAvailableModels()
	if err != nil {
		log.Printf("[Antigravity] Warning: failed to fetch models: %v", err)
		// Use default models
		models = []string{
			"gemini-3-pro-preview",
			"gemini-3-flash-preview",
			"gemini-2.5-flash",
		}
	}
	s.modelsManager.SetAvailableModels(models)

	s.isInitialized = true
	log.Printf("[Antigravity] Initialization complete. Project ID: %s", s.projectID)
	return nil
}

// GenerateContent generates content using a model
func (s *Service) GenerateContent(modelAlias string, request *models.GeminiRequest) (*models.GeminiResponse, error) {
	if !s.isInitialized {
		if err := s.Initialize(); err != nil {
			return nil, err
		}
	}

	// Validate and get model
	selectedModel, err := s.modelsManager.ValidateModel(modelAlias)
	if err != nil {
		log.Printf("[Antigravity] Warning: %v", err)
	}

	// Convert alias to real model name
	actualModelName := Alias2ModelName(selectedModel)

	// Convert request to Antigravity format
	antigravityReq := GeminiToAntigravity(actualModelName, request, s.projectID)

	// Make API call
	var antigravityResp models.AntigravityResponse
	if err := s.apiClient.CallAPI("generateContent", antigravityReq, &antigravityResp); err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	return ToGeminiAPIResponse(&antigravityResp), nil
}

// StreamGenerateContent generates content using streaming
func (s *Service) StreamGenerateContent(modelAlias string, request *models.GeminiRequest) (<-chan *models.GeminiResponse, <-chan error) {
	errorChan := make(chan error, 1)
	responseChan := make(chan *models.GeminiResponse, 10)

	go func() {
		defer close(responseChan)
		defer close(errorChan)

		if !s.isInitialized {
			if err := s.Initialize(); err != nil {
				errorChan <- err
				return
			}
		}

		// Validate and get model
		selectedModel, err := s.modelsManager.ValidateModel(modelAlias)
		if err != nil {
			log.Printf("[Antigravity] Warning: %v", err)
		}

		// Convert alias to real model name
		actualModelName := Alias2ModelName(selectedModel)

		// Convert request to Antigravity format
		antigravityReq := GeminiToAntigravity(actualModelName, request, s.projectID)

		// Make streaming API call
		streamChan, streamErrChan := s.apiClient.StreamAPI("streamGenerateContent", antigravityReq)

		// Forward responses and errors
		for {
			select {
			case resp, ok := <-streamChan:
				if !ok {
					return
				}
				responseChan <- resp
			case err, ok := <-streamErrChan:
				if ok && err != nil {
					errorChan <- err
				}
				return
			}
		}
	}()

	return responseChan, errorChan
}

// ListModels returns the list of available models
func (s *Service) ListModels() (map[string]interface{}, error) {
	if !s.isInitialized {
		if err := s.Initialize(); err != nil {
			return nil, err
		}
	}

	return s.modelsManager.ListModels(), nil
}

// GetUsageLimits returns usage limits for all models
func (s *Service) GetUsageLimits() (*models.UsageLimits, error) {
	if !s.isInitialized {
		if err := s.Initialize(); err != nil {
			return nil, err
		}
	}

	return s.quotaManager.GetUsageLimits()
}

// GetProjectID returns the current project ID
func (s *Service) GetProjectID() string {
	return s.projectID
}
