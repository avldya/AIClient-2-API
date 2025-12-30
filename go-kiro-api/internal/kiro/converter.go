package kiro

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/avldya/AIClient-2-API/go-kiro-api/pkg/models"
)

// Converter handles format conversion between OpenAI/Claude and CodeWhisperer
// 格式转换器
type Converter struct{}

// NewConverter creates a new converter
// 创建新的转换器
func NewConverter() *Converter {
	return &Converter{}
}

// BuildCodeWhispererRequest converts OpenAI/Claude format to CodeWhisperer format
// 将 OpenAI/Claude 格式转换为 CodeWhisperer 格式
func (c *Converter) BuildCodeWhispererRequest(messages []models.Message, model string, tools []models.Tool, systemPrompt string) (*models.CodeWhispererRequest, error) {
	conversationID := GenerateUUID()

	// Make a copy of messages to avoid modifying the original
	processedMessages := make([]models.Message, len(messages))
	copy(processedMessages, messages)

	if len(processedMessages) == 0 {
		return nil, fmt.Errorf("no user messages found")
	}

	// Remove last assistant message if it's just "{"
	if len(processedMessages) > 0 {
		lastMsg := processedMessages[len(processedMessages)-1]
		if lastMsg.Role == "assistant" {
			content := GetContentText(lastMsg.Content)
			if content == "{" {
				processedMessages = processedMessages[:len(processedMessages)-1]
			}
		}
	}

	// Merge adjacent messages with the same role
	processedMessages = c.mergeAdjacentMessages(processedMessages)

	// Get CodeWhisperer model name
	codewhispererModel := ModelMapping[model]
	if codewhispererModel == "" {
		codewhispererModel = ModelMapping[DefaultModel]
	}

	// Build tools context
	var toolsContext *models.ToolsContext
	if len(tools) > 0 {
		toolsContext = &models.ToolsContext{
			Tools: make([]models.ToolSpecification, len(tools)),
		}
		for i, tool := range tools {
			toolsContext.Tools[i] = models.ToolSpecification{
				ToolSpecification: &models.ToolSpec{
					Name:        tool.Name,
					Description: tool.Description,
					InputSchema: &models.InputSchema{
						JSON: tool.InputSchema,
					},
				},
			}
		}
	}

	// Build history
	history := make([]models.HistoryItem, 0)
	startIndex := 0

	// Handle system prompt
	if systemPrompt != "" {
		if processedMessages[0].Role == "user" {
			// Prepend system prompt to first user message
			firstUserContent := GetContentText(processedMessages[0].Content)
			history = append(history, models.HistoryItem{
				UserInputMessage: &models.UserInputMessage{
					Content: fmt.Sprintf("%s\n\n%s", systemPrompt, firstUserContent),
					ModelID: codewhispererModel,
					Origin:  OriginAIEditor,
				},
			})
			startIndex = 1
		} else {
			// Add system prompt as standalone user message
			history = append(history, models.HistoryItem{
				UserInputMessage: &models.UserInputMessage{
					Content: systemPrompt,
					ModelID: codewhispererModel,
					Origin:  OriginAIEditor,
				},
			})
		}
	}

	// Process remaining messages (except the last one which becomes userInputMessage)
	for i := startIndex; i < len(processedMessages)-1; i++ {
		msg := processedMessages[i]
		if msg.Role == "user" {
			userMsg, err := c.buildUserInputMessage(msg, codewhispererModel)
			if err != nil {
				return nil, err
			}
			history = append(history, models.HistoryItem{
				UserInputMessage: userMsg,
			})
		} else if msg.Role == "assistant" {
			assistantMsg := c.buildAssistantResponseMessage(msg)
			history = append(history, models.HistoryItem{
				AssistantResponseMessage: assistantMsg,
			})
		}
	}

	// Last message becomes the current userInputMessage
	lastMsg := processedMessages[len(processedMessages)-1]
	if lastMsg.Role != "user" {
		return nil, fmt.Errorf("last message must be from user")
	}

	userInputMessage, err := c.buildUserInputMessage(lastMsg, codewhispererModel)
	if err != nil {
		return nil, err
	}

	return &models.CodeWhispererRequest{
		ConversationID:   conversationID,
		UserInputMessage: userInputMessage,
		History:          history,
		ChatTriggerType:  ChatTriggerManual,
		Tools:            toolsContext,
	}, nil
}

