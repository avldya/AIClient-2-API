package kiro

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/avldya/AIClient-2-API/go-kiro-api/pkg/models"
)

// StreamParser handles parsing of AWS Event Stream and SSE formats
// 流式响应解析器
type StreamParser struct {
	converter *Converter
}

// NewStreamParser creates a new stream parser
// 创建新的流解析器
func NewStreamParser() *StreamParser {
	return &StreamParser{
		converter: NewConverter(),
	}
}

// ParseAWSEventStream parses AWS Event Stream format
// 解析 AWS Event Stream 格式
func (sp *StreamParser) ParseAWSEventStream(reader io.Reader, model string) <-chan interface{} {
	ch := make(chan interface{})

	go func() {
		defer close(ch)

		buffer := ""
		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024) // Increase buffer size for large responses

		// Track state for tool use accumulation
		var currentToolUse *models.ToolUse
		var accumulatedContent strings.Builder
		var lastContentEvent string

		for scanner.Scan() {
			chunk := scanner.Text()
			buffer += chunk

			// Parse events from buffer
			events, remaining := sp.parseAWSEventStreamBuffer(buffer)
			buffer = remaining

			// Process events
			for _, event := range events {
				switch event.Type {
				case "content":
					// Skip duplicate content events
					if event.Content != "" && event.Content != lastContentEvent {
						lastContentEvent = event.Content
						accumulatedContent.WriteString(event.Content)

						// Check if this looks like a bracket-format tool call
						if strings.HasPrefix(event.Content, "[Called ") {
							// Try to parse as bracket format
							funcName, _, args, ok := ParseBracketFormat(event.Content)
							if ok {
								// Convert to structured tool use
								toolUseID := GenerateUUID()
								if currentToolUse != nil {
									// Send previous tool use
									ch <- sp.createToolUseEvent(currentToolUse)
								}
								currentToolUse = &models.ToolUse{
									Name:      funcName,
									ToolUseID: toolUseID,
									Input:     args,
								}
								continue
							}
						}

						// Send content event
						ch <- sp.createContentEvent(event.Content, model)
					}

				case "toolUse":
					// Start of a new tool use
					if currentToolUse != nil {
						// Send previous tool use
						ch <- sp.createToolUseEvent(currentToolUse)
					}
					currentToolUse = event.ToolUse

				case "toolUseInput":
					// Continuation of tool use input
					if currentToolUse != nil {
						if event.Input != "" {
							// Append to existing input
							if inputStr, ok := currentToolUse.Input["input"].(string); ok {
								currentToolUse.Input["input"] = inputStr + event.Input
							} else {
								// Try to parse as JSON and merge
								var inputData map[string]interface{}
								if err := json.Unmarshal([]byte(event.Input), &inputData); err == nil {
									for k, v := range inputData {
										currentToolUse.Input[k] = v
									}
								} else {
									currentToolUse.Input["input"] = event.Input
								}
							}
						}
					}

				case "toolUseStop":
					// End of tool use
					if currentToolUse != nil {
						ch <- sp.createToolUseEvent(currentToolUse)
						currentToolUse = nil
					}
				}
			}
		}

		// Send any remaining tool use
		if currentToolUse != nil {
			ch <- sp.createToolUseEvent(currentToolUse)
		}

		// Send final message stop event
		ch <- sp.createMessageStopEvent()
	}()

	return ch
}

