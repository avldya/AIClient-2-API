package models

// Message represents a chat message
// 聊天消息结构
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or []ContentPart
}

// ContentPart represents a part of message content (text, image, tool_result, tool_use)
// 消息内容块（文本、图片、工具结果、工具使用）
type ContentPart struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	Source    *ImageSource           `json:"source,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	Name      string                 `json:"name,omitempty"`
	ID        string                 `json:"id,omitempty"`
}

// ImageSource represents image data
// 图片数据源
type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"` // Base64 encoded
}

// Tool represents a function tool definition
// 工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ChatCompletionRequest represents OpenAI/Claude chat completion request
// OpenAI/Claude 聊天请求
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"`
	System      string    `json:"system,omitempty"`
}

// CodeWhispererRequest represents the request format for CodeWhisperer API
// CodeWhisperer API 请求格式
type CodeWhispererRequest struct {
	ConversationID      string               `json:"conversationId"`
	UserInputMessage    *UserInputMessage    `json:"userInputMessage"`
	History             []HistoryItem        `json:"history,omitempty"`
	ChatTriggerType     string               `json:"chatTriggerType"`
	DiagnosticContext   *DiagnosticContext   `json:"diagnosticContext,omitempty"`
	SupplementalContext *SupplementalContext `json:"supplementalContext,omitempty"`
	Tools               *ToolsContext        `json:"tools,omitempty"`
}

// UserInputMessage represents a user input message for CodeWhisperer
// CodeWhisperer 用户输入消息
type UserInputMessage struct {
	Content     string       `json:"content"`
	ModelID     string       `json:"modelId"`
	Origin      string       `json:"origin"`
	Images      []ImageData  `json:"images,omitempty"`
	ToolResults []ToolResult `json:"toolResults,omitempty"`
}

// ImageData represents image data for CodeWhisperer
// CodeWhisperer 图片数据
type ImageData struct {
	Format string      `json:"format"`
	Source *ImageBytes `json:"source"`
}

// ImageBytes contains base64 encoded image bytes
// Base64 编码的图片字节
type ImageBytes struct {
	Bytes string `json:"bytes"`
}

// ToolResult represents the result of a tool call
// 工具调用结果
type ToolResult struct {
	Content   []ToolResultContent `json:"content"`
	Status    string              `json:"status"`
	ToolUseID string              `json:"toolUseId"`
}

// ToolResultContent represents content of a tool result
// 工具结果内容
type ToolResultContent struct {
	Text string `json:"text"`
}

// HistoryItem represents a history item in CodeWhisperer request
// CodeWhisperer 历史记录项
type HistoryItem struct {
	UserInputMessage         *UserInputMessage         `json:"userInputMessage,omitempty"`
	AssistantResponseMessage *AssistantResponseMessage `json:"assistantResponseMessage,omitempty"`
}

// AssistantResponseMessage represents assistant response in history
// 历史记录中的助手响应
type AssistantResponseMessage struct {
	Content  string    `json:"content"`
	ToolUses []ToolUse `json:"toolUses,omitempty"`
}

// ToolUse represents a tool use in assistant response
// 助手响应中的工具使用
type ToolUse struct {
	Name      string                 `json:"name"`
	ToolUseID string                 `json:"toolUseId"`
	Input     map[string]interface{} `json:"input"`
}

// DiagnosticContext placeholder (not used in current implementation)
type DiagnosticContext struct{}

// SupplementalContext placeholder (not used in current implementation)
type SupplementalContext struct{}

// ToolsContext represents tools configuration for CodeWhisperer
// CodeWhisperer 工具配置
type ToolsContext struct {
	Tools []ToolSpecification `json:"tools"`
}

// ToolSpecification represents a tool specification
// 工具规格
type ToolSpecification struct {
	ToolSpecification *ToolSpec `json:"toolSpecification"`
}

// ToolSpec represents tool specification details
// 工具规格详情
type ToolSpec struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	InputSchema *InputSchema `json:"inputSchema"`
}

// InputSchema wraps the JSON schema for tool input
// 工具输入的 JSON Schema 包装
type InputSchema struct {
	JSON map[string]interface{} `json:"json"`
}

// Credentials represents OAuth credentials
// OAuth 凭据
type Credentials struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	AuthMethod   string `json:"authMethod"`
	ExpiresAt    string `json:"expiresAt"`
	ProfileArn   string `json:"profileArn,omitempty"`
	Region       string `json:"region,omitempty"`
	UUID         string `json:"uuid,omitempty"`
}

// StreamEvent represents a streaming event from CodeWhisperer
// CodeWhisperer 流式响应事件
type StreamEvent struct {
	Type    string   `json:"type"`
	Content string   `json:"content,omitempty"`
	ToolUse *ToolUse `json:"toolUse,omitempty"`
	Input   string   `json:"input,omitempty"`
	Stop    bool     `json:"stop,omitempty"`
}

// ChatCompletionChunk represents a streaming response chunk (OpenAI format)
// 流式响应块（OpenAI 格式）
type ChatCompletionChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice represents a choice in streaming response
// 流式响应中的选择
type StreamChoice struct {
	Index        int          `json:"index"`
	Delta        *StreamDelta `json:"delta"`
	FinishReason *string      `json:"finish_reason"`
}

// StreamDelta represents the delta in streaming response
// 流式响应中的增量
type StreamDelta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool call in OpenAI format
// OpenAI 格式的工具调用
type ToolCall struct {
	Index    int           `json:"index,omitempty"`
	ID       string        `json:"id,omitempty"`
	Type     string        `json:"type,omitempty"`
	Function *FunctionCall `json:"function,omitempty"`
}

// FunctionCall represents a function call details
// 函数调用详情
type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// ClaudeStreamEvent represents Claude streaming event
// Claude 流式事件
type ClaudeStreamEvent struct {
	Type         string              `json:"type"`
	Message      *ClaudeMessage      `json:"message,omitempty"`
	Index        int                 `json:"index,omitempty"`
	ContentBlock *ClaudeContentBlock `json:"content_block,omitempty"`
	Delta        *ClaudeDelta        `json:"delta,omitempty"`
	Usage        *ClaudeUsage        `json:"usage,omitempty"`
}

// ClaudeMessage represents a Claude message
// Claude 消息
type ClaudeMessage struct {
	ID           string               `json:"id"`
	Type         string               `json:"type"`
	Role         string               `json:"role"`
	Content      []ClaudeContentBlock `json:"content"`
	Model        string               `json:"model"`
	StopReason   *string              `json:"stop_reason,omitempty"`
	StopSequence *string              `json:"stop_sequence,omitempty"`
	Usage        *ClaudeUsage         `json:"usage,omitempty"`
}

// ClaudeContentBlock represents a content block in Claude format
// Claude 格式的内容块
type ClaudeContentBlock struct {
	Type  string                 `json:"type"`
	Text  string                 `json:"text,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

// ClaudeDelta represents a delta in Claude streaming
// Claude 流式响应中的增量
type ClaudeDelta struct {
	Type        string `json:"type,omitempty"`
	Text        string `json:"text,omitempty"`
	StopReason  string `json:"stop_reason,omitempty"`
	PartialJSON string `json:"partial_json,omitempty"`
}

// ClaudeUsage represents token usage information
// Token 使用信息
type ClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