// mergeAdjacentMessages merges adjacent messages with the same role
// 合并相邻相同角色的消息
func (c *Converter) mergeAdjacentMessages(messages []models.Message) []models.Message {
	if len(messages) == 0 {
		return messages
	}

	merged := make([]models.Message, 0)
	merged = append(merged, messages[0])

	for i := 1; i < len(messages); i++ {
		current := messages[i]
		last := &merged[len(merged)-1]

		if current.Role == last.Role {
			// Merge content
			lastContent := c.normalizeContent(last.Content)
			currentContent := c.normalizeContent(current.Content)
			last.Content = append(lastContent, currentContent...)
		} else {
			merged = append(merged, current)
		}
	}

	return merged
}

// normalizeContent normalizes content to []models.ContentPart
// 规范化内容为 []models.ContentPart
func (c *Converter) normalizeContent(content interface{}) []models.ContentPart {
	// If it's already a slice, try to convert
	if slice, ok := content.([]interface{}); ok {
		parts := make([]models.ContentPart, 0, len(slice))
		for _, item := range slice {
			if part, ok := item.(models.ContentPart); ok {
				parts = append(parts, part)
			} else if m, ok := item.(map[string]interface{}); ok {
				part := c.mapToContentPart(m)
				parts = append(parts, part)
			}
		}
		return parts
	}

	// If it's a string, convert to text content part
	if str, ok := content.(string); ok {
		return []models.ContentPart{{Type: "text", Text: str}}
	}

	// If it's a single ContentPart
	if part, ok := content.(models.ContentPart); ok {
		return []models.ContentPart{part}
	}

	// Default: treat as text
	return []models.ContentPart{{Type: "text", Text: fmt.Sprintf("%v", content)}}
}

// mapToContentPart converts a map to ContentPart
// 将 map 转换为 ContentPart
func (c *Converter) mapToContentPart(m map[string]interface{}) models.ContentPart {
	part := models.ContentPart{}

	if typ, ok := m["type"].(string); ok {
		part.Type = typ
	}
	if text, ok := m["text"].(string); ok {
		part.Text = text
	}
	if name, ok := m["name"].(string); ok {
		part.Name = name
	}
	if id, ok := m["id"].(string); ok {
		part.ID = id
	}
	if toolUseID, ok := m["tool_use_id"].(string); ok {
		part.ToolUseID = toolUseID
	}
	if input, ok := m["input"].(map[string]interface{}); ok {
		part.Input = input
	}
	if source, ok := m["source"].(map[string]interface{}); ok {
		imgSource := &models.ImageSource{}
		if typ, ok := source["type"].(string); ok {
			imgSource.Type = typ
		}
		if mediaType, ok := source["media_type"].(string); ok {
			imgSource.MediaType = mediaType
		}
		if data, ok := source["data"].(string); ok {
			imgSource.Data = data
		}
		part.Source = imgSource
	}

	return part
}

// buildUserInputMessage builds a UserInputMessage from a Message
// 从 Message 构建 UserInputMessage
func (c *Converter) buildUserInputMessage(msg models.Message, modelID string) (*models.UserInputMessage, error) {
	userMsg := &models.UserInputMessage{
		Content: "",
		ModelID: modelID,
		Origin:  OriginAIEditor,
	}

	images := make([]models.ImageData, 0)
	toolResults := make([]models.ToolResult, 0)

	// Process content
	contentParts := c.normalizeContent(msg.Content)
	for _, part := range contentParts {
		switch part.Type {
		case "text":
			userMsg.Content += part.Text
		case "image":
			if part.Source != nil {
				// Extract format from media_type (e.g., "image/png" -> "png")
				format := "png"
				if part.Source.MediaType != "" {
					parts := strings.Split(part.Source.MediaType, "/")
					if len(parts) > 1 {
						format = parts[1]
					}
				}
				images = append(images, models.ImageData{
					Format: format,
					Source: &models.ImageBytes{
						Bytes: part.Source.Data,
					},
				})
			}
		case "tool_result":
			toolResults = append(toolResults, models.ToolResult{
				Content: []models.ToolResultContent{
					{Text: GetContentText(part)},
				},
				Status:    "success",
				ToolUseID: part.ToolUseID,
			})
		}
	}

	// Add non-empty fields
	if len(images) > 0 {
		userMsg.Images = images
	}
	if len(toolResults) > 0 {
		// Deduplicate tool results - Kiro API doesn't accept duplicate toolUseIds
		seen := make(map[string]bool)
		uniqueResults := make([]models.ToolResult, 0)
		for _, tr := range toolResults {
			if !seen[tr.ToolUseID] {
				seen[tr.ToolUseID] = true
				uniqueResults = append(uniqueResults, tr)
			}
		}
		userMsg.ToolResults = uniqueResults
	}

	return userMsg, nil
}

