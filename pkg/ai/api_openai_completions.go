package ai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAICompletionsProvider OpenAI Completions 提供者
type OpenAICompletionsProvider struct {
	client *http.Client
}

// NewOpenAICompletionsApiProvider 创建新的提供者
func NewOpenAICompletionsApiProvider() *OpenAICompletionsProvider {
	return &OpenAICompletionsProvider{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *OpenAICompletionsProvider) GetAPI() string {
	return ApiOpenAICompletions
}

// Stream 流式调用
func (p *OpenAICompletionsProvider) Stream(
	model Model,
	ctx Context,
	opts *StreamOptions,
) *AssistantMessageEventStream {
	stream := NewAssistantMessageEventStream()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				stream.Push(NewAssistantMessageEventError(
					StopReasonError,
					&AssistantMessage{
						StopReason:   StopReasonError,
						ErrorMessage: fmt.Sprintf("panic: %v", r),
						Timestamp:    time.Now().Unix(),
					},
				))
				stream.End(nil)
			}
		}()

		output := &AssistantMessage{
			Role:       MessageRoleAssistant,
			Content:    []ContentBlock{},
			API:        model.GetAPI(),
			Provider:   model.GetProvider(),
			Model:      model.GetID(),
			Usage:      Usage{},
			StopReason: StopReasonStop,
			Timestamp:  time.Now().UnixMilli(),
		}

		if err := p.doStream(model, ctx, opts, stream, output); err != nil {
			output.StopReason = StopReasonError
			output.ErrorMessage = err.Error()
			stream.Push(NewAssistantMessageEventError(
				output.StopReason,
				output,
			))
			stream.End(output)
		} else {
			stream.Push(NewAssistantMessageEventDone(
				output.StopReason,
				output,
			))
			stream.End(output)
		}
	}()

	return stream
}

func (p *OpenAICompletionsProvider) StreamSimple(
	model Model,
	ctx Context,
	opts *SimpleStreamOptions,
) *AssistantMessageEventStream {
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = GetEnvApiKey(model.GetProvider())
	}
	streamOptions := NewStreamOptions(
		apiKey,
		opts.Headers,
		opts.MaxTokens,
		opts.Temperature,
		opts.ReasoningEffort,
	)
	return p.Stream(model, ctx, streamOptions)
}

// doStream 执行流式请求
func (p *OpenAICompletionsProvider) doStream(
	model Model,
	ctx Context,
	opts *StreamOptions,
	stream *AssistantMessageEventStream,
	output *AssistantMessage,
) error {
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = GetEnvApiKey(model.GetProvider())
	}

	req, err := p.buildRequest(model, ctx, opts, apiKey)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return p.processStream(resp.Body, stream, output, model)
}

// buildRequest 构建 HTTP 请求
func (p *OpenAICompletionsProvider) buildRequest(
	model Model,
	ctx Context,
	opts *StreamOptions,
	apiKey string,
) (*http.Request, error) {
	params := p.buildParams(model, ctx, opts)

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}

	req, err := http.NewRequest("POST", model.GetBaseURL()+"/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 合并模型头部
	for k, v := range model.GetHeaders() {
		req.Header.Set(k, v)
	}

	// 合并选项头部
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// 提供商特定 headers
	p.applyProviderHeaders(model, ctx, req)

	return req, nil
}

// applyProviderHeaders 应用提供商特定的 headers
func (p *OpenAICompletionsProvider) applyProviderHeaders(model Model, ctx Context, req *http.Request) {
	// GitHub Copilot 动态 headers
	if model.GetProvider() == "github-copilot" {
		p.applyGitHubCopilotHeaders(ctx, req)
	}

	// OpenRouter 缓存控制
	if strings.Contains(model.GetBaseURL(), "openrouter.ai") {
		p.applyOpenRouterHeaders(ctx, req)
	}
}

// applyGitHubCopilotHeaders 应用 GitHub Copilot 动态 headers
func (p *OpenAICompletionsProvider) applyGitHubCopilotHeaders(ctx Context, req *http.Request) {
	// 检查是否有图片输入
	hasImages := false
	for _, msg := range ctx.Messages {
		if userMsg, ok := msg.(*UserMessage); ok {
			if blocks, ok := userMsg.Content.([]ContentBlock); ok {
				for _, block := range blocks {
					if block.GetType() == ContentBlockTypeImage {
						hasImages = true
						break
					}
				}
			}
		}
		if hasImages {
			break
		}
	}

	// 设置 Copilot 特定的 headers
	req.Header.Set("editor-version", "vscode/1.90.0")
	req.Header.Set("editor-plugin-version", "copilot/1.200.0")

	if hasImages {
		req.Header.Set("copilot-vision-request", "true")
	}
}

