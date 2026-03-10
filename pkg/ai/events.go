package ai

// AssistantMessageEvent 助手消息事件接口
type AssistantMessageEvent interface {
}

// AssistantMessageEventStart 开始事件
type AssistantMessageEventStart struct {
	Type    AssistantMessageEventType `json:"type"`
	Partial *AssistantMessage         `json:"partial"`
}

func NewAssistantMessageEventStart(partial *AssistantMessage) *AssistantMessageEventStart {
	return &AssistantMessageEventStart{
		Type:    AssistantMessageEventTypeStart,
		Partial: partial,
	}
}

// AssistantMessageEventTextStart 文本开始事件
type AssistantMessageEventTextStart struct {
	Type         AssistantMessageEventType `json:"type"`
	ContentIndex int                       `json:"contentIndex"`
	Partial      *AssistantMessage         `json:"partial"`
}

func NewAssistantMessageEventTextStart(contentIndex int, partial *AssistantMessage) *AssistantMessageEventTextStart {
	return &AssistantMessageEventTextStart{
		Type:         AssistantMessageEventTypeTextStart,
		ContentIndex: contentIndex,
		Partial:      partial,
	}
}

// AssistantMessageEventTextDelta 文本增量事件
type AssistantMessageEventTextDelta struct {
	Type         AssistantMessageEventType `json:"type"`
	ContentIndex int                       `json:"contentIndex"`
	Delta        string                    `json:"delta"`
	Partial      *AssistantMessage         `json:"partial"`
}

func NewAssistantMessageEventTextDelta(contentIndex int, delta string, partial *AssistantMessage) *AssistantMessageEventTextDelta {
	return &AssistantMessageEventTextDelta{
		Type:         AssistantMessageEventTypeTextDelta,
		ContentIndex: contentIndex,
		Delta:        delta,
		Partial:      partial,
	}
}

// AssistantMessageEventTextEnd 文本结束事件
type AssistantMessageEventTextEnd struct {
	Type         AssistantMessageEventType `json:"type"`
	ContentIndex int                       `json:"contentIndex"`
	Content      string                    `json:"content"`
	Partial      *AssistantMessage         `json:"partial"`
}

func NewAssistantMessageEventTextEnd(contentIndex int, content string, partial *AssistantMessage) *AssistantMessageEventTextEnd {
	return &AssistantMessageEventTextEnd{
		Type:         AssistantMessageEventTypeTextEnd,
		ContentIndex: contentIndex,
		Content:      content,
		Partial:      partial,
	}
}

// AssistantMessageEventThinkingStart 思考开始事件
type AssistantMessageEventThinkingStart struct {
	Type         AssistantMessageEventType `json:"type"`
	ContentIndex int                       `json:"contentIndex"`
	Partial      *AssistantMessage         `json:"partial"`
}

func (e *AssistantMessageEventThinkingStart) GetType() AssistantMessageEventType {
	return e.Type
}

func NewAssistantMessageEventThinkingStart(contentIndex int, partial *AssistantMessage) *AssistantMessageEventThinkingStart {
	return &AssistantMessageEventThinkingStart{
		Type:         AssistantMessageEventTypeThinkingStart,
		ContentIndex: contentIndex,
		Partial:      partial,
	}
}

// AssistantMessageEventThinkingDelta 思考增量事件
type AssistantMessageEventThinkingDelta struct {
	Type         AssistantMessageEventType `json:"type"`
	ContentIndex int                       `json:"contentIndex"`
	Delta        string                    `json:"delta"`
	Partial      *AssistantMessage         `json:"partial"`
}

func NewAssistantMessageEventThinkingDelta(contentIndex int, delta string, partial *AssistantMessage) *AssistantMessageEventThinkingDelta {
	return &AssistantMessageEventThinkingDelta{
		Type:         AssistantMessageEventTypeThinkingDelta,
		ContentIndex: contentIndex,
		Delta:        delta,
		Partial:      partial,
	}
}