// parseAWSEventStreamBuffer parses events from a buffer
// 从缓冲区解析事件
func (sp *StreamParser) parseAWSEventStreamBuffer(buffer string) ([]models.StreamEvent, string) {
	events := make([]models.StreamEvent, 0)
	remaining := buffer
	searchStart := 0

	for {
		// Look for JSON payload patterns
		contentStart := indexFrom(remaining, `{"content":`, searchStart)
		nameStart := indexFrom(remaining, `{"name":`, searchStart)
		followupStart := indexFrom(remaining, `{"followupPrompt":`, searchStart)
		inputStart := indexFrom(remaining, `{"input":`, searchStart)
		stopStart := indexFrom(remaining, `{"stop":`, searchStart)

		// Find earliest valid JSON pattern
		candidates := []int{contentStart, nameStart, followupStart, inputStart, stopStart}
		jsonStart := -1
		for _, pos := range candidates {
			if pos >= 0 && (jsonStart < 0 || pos < jsonStart) {
				jsonStart = pos
			}
		}

		if jsonStart < 0 {
			break
		}

		// Find matching closing brace
		jsonEnd := FindMatchingBracket(remaining, jsonStart, '{', '}')
		if jsonEnd < 0 {
			// Incomplete JSON, keep in buffer
			remaining = remaining[jsonStart:]
			break
		}

		// Extract and parse JSON
		jsonStr := remaining[jsonStart : jsonEnd+1]
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
			// Process different event types
			if content, ok := parsed["content"].(string); ok && parsed["followupPrompt"] == nil {
				events = append(events, models.StreamEvent{
					Type:    "content",
					Content: content,
				})
			} else if name, ok := parsed["name"].(string); ok {
				if toolUseID, ok := parsed["toolUseId"].(string); ok {
					toolUse := &models.ToolUse{
						Name:      name,
						ToolUseID: toolUseID,
						Input:     make(map[string]interface{}),
					}
					if input, ok := parsed["input"].(map[string]interface{}); ok {
						toolUse.Input = input
					} else if inputStr, ok := parsed["input"].(string); ok {
						toolUse.Input["input"] = inputStr
					}
					events = append(events, models.StreamEvent{
						Type:    "toolUse",
						ToolUse: toolUse,
					})
				}
			} else if input, ok := parsed["input"]; ok && parsed["name"] == nil {
				inputStr := ""
				if str, ok := input.(string); ok {
					inputStr = str
				} else if inputMap, ok := input.(map[string]interface{}); ok {
					inputJSON, _ := json.Marshal(inputMap)
					inputStr = string(inputJSON)
				}
				events = append(events, models.StreamEvent{
					Type:  "toolUseInput",
					Input: inputStr,
				})
			} else if stop, ok := parsed["stop"].(bool); ok {
				events = append(events, models.StreamEvent{
					Type: "toolUseStop",
					Stop: stop,
				})
			}
		}

		searchStart = jsonEnd + 1
		if searchStart >= len(remaining) {
			remaining = ""
			break
		}
	}

	// Trim processed portion
	if searchStart > 0 && remaining != "" {
		remaining = remaining[searchStart:]
	}

	return events, remaining
}

// Helper function to find substring from a starting position
func indexFrom(s, substr string, start int) int {
	if start >= len(s) {
		return -1
	}
	idx := strings.Index(s[start:], substr)
	if idx < 0 {
		return -1
	}
	return start + idx
}

// createContentEvent creates a Claude content event
// 创建 Claude 内容事件
func (sp *StreamParser) createContentEvent(content, model string) models.ClaudeStreamEvent {
	return models.ClaudeStreamEvent{
		Type: "content_block_delta",
		Index: 0,
		Delta: &models.ClaudeDelta{
			Type: "text_delta",
			Text: content,
		},
	}
}

// createToolUseEvent creates a Claude tool use event
// 创建 Claude 工具使用事件
func (sp *StreamParser) createToolUseEvent(toolUse *models.ToolUse) models.ClaudeStreamEvent {
	return models.ClaudeStreamEvent{
		Type:  "content_block_start",
		Index: 0,
		ContentBlock: &models.ClaudeContentBlock{
			Type:  "tool_use",
			ID:    toolUse.ToolUseID,
			Name:  toolUse.Name,
			Input: toolUse.Input,
		},
	}
}

// createMessageStopEvent creates a message stop event
// 创建消息结束事件
func (sp *StreamParser) createMessageStopEvent() models.ClaudeStreamEvent {
	return models.ClaudeStreamEvent{
		Type: "message_stop",
	}
}