// applyOpenRouterHeaders 应用 OpenRouter 缓存控制 headers
func (p *OpenAICompletionsProvider) applyOpenRouterHeaders(ctx Context, req *http.Request) {
	// OpenRouter 支持 Anthropic 风格的缓存控制
	// 在 system prompt 或最后一条消息上添加缓存控制
	if ctx.SystemPrompt != "" {
		// OpenRouter 通过 x-cache-key header 支持缓存
		req.Header.Set("x-cache-key", "system-prompt")
	}
}

// buildParams 构建请求参数
func (p *OpenAICompletionsProvider) buildParams(
	model Model,
	ctx Context,
	opts *StreamOptions,
) map[string]any {
	params := map[string]any{
		"model":    model.GetID(),
		"stream":   true,
		"messages": p.convertMessages(model, ctx),
		"stream_options": map[string]any{
			"include_usage": true,
		},
	}

	if opts.MaxTokens > 0 {
		params["max_completion_tokens"] = opts.MaxTokens
	}

	if opts.Temperature != nil {
		params["temperature"] = *opts.Temperature
	}

	// 工具处理：如果有工具定义或对话历史中有工具使用
	if len(ctx.Tools) > 0 {
		params["tools"] = p.convertTools(ctx.Tools)
	} else if p.hasToolHistory(ctx.Messages) {
		// Anthropic 等提供商要求在有工具使用时必须包含 tools 参数
		params["tools"] = []map[string]any{}
	}

	// 推理模型支持
	if model.GetReasoning() {
		// 检查是否支持 enable_thinking 参数（Z.ai/Qwen）
		if p.supportsEnableThinking(model) {
			params["enable_thinking"] = opts.ReasoningEffort != "" && opts.ReasoningEffort != "none"
		} else if opts.ReasoningEffort != "" {
			// OpenAI-style reasoning_effort
			params["reasoning_effort"] = p.mapReasoningEffort(opts.ReasoningEffort)
		}
	}

	return params
}

// hasToolHistory 检测对话中是否有工具使用
func (p *OpenAICompletionsProvider) hasToolHistory(messages []Message) bool {
	for _, msg := range messages {
		if msg.GetRole() == "toolResult" {
			return true
		}
		if msg.GetRole() == MessageRoleAssistant {
			if assistantMsg, ok := msg.(*AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if block.GetType() == ContentBlockTypeToolCall {
						return true
					}
				}
			}
		}
	}
	return false
}

// supportsEnableThinking 检查是否支持 enable_thinking 参数
func (p *OpenAICompletionsProvider) supportsEnableThinking(model Model) bool {
	// Z.ai 和 Qwen 使用 enable_thinking 参数
	baseURL := model.GetBaseURL()
	if strings.Contains(baseURL, "z.ai") || strings.Contains(baseURL, "zhipu") {
		return true
	}
	if strings.Contains(baseURL, "qwen") {
		return true
	}
	return false
}

// mapReasoningEffort 映射推理努力程度
func (p *OpenAICompletionsProvider) mapReasoningEffort(effort string) string {
	// OpenAI 标准映射
	switch effort {
	case "minimal":
		return "minimal"
	case "low":
		return "low"
	case "medium":
		return "medium"
	case "high":
		return "high"
	default:
		return effort
	}
}

// convertMessages 转换消息格式
func (p *OpenAICompletionsProvider) convertMessages(model Model, ctx Context) []map[string]any {
	var messages []map[string]any

	// 系统提示词：推理模型使用 developer 角色
	if ctx.SystemPrompt != "" {
		role := "system"
		if model.GetReasoning() && p.supportsDeveloperRole(model) {
			role = "developer"
		}
		messages = append(messages, map[string]any{
			"role":    role,
			"content": ctx.SystemPrompt,
		})
	}

	lastRole := ""
	for _, msg := range ctx.Messages {
		switch m := msg.(type) {
		case *UserMessage:
			// 某些提供商不允许 user 消息紧跟在 tool result 后面
			if p.requiresAssistantAfterToolResult(model) && lastRole == "toolResult" {
				messages = append(messages, map[string]any{
					"role":    MessageRoleAssistant,
					"content": "I have processed the tool results.",
				})
			}
			messages = append(messages, p.convertUserMessage(m, model))
			lastRole = MessageRoleUser
		case *AssistantMessage:
			messages = append(messages, p.convertAssistantMessage(m))
			lastRole = MessageRoleAssistant
		case *ToolResultMessage:
			messages = append(messages, p.convertToolResultMessage(m))
			lastRole = "toolResult"
		}
	}

	return messages
}

