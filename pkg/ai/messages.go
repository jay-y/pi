package ai

import "time"

// ContentBlock 内容块接口
type ContentBlock interface {
	GetType() ContentBlockType
}

// TextContent 文本内容块
type TextContentBlock struct {
	Type          ContentBlockType `json:"type"`
	Text          string           `json:"text"`
	TextSignature string           `json:"textSignature,omitempty"`
}

func (c *TextContentBlock) GetType() ContentBlockType {
	return c.Type
}

// NewTextContent 创建新的文本内容块
func NewTextContentBlock(text string) *TextContentBlock {
	return &TextContentBlock{
		Type: ContentBlockTypeText,
		Text: text,
	}
}

// ThinkingContent 思考内容块
type ThinkingContentBlock struct {
	Type              ContentBlockType `json:"type"`
	Thinking          string           `json:"thinking"`
	ThinkingSignature string           `json:"thinkingSignature,omitempty"`
	Redacted          bool             `json:"redacted,omitempty"`
}

func (c *ThinkingContentBlock) GetType() ContentBlockType {
	return c.Type
}

// NewThinkingContent 创建新的思考内容块
func NewThinkingContentBlock(thinking string, signature string) *ThinkingContentBlock {
	return &ThinkingContentBlock{
		Type:              ContentBlockTypeThinking,
		Thinking:          thinking,
		ThinkingSignature: signature,
	}
}

// ImageContent 图片内容块
type ImageContentBlock struct {
	Type     ContentBlockType `json:"type"`
	Data     string           `json:"data"`     // base64 编码的图片数据
	MimeType string           `json:"mimeType"` // 例如 "image/jpeg", "image/png"
}

func (c *ImageContentBlock) GetType() ContentBlockType {
	return c.Type
}

// NewImageContent 创建新的图片内容块
func NewImageContentBlock(data, mimeType string) *ImageContentBlock {
	return &ImageContentBlock{
		Type:     ContentBlockTypeImage,
		Data:     data,
		MimeType: mimeType,
	}
}

// ToolCallContentBlock 工具调用内容块
type ToolCallContentBlock struct {
	Type             ContentBlockType `json:"type"`
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Arguments        any              `json:"arguments"`
	ThoughtSignature *string          `json:"thoughtSignature,omitempty"`
}

func (c *ToolCallContentBlock) GetType() ContentBlockType {
	return c.Type
}

// NewToolCallContentBlock 创建新的工具调用内容块
func NewToolCallContentBlock(id, name string, arguments map[string]any) *ToolCallContentBlock {
	return &ToolCallContentBlock{
		Type:      ContentBlockTypeToolCall,
		ID:        id,
		Name:      name,
		Arguments: &arguments,
	}
}

// Message 消息接口
type Message interface {
	GetRole() string
	GetTimestamp() int64
}

// UserMessage 用户消息
type UserMessage struct {
	Role      string `json:"role"`
	Content   any    `json:"content"`   // string 或 []ContentBlock
	Timestamp int64  `json:"timestamp"` // Unix 毫秒时间戳
}

func (u *UserMessage) GetRole() string {
	return u.Role
}

func (u *UserMessage) GetTimestamp() int64 {
	return u.Timestamp
}

// NewUserMessage 创建新的用户消息
func NewUserMessage(content any) *UserMessage {
	return &UserMessage{
		Role:      MessageRoleUser,
		Content:   content,
		Timestamp: time.Now().UnixMilli(),
	}
}

// AssistantMessage 助手消息
type AssistantMessage struct {
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	API          string         `json:"api"`
	Provider     string         `json:"provider"`
	Model        string         `json:"model"`
	Usage        Usage          `json:"usage"`
	StopReason   StopReason     `json:"stopReason"`
	ErrorMessage string         `json:"errorMessage,omitempty"`
	Timestamp    int64          `json:"timestamp"` // Unix 毫秒时间戳
}

func (a *AssistantMessage) GetRole() string {
	return a.Role
}

func (a *AssistantMessage) GetTimestamp() int64 {
	return a.Timestamp
}

// NewAssistantMessage 创建新的助手消息
func NewAssistantMessage(api string, provider string, model string) *AssistantMessage {
	return &AssistantMessage{
		Role:       MessageRoleAssistant,
		Content:    []ContentBlock{},
		API:        api,
		Provider:   provider,
		Model:      model,
		StopReason: StopReasonStop,
		Timestamp:  time.Now().UnixMilli(),
	}
}

// ToolResultMessage 工具结果消息
type ToolResultMessage struct {
	Role       string         `json:"role"`
	ToolCallID string         `json:"toolCallId"`
	ToolName   string         `json:"toolName"`
	Content    []ContentBlock `json:"content"` // 支持文本和图片
	Details    any            `json:"details,omitempty"`
	IsError    bool           `json:"isError"`
	Timestamp  int64          `json:"timestamp"` // Unix 毫秒时间戳
}

func (t *ToolResultMessage) GetRole() string {
	return t.Role
}

func (t *ToolResultMessage) GetTimestamp() int64 {
	return t.Timestamp
}

// NewToolResultMessage 创建新的工具结果消息
func NewToolResultMessage(toolCallID, toolName string, content []ContentBlock, isError bool) *ToolResultMessage {
	return &ToolResultMessage{
		Role:       MessageRoleToolResult,
		ToolCallID: toolCallID,
		ToolName:   toolName,
		Content:    content,
		IsError:    isError,
		Timestamp:  time.Now().UnixMilli(),
	}
}
