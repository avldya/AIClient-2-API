package kiro

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

// GenerateMachineID generates a unique machine ID based on credentials
// 根据凭据生成唯一的机器码
func GenerateMachineID(credentials map[string]interface{}) string {
	// Priority: UUID > ProfileArn > ClientId > fallback
	var uniqueKey string

	if val, ok := credentials["uuid"].(string); ok && val != "" {
		uniqueKey = val
	} else if val, ok := credentials["profileArn"].(string); ok && val != "" {
		uniqueKey = val
	} else if val, ok := credentials["clientId"].(string); ok && val != "" {
		uniqueKey = val
	} else {
		uniqueKey = "KIRO_DEFAULT_MACHINE"
	}

	hash := sha256.Sum256([]byte(uniqueKey))
	return hex.EncodeToString(hash[:])
}

// GetSystemRuntimeInfo returns system runtime information
// 获取系统运行时信息
func GetSystemRuntimeInfo() map[string]string {
	osName := runtime.GOOS
	osArch := runtime.GOARCH
	goVersion := runtime.Version()

	// Format OS name similar to Node.js implementation
	if osName == "windows" {
		osName = fmt.Sprintf("windows#%s", osArch)
	} else if osName == "darwin" {
		osName = fmt.Sprintf("macos#%s", osArch)
	} else {
		osName = fmt.Sprintf("%s#%s", osName, osArch)
	}

	return map[string]string{
		"osName":    osName,
		"goVersion": strings.TrimPrefix(goVersion, "go"),
	}
}

// GenerateUUID generates a new UUID v4
// 生成 UUID v4
func GenerateUUID() string {
	return uuid.New().String()
}

// FindMatchingBracket finds the matching closing bracket for an opening bracket
// 通用的括号匹配函数 - 支持多种括号类型
// Returns the position of the matching closing bracket, or -1 if not found
func FindMatchingBracket(text string, startPos int, openChar, closeChar byte) int {
	if startPos >= len(text) || text[startPos] != openChar {
		return -1
	}

	count := 0
	inString := false
	escapeNext := false

	for i := startPos; i < len(text); i++ {
		char := text[i]

		if escapeNext {
			escapeNext = false
			continue
		}

		if char == '\\' {
			escapeNext = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		if !inString {
			if char == openChar {
				count++
			} else if char == closeChar {
				count--
				if count == 0 {
					return i
				}
			}
		}
	}

	return -1
}

// FixJSON attempts to fix common JSON formatting issues
// 修复常见的 JSON 格式问题
func FixJSON(jsonStr string) string {
	// Remove trailing commas before } or ]
	jsonStr = strings.ReplaceAll(jsonStr, ",}", "}")
	jsonStr = strings.ReplaceAll(jsonStr, ",]", "]")

	// Trim whitespace
	jsonStr = strings.TrimSpace(jsonStr)

	return jsonStr
}

// ParseBracketFormat parses tool calls from bracket format
// 解析括号格式的工具调用 [Called function with args: {...}]
func ParseBracketFormat(content string) (string, string, map[string]interface{}, bool) {
	// Look for pattern: [Called xxx with args: {...}]
	if !strings.HasPrefix(content, "[Called ") {
		return "", "", nil, false
	}

	// Find "with args:"
	argsPos := strings.Index(content, " with args: ")
	if argsPos == -1 {
		return "", "", nil, false
	}

	// Extract function name
	functionName := strings.TrimSpace(content[8:argsPos]) // Skip "[Called "

	// Find the opening { for args
	argsStart := strings.Index(content[argsPos:], "{")
	if argsStart == -1 {
		return "", "", nil, false
	}
	argsStart += argsPos

	// Find matching closing }
	argsEnd := FindMatchingBracket(content, argsStart, '{', '}')
	if argsEnd == -1 {
		return "", "", nil, false
	}

	// Extract and parse JSON args
	argsJSON := content[argsStart : argsEnd+1]
	argsJSON = FixJSON(argsJSON)

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", "", nil, false
	}

	return functionName, argsJSON, args, true
}

// CountTokens estimates token count (simple approximation: chars / 4)
// 估算 Token 数量（简单估算：字符数 / 4）
func CountTokens(text string) int {
	return len(text) / 4
}

// DeduplicateToolUses removes duplicate tool uses based on toolUseId
// 去重工具调用（基于 toolUseId）
func DeduplicateToolUses(toolUses []map[string]interface{}) []map[string]interface{} {
	seen := make(map[string]bool)
	result := make([]map[string]interface{}, 0)

	for _, toolUse := range toolUses {
		if id, ok := toolUse["toolUseId"].(string); ok {
			if !seen[id] {
				seen[id] = true
				result = append(result, toolUse)
			}
		} else {
			// If no toolUseId, keep it
			result = append(result, toolUse)
		}
	}

	return result
}

// GetContentText extracts text content from various message content formats
// 从各种消息内容格式中提取文本
func GetContentText(content interface{}) string {
	if content == nil {
		return ""
	}

	// If it's a string, return directly
	if str, ok := content.(string); ok {
		return str
	}

	// If it's a map with "content" key
	if m, ok := content.(map[string]interface{}); ok {
		if val, exists := m["content"]; exists {
			return GetContentText(val)
		}
	}

	// If it's an array of content parts
	if arr, ok := content.([]interface{}); ok {
		var texts []string
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				if typ, exists := m["type"]; exists && typ == "text" {
					if text, ok := m["text"].(string); ok {
						texts = append(texts, text)
					}
				}
			}
		}
		return strings.Join(texts, "\n")
	}

	return ""
}

// ReplaceRegion replaces {{region}} placeholder in URL with actual region
// 替换 URL 中的 {{region}} 占位符
func ReplaceRegion(url, region string) string {
	return strings.ReplaceAll(url, "{{region}}", region)
}