// supportsDeveloperRole 检查是否支持 developer 角色
func (p *OpenAICompletionsProvider) supportsDeveloperRole(model Model) bool {
	// OpenAI o1/o3 系列使用 developer 角色
	baseURL := model.GetBaseURL()
	if strings.Contains(baseURL, "openai.com") {
		modelID := model.GetID()
		if strings.HasPrefix(modelID, "o1") || strings.HasPrefix(modelID, "o3") {
			return true
		}
	}
	return false
}

// requiresAssistantAfterToolResult 检查是否需要在 tool result 后插入 assistant 消息
func (p *OpenAICompletionsProvider) requiresAssistantAfterToolResult(model Model) bool {
	// Anthropic 等提供商需要
	baseURL := model.GetBaseURL()
	if strings.Contains(baseURL, "anthropic") {
		return true
	}
	return false
}

// convertUserMessage 转换用户消息
func (p *OpenAICompletionsProvider) convertUserMessage(msg *UserMessage, model Model) map[string]any {
	if content, ok := msg.Content.(string); ok {
		return map[string]any{
			"role":    MessageRoleUser,
			"content": content,
		}
	}

	// 处理多模态内容
	var contentParts []map[string]any
	for _, block := range msg.Content.([]ContentBlock) {
		switch b := block.(type) {
		case *TextContentBlock:
			contentParts = append(contentParts, map[string]any{
				"type": "text",
				"text": b.Text,
			})
		case *ImageContentBlock:
			// 如果模型不支持图片，跳过
			if !p.supportsImageInput(model) {
				continue
			}
			contentParts = append(contentParts, map[string]any{
				"type": "image_url",
				"image_url": map[string]any{
					"url": fmt.Sprintf("data:%s;base64,%s", b.MimeType, b.Data),
				},
			})
		}
	}

	// 如果没有内容，返回 nil
	if len(contentParts) == 0 {
		return nil
	}

	return map[string]any{
		"role":    MessageRoleUser,
		"content": contentParts,
	}
}

// supportsImageInput 检查是否支持图片输入
func (p *OpenAICompletionsProvider) supportsImageInput(model Model) bool {
	input := model.GetInput()
	for _, i := range input {
		if i == "image" {
			return true
		}
	}
	return false
}

// convertAssistantMessage 转换助手消息
func (p *OpenAICompletionsProvider) convertAssistantMessage(msg *AssistantMessage) map[string]any {
	var content string
	var toolCalls []map[string]any

	for _, block := range msg.Content {
		switch b := block.(type) {
		case *TextContentBlock:
			content += b.Text
		case *ToolCallContentBlock:
			// 规范化 Tool call ID
			normalizedID := p.normalizeToolCallID(b.ID)

			// 将 Arguments 转换为 JSON 字符串
			var argsJSON string
			if b.Arguments != nil {
				if bytes, err := json.Marshal(b.Arguments); err == nil {
					argsJSON = string(bytes)
				}
			}

			toolCalls = append(toolCalls, map[string]any{
				"id":   normalizedID,
				"type": "function",
				"function": map[string]any{
					"name":      b.Name,
					"arguments": argsJSON,
				},
			})
		}
	}

	result := map[string]any{
		"role":    MessageRoleAssistant,
		"content": content, // 始终设置 content，即使是空字符串
	}

	if len(toolCalls) > 0 {
		result["tool_calls"] = toolCalls
	}

	return result
}

// normalizeToolCallID 规范化 Tool call ID
func (p *OpenAICompletionsProvider) normalizeToolCallID(id string) string {
	// 处理 OpenAI Responses API 的 pipe 分隔 ID
	// 格式：{call_id}|{id}，其中 {id} 可能包含特殊字符
	if strings.Contains(id, "|") {
		parts := strings.Split(id, "|")
		id = parts[0]
	}

	// OpenAI 限制 ID 长度为 40 字符
	if len(id) > 40 {
		id = id[:40]
	}

	return id
}

// convertToolResultMessage 转换工具结果消息
func (p *OpenAICompletionsProvider) convertToolResultMessage(msg *ToolResultMessage) map[string]any {
	// 将 ContentBlock 数组转换为字符串
	var contentStr string
	for _, block := range msg.Content {
		if tc, ok := block.(*TextContentBlock); ok {
			contentStr += tc.Text
		}
	}

	return map[string]any{
		"role":         "tool",
		"tool_call_id": msg.ToolCallID,
		"content":      contentStr,
	}
}

