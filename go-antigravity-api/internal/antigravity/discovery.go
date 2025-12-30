package antigravity

import (
	"fmt"
	"log"
	"time"

	"go-antigravity-api/pkg/models"
)

// DiscoveryManager handles project ID discovery and initialization
type DiscoveryManager struct {
	client    *APIClient
	projectID string
}

// NewDiscoveryManager creates a new DiscoveryManager
func NewDiscoveryManager(client *APIClient) *DiscoveryManager {
	return &DiscoveryManager{
		client: client,
	}
}

// DiscoverProject discovers or creates a project ID
func (dm *DiscoveryManager) DiscoverProject() (string, error) {
	if dm.projectID != "" {
		log.Printf("[Antigravity] Using pre-configured Project ID: %s", dm.projectID)
		return dm.projectID, nil
	}

	log.Println("[Antigravity] Discovering Project ID...")

	// Prepare client metadata
	clientMetadata := &models.ClientMetadata{
		IDEType:     "IDE_UNSPECIFIED",
		Platform:    "PLATFORM_UNSPECIFIED",
		PluginType:  "GEMINI",
		DuetProject: "",
	}

	// Call loadCodeAssist to discover project ID
	loadRequest := &models.LoadCodeAssistRequest{
		CloudAICompanionProject: "",
		Metadata:                clientMetadata,
	}

	var loadResponse models.LoadCodeAssistResponse
	if err := dm.client.CallAPI("loadCodeAssist", loadRequest, &loadResponse); err == nil {
		if loadResponse.CloudAICompanionProject != "" {
			log.Printf("[Antigravity] Discovered existing Project ID: %s", loadResponse.CloudAICompanionProject)
			dm.projectID = loadResponse.CloudAICompanionProject
			return dm.projectID, nil
		}

		// If no existing project, onboard user
		return dm.onboardUser(clientMetadata, loadResponse.AllowedTiers)
	}

	// If discovery fails, generate a random project ID
	log.Println("[Antigravity] Failed to discover Project ID, generating fallback...")
	fallbackID := GenerateProjectID()
	log.Printf("[Antigravity] Generated fallback Project ID: %s", fallbackID)
	dm.projectID = fallbackID
	return dm.projectID, nil
}

// onboardUser onboards a new user and creates a project
func (dm *DiscoveryManager) onboardUser(metadata *models.ClientMetadata, tiers []models.Tier) (string, error) {
	// Find default tier
	tierID := "free-tier"
	for _, tier := range tiers {
		if tier.IsDefault {
			tierID = tier.ID
			break
		}
	}

	onboardRequest := &models.OnboardUserRequest{
		TierID:                  tierID,
		CloudAICompanionProject: "",
		Metadata:                metadata,
	}

	var lroResponse models.OnboardUserResponse
	if err := dm.client.CallAPI("onboardUser", onboardRequest, &lroResponse); err != nil {
		return "", fmt.Errorf("failed to start onboarding: %w", err)
	}

	// Poll until operation is complete
	maxRetries := 30 // 60 seconds total
	retryCount := 0

	for !lroResponse.Done && retryCount < maxRetries {
		time.Sleep(2 * time.Second)
		if err := dm.client.CallAPI("onboardUser", onboardRequest, &lroResponse); err != nil {
			log.Printf("[Antigravity] Error polling onboarding status: %v", err)
		}
		retryCount++
	}

	if !lroResponse.Done {
		return "", fmt.Errorf("onboarding timeout: operation did not complete within expected time")
	}

	if lroResponse.Response != nil && lroResponse.Response.CloudAICompanionProject != nil {
		discoveredID := lroResponse.Response.CloudAICompanionProject.ID
		log.Printf("[Antigravity] Onboarded and discovered Project ID: %s", discoveredID)
		dm.projectID = discoveredID
		return dm.projectID, nil
	}

	return "", fmt.Errorf("failed to get project ID from onboarding response")
}

// SetProjectID sets the project ID
func (dm *DiscoveryManager) SetProjectID(projectID string) {
	dm.projectID = projectID
}

// GetProjectID returns the current project ID
func (dm *DiscoveryManager) GetProjectID() string {
	return dm.projectID
}