// AssistantMessageEventThinkingEnd 思考结束事件
type AssistantMessageEventThinkingEnd struct {
	Type         AssistantMessageEventType `json:"type"`
	ContentIndex int                       `json:"contentIndex"`
	Content      string                    `json:"content"`
	Partial      *AssistantMessage         `json:"partial"`
}

func NewAssistantMessageEventThinkingEnd(contentIndex int, content string, partial *AssistantMessage) *AssistantMessageEventThinkingEnd {
	return &AssistantMessageEventThinkingEnd{
		Type:         AssistantMessageEventTypeThinkingEnd,
		ContentIndex: contentIndex,
		Content:      content,
		Partial:      partial,
	}
}

// AssistantMessageEventToolCallStart 工具调用开始事件
type AssistantMessageEventToolCallStart struct {
	Type         AssistantMessageEventType `json:"type"`
	ContentIndex int                       `json:"contentIndex"`
	Partial      *AssistantMessage         `json:"partial"`
}

func NewAssistantMessageEventToolCallStart(contentIndex int, partial *AssistantMessage) *AssistantMessageEventToolCallStart {
	return &AssistantMessageEventToolCallStart{
		Type:         AssistantMessageEventTypeToolCallStart,
		ContentIndex: contentIndex,
		Partial:      partial,
	}
}

// AssistantMessageEventToolCallDelta 工具调用增量事件
type AssistantMessageEventToolCallDelta struct {
	Type         AssistantMessageEventType `json:"type"`
	ContentIndex int                       `json:"contentIndex"`
	Delta        string                    `json:"delta"`
	Partial      *AssistantMessage         `json:"partial"`
}

func NewAssistantMessageEventToolCallDelta(contentIndex int, delta string, partial *AssistantMessage) *AssistantMessageEventToolCallDelta {
	return &AssistantMessageEventToolCallDelta{
		Type:         AssistantMessageEventTypeToolCallDelta,
		ContentIndex: contentIndex,
		Delta:        delta,
		Partial:      partial,
	}
}

// AssistantMessageEventToolCallEnd 工具调用结束事件
type AssistantMessageEventToolCallEnd struct {
	Type         AssistantMessageEventType `json:"type"`
	ContentIndex int                       `json:"contentIndex"`
	ToolCall     *ToolCallContentBlock     `json:"toolCall"`
	Partial      *AssistantMessage         `json:"partial"`
}

func NewAssistantMessageEventToolCallEnd(contentIndex int, toolCall *ToolCallContentBlock, partial *AssistantMessage) *AssistantMessageEventToolCallEnd {
	return &AssistantMessageEventToolCallEnd{
		Type:         AssistantMessageEventTypeToolCallEnd,
		ContentIndex: contentIndex,
		ToolCall:     toolCall,
		Partial:      partial,
	}
}

// AssistantMessageEventDone 完成事件
type AssistantMessageEventDone struct {
	Type    AssistantMessageEventType `json:"type"`
	Reason  StopReason                `json:"reason"`
	Message *AssistantMessage         `json:"message"`
}

func NewAssistantMessageEventDone(reason StopReason, message *AssistantMessage) *AssistantMessageEventDone {
	return &AssistantMessageEventDone{
		Type:    AssistantMessageEventTypeDone,
		Reason:  reason,
		Message: message,
	}
}

// AssistantMessageEventError 错误事件
type AssistantMessageEventError struct {
	Type   AssistantMessageEventType `json:"type"`
	Reason StopReason                `json:"reason"`
	Error  *AssistantMessage         `json:"error"`
}

func NewAssistantMessageEventError(reason StopReason, error *AssistantMessage) *AssistantMessageEventError {
	return &AssistantMessageEventError{
		Type:   AssistantMessageEventTypeError,
		Reason: reason,
		Error:  error,
	}
}
