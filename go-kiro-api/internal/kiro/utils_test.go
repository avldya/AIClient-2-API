package kiro

import (
	"testing"
)

func TestGenerateMachineID(t *testing.T) {
	tests := []struct {
		name        string
		credentials map[string]interface{}
		wantLen     int
	}{
		{
			name: "with uuid",
			credentials: map[string]interface{}{
				"uuid": "test-uuid-123",
			},
			wantLen: 64, // SHA256 hex string length
		},
		{
			name: "with profileArn",
			credentials: map[string]interface{}{
				"profileArn": "arn:aws:iam::123456789012:role/test",
			},
			wantLen: 64,
		},
		{
			name: "with clientId",
			credentials: map[string]interface{}{
				"clientId": "test-client-id",
			},
			wantLen: 64,
		},
		{
			name:        "with no identifiers",
			credentials: map[string]interface{}{},
			wantLen:     64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateMachineID(tt.credentials)
			if len(result) != tt.wantLen {
				t.Errorf("GenerateMachineID() length = %v, want %v", len(result), tt.wantLen)
			}
		})
	}
}

func TestGetSystemRuntimeInfo(t *testing.T) {
	info := GetSystemRuntimeInfo()

	if info["osName"] == "" {
		t.Error("GetSystemRuntimeInfo() osName is empty")
	}
	if info["goVersion"] == "" {
		t.Error("GetSystemRuntimeInfo() goVersion is empty")
	}
}

func TestFindMatchingBracket(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		startPos  int
		openChar  byte
		closeChar byte
		want      int
	}{
		{
			name:      "simple brackets",
			text:      "[test]",
			startPos:  0,
			openChar:  '[',
			closeChar: ']',
			want:      5,
		},
		{
			name:      "nested brackets",
			text:      "[test [nested] end]",
			startPos:  0,
			openChar:  '[',
			closeChar: ']',
			want:      18,
		},
		{
			name:      "curly braces",
			text:      `{"key": "value"}`,
			startPos:  0,
			openChar:  '{',
			closeChar: '}',
			want:      15,
		},
		{
			name:      "nested with quotes",
			text:      `{"key": "val[ue]"}`,
			startPos:  0,
			openChar:  '{',
			closeChar: '}',
			want:      17,
		},
		{
			name:      "not found",
			text:      "[incomplete",
			startPos:  0,
			openChar:  '[',
			closeChar: ']',
			want:      -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindMatchingBracket(tt.text, tt.startPos, tt.openChar, tt.closeChar)
			if got != tt.want {
				t.Errorf("FindMatchingBracket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountTokens(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "simple text",
			text: "Hello, world!",
			want: 3, // 13 / 4 = 3
		},
		{
			name: "empty text",
			text: "",
			want: 0,
		},
		{
			name: "longer text",
			text: "This is a longer piece of text to test token counting",
			want: 13, // 54 / 4 = 13
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountTokens(tt.text)
			if got != tt.want {
				t.Errorf("CountTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFixJSON(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{
			name: "trailing comma in object",
			json: `{"key": "value",}`,
			want: `{"key": "value"}`,
		},
		{
			name: "trailing comma in array",
			json: `["item1", "item2",]`,
			want: `["item1", "item2"]`,
		},
		{
			name: "no trailing comma",
			json: `{"key": "value"}`,
			want: `{"key": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FixJSON(tt.json)
			if got != tt.want {
				t.Errorf("FixJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceRegion(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		region string
		want   string
	}{
		{
			name:   "replace region in URL",
			url:    "https://api.{{region}}.example.com",
			region: "us-east-1",
			want:   "https://api.us-east-1.example.com",
		},
		{
			name:   "multiple replacements",
			url:    "https://{{region}}.api.{{region}}.example.com",
			region: "eu-west-1",
			want:   "https://eu-west-1.api.eu-west-1.example.com",
		},
		{
			name:   "no placeholder",
			url:    "https://api.example.com",
			region: "us-east-1",
			want:   "https://api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceRegion(tt.url, tt.region)
			if got != tt.want {
				t.Errorf("ReplaceRegion() = %v, want %v", got, tt.want)
			}
		})
	}
}
