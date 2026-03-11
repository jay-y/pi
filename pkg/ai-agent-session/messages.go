package session

// CustomMessage 自定义消息
type CustomMessage struct {
	Role       string      `json:"role"`
	CustomType string      `json:"customType,omitempty"`
	Content    interface{} `json:"content"`
	Display    bool        `json:"display,omitempty"`
	Details    interface{} `json:"details,omitempty"`
	Timestamp  int64       `json:"timestamp,omitempty"`
}

// GetRole 实现 Message 接口
func (c *CustomMessage) GetRole() string {
	return c.Role
}

func (c *CustomMessage) GetTimestamp() int64 {
	return c.Timestamp
}

func NewCustomMessage(role string, customType string, content interface{}, display bool, details interface{}) *CustomMessage {
	return &CustomMessage{
		Role:       role,
		CustomType: customType,
		Content:    content,
		Display:    display,
		Details:    details,
	}
}

// BashExecutionMessage Bash 执行消息
type BashExecutionMessage struct {
	Role               string `json:"role"`
	Type               string `json:"type"`
	Command            string `json:"command"`
	Output             string `json:"output"`
	ExitCode           *int   `json:"exitCode,omitempty"`
	Cancelled          bool   `json:"cancelled"`
	Truncated          bool   `json:"truncated"`
	FullOutputPath     string `json:"fullOutputPath,omitempty"`
	ExcludeFromContext bool   `json:"excludeFromContext,omitempty"`
	Timestamp          int64  `json:"timestamp"`
}

// GetRole 实现 Message 接口
func (m *BashExecutionMessage) GetRole() string {
	return m.Role
}

func (m *BashExecutionMessage) GetTimestamp() int64 {
	return m.Timestamp
}

// ToMap 实现 Message 接口
func (m *BashExecutionMessage) ToMap() map[string]any {
	return map[string]any{
		"role":               m.Role,
		"type":               m.Type,
		"command":            m.Command,
		"output":             m.Output,
		"exitCode":           m.ExitCode,
		"cancelled":          m.Cancelled,
		"truncated":          m.Truncated,
		"fullOutputPath":     m.FullOutputPath,
		"excludeFromContext": m.ExcludeFromContext,
		"timestamp":          m.Timestamp,
	}
}
