package session

import agent "github.com/jay-y/pi/pkg/ai-agent"

// AgentEventType 代理事件类型
const (
	AgentEventTypeAutoRetryEnd  agent.AgentEventType = "auto_retry_end"
	AgentEventTypeSessionSwitch agent.AgentEventType = "session_switch"
)

// SessionSwitchReason 会话切换原因
type SessionSwitchReason string

const (
	SessionSwitchReasonResume SessionSwitchReason = "resume"
	SessionSwitchReasonSwitch SessionSwitchReason = "switch"
	SessionSwitchReasonFork   SessionSwitchReason = "fork"
	SessionSwitchReasonNew    SessionSwitchReason = "new"
)