// ConvertToOpenAIStream converts Claude stream events to OpenAI format
// 将 Claude 流事件转换为 OpenAI 格式
func (sp *StreamParser) ConvertToOpenAIStream(claudeEvents <-chan interface{}, model string) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		messageID := GenerateUUID()
		created := time.Now().Unix()

		for event := range claudeEvents {
			if claudeEvent, ok := event.(models.ClaudeStreamEvent); ok {
				switch claudeEvent.Type {
				case "content_block_delta":
					if claudeEvent.Delta != nil && claudeEvent.Delta.Text != "" {
						chunk := models.ChatCompletionChunk{
							ID:      messageID,
							Object:  "chat.completion.chunk",
							Created: created,
							Model:   model,
							Choices: []models.StreamChoice{
								{
									Index: 0,
									Delta: &models.StreamDelta{
										Content: claudeEvent.Delta.Text,
									},
									FinishReason: nil,
								},
							},
						}
						if data, err := json.Marshal(chunk); err == nil {
							ch <- fmt.Sprintf("data: %s\n\n", string(data))
						}
					}

				case "content_block_start":
					if claudeEvent.ContentBlock != nil && claudeEvent.ContentBlock.Type == "tool_use" {
						argsJSON, _ := json.Marshal(claudeEvent.ContentBlock.Input)
						chunk := models.ChatCompletionChunk{
							ID:      messageID,
							Object:  "chat.completion.chunk",
							Created: created,
							Model:   model,
							Choices: []models.StreamChoice{
								{
									Index: 0,
									Delta: &models.StreamDelta{
										ToolCalls: []models.ToolCall{
											{
												Index: 0,
												ID:    claudeEvent.ContentBlock.ID,
												Type:  "function",
												Function: &models.FunctionCall{
													Name:      claudeEvent.ContentBlock.Name,
													Arguments: string(argsJSON),
												},
											},
										},
									},
									FinishReason: nil,
								},
							},
						}
						if data, err := json.Marshal(chunk); err == nil {
							ch <- fmt.Sprintf("data: %s\n\n", string(data))
						}
					}

				case "message_stop":
					finishReason := "stop"
					chunk := models.ChatCompletionChunk{
						ID:      messageID,
						Object:  "chat.completion.chunk",
						Created: created,
						Model:   model,
						Choices: []models.StreamChoice{
							{
								Index:        0,
								Delta:        &models.StreamDelta{},
								FinishReason: &finishReason,
							},
						},
					}
					if data, err := json.Marshal(chunk); err == nil {
						ch <- fmt.Sprintf("data: %s\n\n", string(data))
					}
					ch <- "data: [DONE]\n\n"
				}
			}
		}
	}()

	return ch
}

// ConvertToClaudeStream converts internal events to Claude SSE format
// 将内部事件转换为 Claude SSE 格式
func (sp *StreamParser) ConvertToClaudeStream(events <-chan interface{}) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		// Send message_start event
		messageID := GenerateUUID()
		messageStart := models.ClaudeStreamEvent{
			Type: "message_start",
			Message: &models.ClaudeMessage{
				ID:      messageID,
				Type:    "message",
				Role:    "assistant",
				Content: []models.ClaudeContentBlock{},
				Model:   DefaultModel,
			},
		}
		if data, err := json.Marshal(messageStart); err == nil {
			ch <- fmt.Sprintf("event: message_start\ndata: %s\n\n", string(data))
		}

		for event := range events {
			if claudeEvent, ok := event.(models.ClaudeStreamEvent); ok {
				var eventType string
				switch claudeEvent.Type {
				case "content_block_delta":
					eventType = "content_block_delta"
				case "content_block_start":
					eventType = "content_block_start"
				case "message_stop":
					eventType = "message_stop"
				default:
					continue
				}

				if data, err := json.Marshal(claudeEvent); err == nil {
					ch <- fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(data))
				}
			}
		}
	}()

	return ch
}
