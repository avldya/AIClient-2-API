package models

import "time"

// OAuthCredentials represents the OAuth2 credentials
type OAuthCredentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiryDate   int64  `json:"expiry_date"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

// GeminiRequest represents a Gemini API request
type GeminiRequest struct {
	Contents          []Content          `json:"contents,omitempty"`
	SystemInstruction *Content           `json:"systemInstruction,omitempty"`
	Tools             []Tool             `json:"tools,omitempty"`
	ToolConfig        *ToolConfig        `json:"toolConfig,omitempty"`
	SafetySettings    []SafetySetting    `json:"safetySettings,omitempty"`
	GenerationConfig  *GenerationConfig  `json:"generationConfig,omitempty"`
}

// Content represents a content object
type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts,omitempty"`
}

// Part represents a part of content
type Part struct {
	Text         string        `json:"text,omitempty"`
	InlineData   *InlineData   `json:"inlineData,omitempty"`
	FunctionCall *FunctionCall `json:"functionCall,omitempty"`
}

// InlineData represents inline data (e.g., images)
type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// Tool represents a tool definition
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations,omitempty"`
}

// FunctionDeclaration represents a function declaration
type FunctionDeclaration struct {
	Name                 string                 `json:"name"`
	Description          string                 `json:"description"`
	Parameters           map[string]interface{} `json:"parameters,omitempty"`
	ParametersJSONSchema map[string]interface{} `json:"parametersJsonSchema,omitempty"`
}

// ToolConfig represents tool configuration
type ToolConfig struct {
	FunctionCallingConfig *FunctionCallingConfig `json:"functionCallingConfig,omitempty"`
}

// FunctionCallingConfig represents function calling configuration
type FunctionCallingConfig struct {
	Mode string `json:"mode,omitempty"`
}

// SafetySetting represents a safety setting
type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// GenerationConfig represents generation configuration
type GenerationConfig struct {
	Temperature      *float64        `json:"temperature,omitempty"`
	TopP             *float64        `json:"topP,omitempty"`
	TopK             *int            `json:"topK,omitempty"`
	MaxOutputTokens  *int            `json:"maxOutputTokens,omitempty"`
	StopSequences    []string        `json:"stopSequences,omitempty"`
	ThinkingConfig   *ThinkingConfig `json:"thinkingConfig,omitempty"`
}

// ThinkingConfig represents thinking configuration
type ThinkingConfig struct {
	ThinkingLevel  string `json:"thinkingLevel,omitempty"`
	ThinkingBudget int    `json:"thinkingBudget,omitempty"`
}

// AntigravityRequest represents an Antigravity API request
type AntigravityRequest struct {
	Model     string         `json:"model"`
	UserAgent string         `json:"userAgent"`
	Project   string         `json:"project"`
	RequestID string         `json:"requestId"`
	Request   *GeminiRequest `json:"request"`
}

// GeminiResponse represents a Gemini API response
type GeminiResponse struct {
	Candidates                      []Candidate     `json:"candidates,omitempty"`
	UsageMetadata                   *UsageMetadata  `json:"usageMetadata,omitempty"`
	PromptFeedback                  *PromptFeedback `json:"promptFeedback,omitempty"`
	AutomaticFunctionCallingHistory []interface{}   `json:"automaticFunctionCallingHistory,omitempty"`
}

// Candidate represents a response candidate
type Candidate struct {
	Content       *Content       `json:"content,omitempty"`
	FinishReason  string         `json:"finishReason,omitempty"`
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
	Index         int            `json:"index,omitempty"`
}

// SafetyRating represents a safety rating
type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// UsageMetadata represents usage metadata
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount,omitempty"`
	CandidatesTokenCount int `json:"candidatesTokenCount,omitempty"`
	TotalTokenCount      int `json:"totalTokenCount,omitempty"`
}

// PromptFeedback represents prompt feedback
type PromptFeedback struct {
	SafetyRatings []SafetyRating `json:"safetyRatings,omitempty"`
}

