package session

// AgentEventType 代理事件类型
const (
	AgentEventTypeAutoRetryStart string = "auto_retry_start"
	AgentEventTypeAutoRetryEnd   string = "auto_retry_end"
	AgentEventTypeSessionSwitch  string = "session_switch"
	// AgentEventTypeModelSelect         string = "model_select"
	// AgentEventTypeSessionStart        string = "session_start"
	// AgentEventTypeSessionEnd          string = "session_end"
	AgentEventTypeAutoCompactionStart string = "auto_compaction_start"
	AgentEventTypeAutoCompactionEnd   string = "auto_compaction_end"
)

// SessionSwitchReason 会话切换原因
type SessionSwitchReason string

const (
	SessionSwitchReasonResume SessionSwitchReason = "resume"
	SessionSwitchReasonSwitch SessionSwitchReason = "switch"
	SessionSwitchReasonFork   SessionSwitchReason = "fork"
	SessionSwitchReasonNew    SessionSwitchReason = "new"
)
