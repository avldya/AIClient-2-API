package antigravity

const (
	// OAuth Configuration
	OAuthClientID     = "1071006060591-tmhssin2h21lcre235vtolojh4g403ep.apps.googleusercontent.com"
	OAuthClientSecret = "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"
	RefreshSkew       = 3000 // 50 minutes in seconds

	// Credentials Storage
	CredentialsDir  = ".antigravity"
	CredentialsFile = "oauth_creds.json"

	// API Configuration
	APIVersion          = "v1internal"
	DefaultUserAgent    = "antigravity/1.11.5 windows/amd64"
	BaseURLDaily        = "https://daily-cloudcode-pa.sandbox.googleapis.com"
	BaseURLAutopush     = "https://autopush-cloudcode-pa.sandbox.googleapis.com"

	// Retry Configuration
	DefaultMaxRetries = 3
	DefaultBaseDelay  = 1000 // milliseconds
)

// Model alias mappings
var ModelAliasMap = map[string]string{
	"gemini-2.5-computer-use-preview-10-2025": "rev19-uic3-1p",
	"gemini-3-pro-image-preview":              "gemini-3-pro-image",
	"gemini-3-pro-preview":                    "gemini-3-pro-high",
	"gemini-3-flash-preview":                  "gemini-3-flash",
	"gemini-2.5-flash":                        "gemini-2.5-flash",
	"gemini-claude-sonnet-4-5":                "claude-sonnet-4-5",
	"gemini-claude-sonnet-4-5-thinking":       "claude-sonnet-4-5-thinking",
	"gemini-claude-opus-4-5-thinking":         "claude-opus-4-5-thinking",
}

// Model name mappings (reverse)
var ModelNameMap = map[string]string{
	"rev19-uic3-1p":                 "gemini-2.5-computer-use-preview-10-2025",
	"gemini-3-pro-image":            "gemini-3-pro-image-preview",
	"gemini-3-pro-high":             "gemini-3-pro-preview",
	"gemini-3-flash":                "gemini-3-flash-preview",
	"gemini-2.5-flash":              "gemini-2.5-flash",
	"claude-sonnet-4-5":             "gemini-claude-sonnet-4-5",
	"claude-sonnet-4-5-thinking":    "gemini-claude-sonnet-4-5-thinking",
	"claude-opus-4-5-thinking":      "gemini-claude-opus-4-5-thinking",
}