// AntigravityResponse represents an Antigravity API response
type AntigravityResponse struct {
	Response *GeminiResponse `json:"response,omitempty"`
}

// ModelInfo represents model information
type ModelInfo struct {
	Name                       string                 `json:"name"`
	Version                    string                 `json:"version"`
	DisplayName                string                 `json:"displayName"`
	Description                string                 `json:"description"`
	InputTokenLimit            int                    `json:"inputTokenLimit"`
	OutputTokenLimit           int                    `json:"outputTokenLimit"`
	SupportedGenerationMethods []string               `json:"supportedGenerationMethods"`
	Object                     string                 `json:"object"`
	Created                    int64                  `json:"created"`
	OwnedBy                    string                 `json:"ownedBy"`
	Type                       string                 `json:"type"`
	Thinking                   *ThinkingCapabilities  `json:"thinking,omitempty"`
}

// ThinkingCapabilities represents thinking capabilities
type ThinkingCapabilities struct {
	Min            int  `json:"min"`
	Max            int  `json:"max"`
	ZeroAllowed    bool `json:"zeroAllowed"`
	DynamicAllowed bool `json:"dynamicAllowed"`
}

// QuotaInfo represents quota information for a model
type QuotaInfo struct {
	Remaining     float64    `json:"remaining"`
	ResetTime     *time.Time `json:"resetTime,omitempty"`
	ResetTimeRaw  string     `json:"resetTimeRaw,omitempty"`
}

// UsageLimits represents usage limits for all models
type UsageLimits struct {
	LastUpdated int64                `json:"lastUpdated"`
	Models      map[string]QuotaInfo `json:"models"`
}

// LoadCodeAssistRequest represents a request to loadCodeAssist
type LoadCodeAssistRequest struct {
	CloudAICompanionProject string          `json:"cloudaicompanionProject"`
	Metadata                *ClientMetadata `json:"metadata"`
}

// ClientMetadata represents client metadata
type ClientMetadata struct {
	IDEType     string `json:"ideType"`
	Platform    string `json:"platform"`
	PluginType  string `json:"pluginType"`
	DuetProject string `json:"duetProject"`
}

// LoadCodeAssistResponse represents a response from loadCodeAssist
type LoadCodeAssistResponse struct {
	CloudAICompanionProject string `json:"cloudaicompanionProject,omitempty"`
	AllowedTiers            []Tier `json:"allowedTiers,omitempty"`
}

// Tier represents a tier
type Tier struct {
	ID        string `json:"id"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

// OnboardUserRequest represents a request to onboardUser
type OnboardUserRequest struct {
	TierID                  string          `json:"tierId"`
	CloudAICompanionProject string          `json:"cloudaicompanionProject"`
	Metadata                *ClientMetadata `json:"metadata"`
}

// OnboardUserResponse represents a response from onboardUser
type OnboardUserResponse struct {
	Done     bool                         `json:"done,omitempty"`
	Response *OnboardUserResponseResponse `json:"response,omitempty"`
}

// OnboardUserResponseResponse represents the nested response
type OnboardUserResponseResponse struct {
	CloudAICompanionProject *CloudAICompanionProject `json:"cloudaicompanionProject,omitempty"`
}

// CloudAICompanionProject represents a cloud AI companion project
type CloudAICompanionProject struct {
	ID string `json:"id"`
}

// FetchAvailableModelsResponse represents a response from fetchAvailableModels
type FetchAvailableModelsResponse struct {
	Models map[string]ModelData `json:"models"`
}

// ModelData represents model data with quota info
type ModelData struct {
	QuotaInfo *ModelQuotaInfo `json:"quotaInfo,omitempty"`
}

// ModelQuotaInfo represents quota information from the API
type ModelQuotaInfo struct {
	RemainingFraction float64 `json:"remainingFraction,omitempty"`
	Remaining         float64 `json:"remaining,omitempty"`
	ResetTime         string  `json:"resetTime,omitempty"`
}
