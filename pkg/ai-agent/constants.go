package agent

// AgentEventType 代理事件类型
type AgentEventType string

const (
	AgentEventTypeStart               AgentEventType = "agent_start"
	AgentEventTypeEnd                 AgentEventType = "agent_end"
	AgentEventTypeMessageStart        AgentEventType = "message_start"
	AgentEventTypeMessageUpdate       AgentEventType = "message_update"
	AgentEventTypeMessageEnd          AgentEventType = "message_end"
	AgentEventTypeTurnStart           AgentEventType = "turn_start"
	AgentEventTypeTurnEnd             AgentEventType = "turn_end"
	AgentEventTypeToolExecutionStart  AgentEventType = "tool_execution_start"
	AgentEventTypeToolExecutionUpdate AgentEventType = "tool_execution_update"
	AgentEventTypeToolExecutionEnd    AgentEventType = "tool_execution_end"
)