// convertTools 转换工具定义
func (p *OpenAICompletionsProvider) convertTools(tools []Tool) []map[string]any {
	var result []map[string]any
	for _, tool := range tools {
		result = append(result, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		})
	}
	return result
}

// processStream 处理流式响应
func (p *OpenAICompletionsProvider) processStream(
	reader io.Reader,
	stream *AssistantMessageEventStream,
	output *AssistantMessage,
	model Model,
) error {
	// 使用 SSE 解析器
	parser := NewSSEParser(reader)

	var currentBlock ContentBlock
	var blocks []ContentBlock

	// 发送开始事件
	// stream.Push(&AssistantMessageEventStart{
	// 	Type:    AssistantMessageEventTypeStart,
	// 	Partial: output,
	// })

	for {
		event, err := parser.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("parse stream: %w", err)
		}

		if event.Data == "[DONE]" {
			break
		}

		var chunk ChatCompletionChunk
		if err := json.Unmarshal([]byte(event.Data), &chunk); err != nil {
			continue // 忽略解析错误
		}

		// 处理 usage
		if chunk.Usage != nil {
			output.Usage = parseChunkUsage(chunk.Usage, model)
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]

		// 处理停止原因
		if choice.FinishReason != "" {
			output.StopReason = p.mapStopReason(choice.FinishReason)
		}

		// 处理内容增量
		if choice.Delta.Content != "" {
			if currentBlock == nil || currentBlock.GetType() != ContentBlockTypeText {
				// 结束当前块
				if currentBlock != nil {
					p.finishBlock(stream, currentBlock, len(blocks)-1, output)
				}

				// 开始新块
				currentBlock = NewTextContentBlock("")
				blocks = append(blocks, currentBlock)
				output.Content = blocks

				stream.Push(NewAssistantMessageEventTextStart(
					len(blocks)-1,
					output,
				))
			}

			if textBlock, ok := currentBlock.(*TextContentBlock); ok {
				textBlock.Text += choice.Delta.Content
				stream.Push(NewAssistantMessageEventTextDelta(
					len(blocks)-1,
					choice.Delta.Content,
					output,
				))
			}
		}

		// Some endpoints return reasoning in reasoning_content (llama.cpp),
		// or reasoning (other openai compatible endpoints)
		// Use the first non-empty reasoning field to avoid duplication
		// (e.g., chutes.ai returns both reasoning_content and reasoning with same content)
		var foundReasoningField *string
		var reasoningContent string
		if choice.Delta.ReasoningContent != "" {
			foundReasoningField = &choice.Delta.ReasoningContent
			reasoningContent = choice.Delta.ReasoningContent
		} else if choice.Delta.Reasoning != "" {
			foundReasoningField = &choice.Delta.Reasoning
			reasoningContent = choice.Delta.Reasoning
		} else if choice.Delta.ReasoningText != "" {
			foundReasoningField = &choice.Delta.ReasoningText
			reasoningContent = choice.Delta.ReasoningText
		}

		// 处理思考内容
		if foundReasoningField != nil {
			if currentBlock == nil || currentBlock.GetType() != ContentBlockTypeThinking {
				if currentBlock != nil {
					p.finishBlock(stream, currentBlock, len(blocks)-1, output)
				}
				currentBlock = NewThinkingContentBlock(
					"",
					*foundReasoningField,
				)
				blocks = append(blocks, currentBlock)
				output.Content = blocks

				stream.Push(NewAssistantMessageEventThinkingStart(
					len(blocks)-1,
					output,
				))
			}

			if thinkingBlock, ok := currentBlock.(*ThinkingContentBlock); ok {
				thinkingBlock.Thinking += reasoningContent
				stream.Push(NewAssistantMessageEventThinkingDelta(
					len(blocks)-1,
					reasoningContent,
					output,
				))
			}
		}

		// 处理工具调用
		for _, tCall := range choice.Delta.ToolCalls {
			if currentBlock == nil || currentBlock.GetType() != ContentBlockTypeToolCall {
				if currentBlock != nil {
					p.finishBlock(stream, currentBlock, len(blocks)-1, output)
				}

				currentBlock = NewToolCallContentBlock(
					tCall.ID,
					tCall.Function.Name,
					nil,
				)
				blocks = append(blocks, currentBlock)
				output.Content = blocks

				stream.Push(NewAssistantMessageEventToolCallStart(
					len(blocks)-1,
					output,
				))
			}

			if tc, ok := currentBlock.(*ToolCallContentBlock); ok {
				if tCall.ID != "" {
					tc.ID = tCall.ID
				}
				if tCall.Function.Name != "" {
					tc.Name = tCall.Function.Name
				}
				if tCall.Function.Arguments != "" {
					tc.Arguments = p.parseStreamingJSON(tCall.Function.Arguments)
				}

				stream.Push(NewAssistantMessageEventToolCallDelta(
					len(blocks)-1,
					tCall.Function.Arguments,
					output,
				))
			}
		}
	}

	// 结束最后一个块
	if currentBlock != nil {
		p.finishBlock(stream, currentBlock, len(blocks)-1, output)
	}

	return nil
}

