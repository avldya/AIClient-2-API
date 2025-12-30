package antigravity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"go-antigravity-api/pkg/models"
)

// QuotaManager handles quota queries
type QuotaManager struct {
	client *APIClient
}

// NewQuotaManager creates a new QuotaManager
func NewQuotaManager(client *APIClient) *QuotaManager {
	return &QuotaManager{
		client: client,
	}
}

// GetUsageLimits fetches usage limits for all models
func (qm *QuotaManager) GetUsageLimits() (*models.UsageLimits, error) {
	result := &models.UsageLimits{
		LastUpdated: time.Now().Unix(),
		Models:      make(map[string]models.QuotaInfo),
	}

	// Try each base URL
	for _, baseURL := range qm.client.baseURLs {
		url := fmt.Sprintf("%s/%s:fetchAvailableModels", baseURL, APIVersion)

		accessToken, err := qm.client.authManager.GetAccessToken()
		if err != nil {
			continue
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader([]byte("{}")))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		req.Header.Set("User-Agent", qm.client.userAgent)

		resp, err := qm.client.httpClient.Do(req)
		if err != nil {
			log.Printf("[Antigravity] Failed to fetch quotas from %s: %v", baseURL, err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var apiResponse models.FetchAvailableModelsResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			// Extract quota information for each model
			for modelID, modelData := range apiResponse.Models {
				aliasName := ModelName2Alias(modelID)
				if aliasName == "" {
					continue // Skip unsupported models
				}

				quotaInfo := models.QuotaInfo{
					Remaining: 0,
				}

				if modelData.QuotaInfo != nil {
					// Use remainingFraction or remaining
					if modelData.QuotaInfo.RemainingFraction > 0 {
						quotaInfo.Remaining = modelData.QuotaInfo.RemainingFraction
					} else {
						quotaInfo.Remaining = modelData.QuotaInfo.Remaining
					}

					// Parse reset time
					if modelData.QuotaInfo.ResetTime != "" {
						quotaInfo.ResetTimeRaw = modelData.QuotaInfo.ResetTime
						if resetTime, err := time.Parse(time.RFC3339, modelData.QuotaInfo.ResetTime); err == nil {
							quotaInfo.ResetTime = &resetTime
						}
					}
				}

				result.Models[aliasName] = quotaInfo
			}

			// Sort models alphabetically
			sortedModels := make(map[string]models.QuotaInfo)
			keys := make([]string, 0, len(result.Models))
			for k := range result.Models {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				sortedModels[k] = result.Models[k]
			}
			result.Models = sortedModels

			log.Printf("[Antigravity] Successfully fetched quotas for %d models", len(result.Models))
			return result, nil
		}
		resp.Body.Close()
	}

	return nil, fmt.Errorf("failed to fetch quotas from all endpoints")
}
