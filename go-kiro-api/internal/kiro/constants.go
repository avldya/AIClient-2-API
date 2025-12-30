package kiro

// Constants for Kiro API
// 常量定义
const (
	// API URLs
	RefreshURL     = "https://prod.{{region}}.auth.desktop.kiro.dev/refreshToken"
	RefreshIDCURL  = "https://oidc.{{region}}.amazonaws.com/token"
	BaseURL        = "https://codewhisperer.{{region}}.amazonaws.com/generateAssistantResponse"
	AmazonQURL     = "https://codewhisperer.{{region}}.amazonaws.com/SendMessageStreaming"
	UsageLimitsURL = "https://q.{{region}}.amazonaws.com/getUsageLimits"

	// Default values
	DefaultModel      = "claude-opus-4-5"
	DefaultRegion     = "us-east-1"
	KiroVersion       = "0.7.5"
	UserAgent         = "KiroIDE"
	ContentTypeJSON   = "application/json"
	AcceptJSON        = "application/json"
	AuthMethodSocial  = "social"
	ChatTriggerManual = "MANUAL"
	OriginAIEditor    = "AI_EDITOR"

	// Timeout and retry settings
	AxiosTimeout      = 300000 // 5 minutes in milliseconds
	DefaultMaxRetries = 3
	DefaultBaseDelay  = 1000 // milliseconds

	// Token refresh
	DefaultCronNearMinutes = 10 // Refresh token 10 minutes before expiry

	// File names
	KiroAuthTokenFile = "kiro-auth-token.json"
)

// ModelMapping maps user-facing model names to CodeWhisperer internal model names
// 模型映射表：将用户使用的模型名称映射到 CodeWhisperer 内部的模型名称
var ModelMapping = map[string]string{
	"claude-opus-4-5":            "claude-opus-4.5",
	"claude-opus-4-5-20251101":   "claude-opus-4.5",
	"claude-haiku-4-5":           "claude-haiku-4.5",
	"claude-sonnet-4-5":          "CLAUDE_SONNET_4_5_20250929_V1_0",
	"claude-sonnet-4-5-20250929": "CLAUDE_SONNET_4_5_20250929_V1_0",
	"claude-sonnet-4-20250514":   "CLAUDE_SONNET_4_20250514_V1_0",
	"claude-3-7-sonnet-20250219": "CLAUDE_3_7_SONNET_20250219_V1_0",
}

// SupportedModels lists all supported model names
// 支持的模型列表
var SupportedModels = []string{
	"claude-opus-4-5",
	"claude-opus-4-5-20251101",
	"claude-haiku-4-5",
	"claude-sonnet-4-5",
	"claude-sonnet-4-5-20250929",
	"claude-sonnet-4-20250514",
	"claude-3-7-sonnet-20250219",
}