// finishBlock 结束内容块
func (p *OpenAICompletionsProvider) finishBlock(
	stream *AssistantMessageEventStream,
	block ContentBlock,
	index int,
	output *AssistantMessage,
) {
	switch b := block.(type) {
	case *TextContentBlock:
		stream.Push(NewAssistantMessageEventTextEnd(
			index,
			b.Text,
			output,
		))
	case *ThinkingContentBlock:
		stream.Push(NewAssistantMessageEventThinkingEnd(
			index,
			b.Thinking,
			output,
		))
	case *ToolCallContentBlock:
		stream.Push(NewAssistantMessageEventToolCallEnd(
			index,
			b,
			output,
		))
	}
}

// mapStopReason 映射停止原因
func (p *OpenAICompletionsProvider) mapStopReason(reason string) StopReason {
	switch reason {
	case "stop":
		return StopReasonStop
	case "length":
		return StopReasonLength
	case "tool_calls":
		return StopReasonToolUse
	case "content_filter":
		return StopReasonError
	default:
		return StopReasonStop
	}
}

// parseStreamingJSON 解析流式 JSON
func (p *OpenAICompletionsProvider) parseStreamingJSON(data string) map[string]any {
	var result map[string]any
	json.Unmarshal([]byte(data), &result)
	return result
}

// 确保 OpenAICompletionsProvider 实现 ApiProvider 接口
var _ ApiProvider = (*OpenAICompletionsProvider)(nil)

type ChatCompletionChunkUsage struct {
	PromptTokens        int `json:"prompt_tokens"`
	PromptTokensDetails *struct {
		AudioTokens  int `json:"audio_tokens"`
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
	CompletionTokens        int `json:"completion_tokens"`
	CompletionTokensDetails *struct {
		AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
		AudioTokens              int `json:"audio_tokens"`
		ReasoningTokens          int `json:"reasoning_tokens"`
		RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
	} `json:"completion_tokens_details"`
	TotalTokens  int `json:"total_tokens"`
	CachedTokens int `json:"cached_tokens"`
}

// ChatCompletionChunk OpenAI 流式响应块
type ChatCompletionChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role             string `json:"role"`
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
			Reasoning        string `json:"reasoning"`
			ReasoningText    string `json:"reasoning_text"`
			ToolCalls        []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *ChatCompletionChunkUsage `json:"usage"`
}

// SSEParser SSE 解析器
type SSEParser struct {
	reader *bufio.Reader
}

func NewSSEParser(reader io.Reader) *SSEParser {
	return &SSEParser{reader: bufio.NewReader(reader)}
}

func (p *SSEParser) Next() (*SSEEvent, error) {
	for {
		line, err := p.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			return &SSEEvent{Data: data}, nil
		}
	}
}

type SSEEvent struct {
	Data string
}

func parseChunkUsage(rawUsage *ChatCompletionChunkUsage, model Model) Usage {
	var cachedTokens int
	if promptTokensDetails := rawUsage.PromptTokensDetails; promptTokensDetails != nil {
		cachedTokens = promptTokensDetails.CachedTokens
	} else {
		cachedTokens = 0
	}
	var reasoningTokens int
	if completionTokensDetails := rawUsage.CompletionTokensDetails; completionTokensDetails != nil {
		reasoningTokens = completionTokensDetails.ReasoningTokens
	} else {
		reasoningTokens = 0
	}
	var outputTokens int
	if completionTokensDetails := rawUsage.CompletionTokensDetails; completionTokensDetails != nil {
		outputTokens = completionTokensDetails.AcceptedPredictionTokens + completionTokensDetails.RejectedPredictionTokens
	} else {
		// 如果没有 CompletionTokensDetails，直接使用 CompletionTokens
		outputTokens = rawUsage.CompletionTokens
	}
	inputTokens := rawUsage.PromptTokens - cachedTokens
	usage := Usage{
		Input:       inputTokens,
		Output:      outputTokens + reasoningTokens,
		CacheRead:   cachedTokens,
		CacheWrite:  0,
		TotalTokens: inputTokens + outputTokens + cachedTokens,
	}
	usage.Cost = CalculateCost(model, &usage)
	return usage
}