// buildAssistantResponseMessage builds an AssistantResponseMessage from a Message
// 从 Message 构建 AssistantResponseMessage
func (c *Converter) buildAssistantResponseMessage(msg models.Message) *models.AssistantResponseMessage {
	assistantMsg := &models.AssistantResponseMessage{
		Content: "",
	}

	toolUses := make([]models.ToolUse, 0)

	// Process content
	contentParts := c.normalizeContent(msg.Content)
	for _, part := range contentParts {
		switch part.Type {
		case "text":
			assistantMsg.Content += part.Text
		case "tool_use":
			toolUse := models.ToolUse{
				Name:      part.Name,
				ToolUseID: part.ID,
			}
			if part.Input != nil {
				toolUse.Input = part.Input
			} else {
				toolUse.Input = make(map[string]interface{})
			}
			toolUses = append(toolUses, toolUse)
		}
	}

	if len(toolUses) > 0 {
		assistantMsg.ToolUses = toolUses
	}

	return assistantMsg
}

// ConvertToClaudeMessage converts CodeWhisperer response to Claude message format
// 将 CodeWhisperer 响应转换为 Claude 消息格式
func (c *Converter) ConvertToClaudeMessage(content string, toolUses []models.ToolUse, model string) *models.ClaudeMessage {
	messageID := GenerateUUID()

	msg := &models.ClaudeMessage{
		ID:      messageID,
		Type:    "message",
		Role:    "assistant",
		Model:   model,
		Content: make([]models.ClaudeContentBlock, 0),
	}

	// Add text content
	if content != "" {
		msg.Content = append(msg.Content, models.ClaudeContentBlock{
			Type: "text",
			Text: content,
		})
	}

	// Add tool uses
	for _, toolUse := range toolUses {
		msg.Content = append(msg.Content, models.ClaudeContentBlock{
			Type:  "tool_use",
			ID:    toolUse.ToolUseID,
			Name:  toolUse.Name,
			Input: toolUse.Input,
		})
	}

	// Set stop reason
	if len(toolUses) > 0 {
		stopReason := "tool_use"
		msg.StopReason = &stopReason
	} else {
		stopReason := "end_turn"
		msg.StopReason = &stopReason
	}

	// Estimate usage
	inputTokens := CountTokens(content)
	msg.Usage = &models.ClaudeUsage{
		InputTokens:  inputTokens,
		OutputTokens: inputTokens,
	}

	return msg
}

// ConvertToOpenAIResponse converts CodeWhisperer response to OpenAI format
// 将 CodeWhisperer 响应转换为 OpenAI 格式
func (c *Converter) ConvertToOpenAIResponse(content string, toolUses []models.ToolUse, model string) map[string]interface{} {
	messageID := GenerateUUID()

	message := map[string]interface{}{
		"role":    "assistant",
		"content": content,
	}

	// Add tool calls if present
	if len(toolUses) > 0 {
		toolCalls := make([]map[string]interface{}, len(toolUses))
		for i, toolUse := range toolUses {
			argsJSON, _ := json.Marshal(toolUse.Input)
			toolCalls[i] = map[string]interface{}{
				"id":   toolUse.ToolUseID,
				"type": "function",
				"function": map[string]interface{}{
					"name":      toolUse.Name,
					"arguments": string(argsJSON),
				},
			}
		}
		message["tool_calls"] = toolCalls
	}

	finishReason := "stop"
	if len(toolUses) > 0 {
		finishReason = "tool_calls"
	}

	return map[string]interface{}{
		"id":      messageID,
		"object":  "chat.completion",
		"created": GenerateUUID(),
		"model":   model,
		"choices": []map[string]interface{}{
			{
				"index":         0,
				"message":       message,
				"finish_reason": finishReason,
			},
		},
	}
}
