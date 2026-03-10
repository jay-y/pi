package agent

import "github.com/jay-y/pi/pkg/ai"

// AgentEvent 代理事件
type AgentEvent interface {
	GetType() AgentEventType
}

// AgentEventStart 代理开始事件
type AgentEventStart struct {
	Type AgentEventType `json:"type"`
}

func (e *AgentEventStart) GetType() AgentEventType {
	return e.Type
}

func NewAgentEventStart() *AgentEventStart {
	return &AgentEventStart{
		Type: AgentEventTypeStart,
	}
}

// AgentEventEnd 代理结束事件
type AgentEventEnd struct {
	Type     AgentEventType `json:"type"`
	Messages []ai.Message   `json:"messages"`
}

func (e *AgentEventEnd) GetType() AgentEventType {
	return e.Type
}

func NewAgentEventEnd(messages []ai.Message) *AgentEventEnd {
	return &AgentEventEnd{
		Type:     AgentEventTypeEnd,
		Messages: messages,
	}
}

// AgentEventTurnStart 轮次开始事件
type AgentEventTurnStart struct {
	Type AgentEventType `json:"type"`
}

func (e *AgentEventTurnStart) GetType() AgentEventType {
	return e.Type
}

func NewAgentEventTurnStart() *AgentEventTurnStart {
	return &AgentEventTurnStart{
		Type: AgentEventTypeTurnStart,
	}
}

// AgentEventTurnEnd 轮次结束事件
type AgentEventTurnEnd struct {
	Type        AgentEventType         `json:"type"`
	Message     ai.Message             `json:"message"`
	ToolResults []ai.ToolResultMessage `json:"toolResults"`
}

func (e *AgentEventTurnEnd) GetType() AgentEventType {
	return e.Type
}

func NewAgentEventTurnEnd(message ai.Message, toolResults []ai.ToolResultMessage) *AgentEventTurnEnd {
	return &AgentEventTurnEnd{
		Type:        AgentEventTypeTurnEnd,
		Message:     message,
		ToolResults: toolResults,
	}
}

// AgentEventMessageStart 消息开始事件
type AgentEventMessageStart struct {
	Type    AgentEventType `json:"type"`
	Message ai.Message     `json:"message"`
}

func (e *AgentEventMessageStart) GetType() AgentEventType {
	return e.Type
}

func NewAgentEventMessageStart(message ai.Message) *AgentEventMessageStart {
	return &AgentEventMessageStart{
		Type:    AgentEventTypeMessageStart,
		Message: message,
	}
}

// AgentEventMessageUpdate 消息更新事件
type AgentEventMessageUpdate struct {
	Type                  AgentEventType           `json:"type"`
	Message               ai.Message               `json:"message"`
	AssistantMessageEvent ai.AssistantMessageEvent `json:"assistantMessageEvent"`
}

func (e *AgentEventMessageUpdate) GetType() AgentEventType {
	return e.Type
}

func NewAgentEventMessageUpdate(message ai.Message, assistantMessageEvent ai.AssistantMessageEvent) *AgentEventMessageUpdate {
	return &AgentEventMessageUpdate{
		Type:                  AgentEventTypeMessageUpdate,
		Message:               message,
		AssistantMessageEvent: assistantMessageEvent,
	}
}

// AgentEventMessageEnd 消息结束事件
type AgentEventMessageEnd struct {
	Type    AgentEventType `json:"type"`
	Message ai.Message     `json:"message"`
}

func (e *AgentEventMessageEnd) GetType() AgentEventType {
	return e.Type
}

func NewAgentEventMessageEnd(message ai.Message) *AgentEventMessageEnd {
	return &AgentEventMessageEnd{
		Type:    AgentEventTypeMessageEnd,
		Message: message,
	}
}

// AgentEventToolExecutionStart 工具执行开始事件
type AgentEventToolExecutionStart struct {
	Type       AgentEventType `json:"type"`
	ToolCallID string         `json:"toolCallId"`
	ToolName   string         `json:"toolName"`
	Args       any            `json:"args"`
}

func (e *AgentEventToolExecutionStart) GetType() AgentEventType {
	return e.Type
}

func NewAgentEventToolExecutionStart(toolCallID, toolName string, args any) *AgentEventToolExecutionStart {
	return &AgentEventToolExecutionStart{
		Type:       AgentEventTypeToolExecutionStart,
		ToolCallID: toolCallID,
		ToolName:   toolName,
		Args:       args,
	}
}

// AgentEventToolExecutionUpdate 工具执行更新事件
type AgentEventToolExecutionUpdate struct {
	Type          AgentEventType `json:"type"`
	ToolCallID    string         `json:"toolCallId"`
	ToolName      string         `json:"toolName"`
	Args          any            `json:"args"`
	PartialResult any            `json:"partialResult"`
}

func (e *AgentEventToolExecutionUpdate) GetType() AgentEventType {
	return e.Type
}

func NewAgentEventToolExecutionUpdate(toolCallID, toolName string, args, partialResult any) *AgentEventToolExecutionUpdate {
	return &AgentEventToolExecutionUpdate{
		Type:          AgentEventTypeToolExecutionUpdate,
		ToolCallID:    toolCallID,
		ToolName:      toolName,
		Args:          args,
		PartialResult: partialResult,
	}
}

// AgentEventToolExecutionEnd 工具执行结束事件
type AgentEventToolExecutionEnd struct {
	Type       AgentEventType `json:"type"`
	ToolCallID string         `json:"toolCallId"`
	ToolName   string         `json:"toolName"`
	Result     any            `json:"result"`
	IsError    bool           `json:"isError"`
}

func (e *AgentEventToolExecutionEnd) GetType() AgentEventType {
	return e.Type
}

func NewAgentEventToolExecutionEnd(toolCallID, toolName string, result any, isError bool) *AgentEventToolExecutionEnd {
	return &AgentEventToolExecutionEnd{
		Type:       AgentEventTypeToolExecutionEnd,
		ToolCallID: toolCallID,
		ToolName:   toolName,
		Result:     result,
		IsError:    isError,
	}
}
