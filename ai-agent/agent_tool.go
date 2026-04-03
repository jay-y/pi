package agent

import (
	"context"

	"github.com/jay-y/pi/ai"
)

// AgentToolConfig 工具配置
type AgentToolConfig interface {
	GetName() string
	GetLabel() string
	GetDescription() string
	GetParameters() map[string]any
	Execute(ctx context.Context, params map[string]any, onUpdate func(partialResult *AgentToolResult)) (*AgentToolResult, error)
}

type AgentToolResultDetails interface {
	GetSummary() string
}

// AgentToolResult 工具执行结果
type AgentToolResult struct {
	Content []ai.ContentBlock `json:"content"`
	Details AgentToolResultDetails `json:"details"`
}

func NewAgentToolResult(content []ai.ContentBlock, details AgentToolResultDetails) *AgentToolResult {
	return &AgentToolResult{
		Content: content,
		Details: details,
	}
}

// AgentTool 代理工具
type AgentTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  map[string]any `json:"parameters"` // JSON Schema 字符串
	Label       string `json:"label"`
	Execute     func(ctx context.Context, params map[string]any, onUpdate func(partialResult *AgentToolResult)) (*AgentToolResult, error)
}

func NewAgentTool(tool AgentToolConfig) AgentTool {
	return AgentTool{
		Name:        tool.GetName(),
		Description: tool.GetDescription(),
		Parameters:  tool.GetParameters(),
		Label:       tool.GetLabel(),
		Execute:     tool.Execute,
	}
}