package session

import agent "github.com/jay-y/pi/pkg/ai-agent"

// AgentEventType 代理事件类型
const (
	AgentEventTypeAutoRetryStart agent.AgentEventType = "auto_retry_start"
	AgentEventTypeAutoRetryEnd   agent.AgentEventType = "auto_retry_end"
	AgentEventTypeSessionSwitch  agent.AgentEventType = "session_switch"
	// AgentEventTypeModelSelect         agent.AgentEventType = "model_select"
	// AgentEventTypeSessionStart        agent.AgentEventType = "session_start"
	// AgentEventTypeSessionEnd          agent.AgentEventType = "session_end"
	AgentEventTypeAutoCompactionStart agent.AgentEventType = "auto_compaction_start"
	AgentEventTypeAutoCompactionEnd   agent.AgentEventType = "auto_compaction_end"
)

// SessionSwitchReason 会话切换原因
type SessionSwitchReason string

const (
	SessionSwitchReasonResume SessionSwitchReason = "resume"
	SessionSwitchReasonSwitch SessionSwitchReason = "switch"
	SessionSwitchReasonFork   SessionSwitchReason = "fork"
	SessionSwitchReasonNew    SessionSwitchReason = "new"
)
