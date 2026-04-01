package ai

// ModelApi 调用的 API 接口

const (
	ApiOpenAICompletions     string = "openai-completions"
	ApiOpenAIResponses       string = "openai-responses"
	ApiAzureOpenAIResponses  string = "azure-openai-responses"
	ApiOpenAICodexResponses  string = "openai-codex-responses"
	ApiAnthropicMessages     string = "anthropic-messages"
	ApiBedrockConverseStream string = "bedrock-converse-stream"
	ApiGoogleGenerativeAI    string = "google-generative-ai"
	ApiGoogleGeminiCLI       string = "google-gemini-cli"
	ApiGoogleVertex          string = "google-vertex"
)

// ModelProvider 模型供应商

const (
	ProviderAmazonBedrock        string = "amazon-bedrock"
	ProviderAnthropic            string = "anthropic"
	ProviderGoogle               string = "google"
	ProviderGoogleGeminiCLI      string = "google-gemini-cli"
	ProviderGoogleAntigravity    string = "google-antigravity"
	ProviderGoogleVertex         string = "google-vertex"
	ProviderOpenAI               string = "openai"
	ProviderAzureOpenAIResponses string = "azure-openai-responses"
	ProviderOpenAICodex          string = "openai-codex"
	ProviderGitHubCopilot        string = "github-copilot"
	ProviderXAI                  string = "xai"
	ProviderGroq                 string = "groq"
	ProviderCerebras             string = "cerebras"
	ProviderOpenRouter           string = "openrouter"
	ProviderVercelAIGateway      string = "vercel-ai-gateway"
	ProviderZAI                  string = "zai"
	ProviderMistral              string = "mistral"
	ProviderMinimax              string = "minimax"
	ProviderMinimaxCN            string = "minimax-cn"
	ProviderHuggingFace          string = "huggingface"
	ProviderOpenCode             string = "opencode"
	ProviderKimiCoding           string = "kimi-coding"
)

// ThinkingLevel 思考级别
type ThinkingLevel string

const (
	ThinkingLevelOff     ThinkingLevel = "off"
	ThinkingLevelMinimal ThinkingLevel = "minimal"
	ThinkingLevelLow     ThinkingLevel = "low"
	ThinkingLevelMedium  ThinkingLevel = "medium"
	ThinkingLevelHigh    ThinkingLevel = "high"
	ThinkingLevelXHigh   ThinkingLevel = "xhigh"
)

// ThinkingLevels 标准思考级别
var ThinkingLevels = []ThinkingLevel{
	ThinkingLevelOff,
	ThinkingLevelMinimal,
	ThinkingLevelLow,
	ThinkingLevelMedium,
	ThinkingLevelHigh,
}

// ThinkingLevelsWithXHigh 包含 xhigh 的思考级别（用于支持的模型）
var ThinkingLevelsWithXHigh = []ThinkingLevel{
	ThinkingLevelOff,
	ThinkingLevelMinimal,
	ThinkingLevelLow,
	ThinkingLevelMedium,
	ThinkingLevelHigh,
	ThinkingLevelXHigh,
}

// CacheRetention 缓存保留时间
type CacheRetention string

const (
	CacheRetentionNone  CacheRetention = "none"
	CacheRetentionShort CacheRetention = "short"
	CacheRetentionLong  CacheRetention = "long"
)

// Transport 传输协议
type Transport string

const (
	TransportSSE       Transport = "sse"
	TransportWebSocket Transport = "websocket"
	TransportAuto      Transport = "auto"
)

// StopReason 停止原因
type StopReason string

const (
	StopReasonStop    StopReason = "stop"
	StopReasonLength  StopReason = "length"
	StopReasonToolUse StopReason = "toolUse"
	StopReasonError   StopReason = "error"
	StopReasonAborted StopReason = "aborted"
)

// AssistantMessageEvent 助手消息事件类型
type AssistantMessageEventType string

const (
	AssistantMessageEventTypeStart         AssistantMessageEventType = "start"
	AssistantMessageEventTypeTextStart     AssistantMessageEventType = "text_start"
	AssistantMessageEventTypeTextDelta     AssistantMessageEventType = "text_delta"
	AssistantMessageEventTypeTextEnd       AssistantMessageEventType = "text_end"
	AssistantMessageEventTypeThinkingStart AssistantMessageEventType = "thinking_start"
	AssistantMessageEventTypeThinkingDelta AssistantMessageEventType = "thinking_delta"
	AssistantMessageEventTypeThinkingEnd   AssistantMessageEventType = "thinking_end"
	AssistantMessageEventTypeToolCallStart AssistantMessageEventType = "toolcall_start"
	AssistantMessageEventTypeToolCallDelta AssistantMessageEventType = "toolcall_delta"
	AssistantMessageEventTypeToolCallEnd   AssistantMessageEventType = "toolcall_end"
	AssistantMessageEventTypeDone          AssistantMessageEventType = "done"
	AssistantMessageEventTypeError         AssistantMessageEventType = "error"
)

// ContentBlockType 内容块类型
type ContentBlockType string

const (
	ContentBlockTypeText     ContentBlockType = "text"
	ContentBlockTypeThinking ContentBlockType = "thinking"
	ContentBlockTypeToolCall ContentBlockType = "toolCall"
	ContentBlockTypeImage    ContentBlockType = "image"
)

// MessageRole 消息角色

const (
	MessageRoleUser       string = "user"
	MessageRoleAssistant  string = "assistant"
	MessageRoleSystem     string = "system"
	MessageRoleCustom     string = "custom"
	MessageRoleToolResult string = "toolResult"
)
