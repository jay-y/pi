package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jay-y/pi/ai"
	agent "github.com/jay-y/pi/ai-agent"
)

// FindToolInput Find 工具输入
type FindToolInput struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// FindToolDetails Find 工具详细信息
type FindToolDetails struct {
	Summary           string           `json:"summary,omitempty"`
	Truncation       *TruncationResult `json:"truncation,omitempty"`
	// ResultLimitReached int             `json:"resultLimitReached,omitempty"`
}

func NewFindToolDetails(path string, resultCount int, truncation *TruncationResult) *FindToolDetails {
	return &FindToolDetails{
		Summary:          fmt.Sprintf("in %s (%d results)", path, resultCount),
		Truncation:       truncation,
		// ResultLimitReached: resultLimitReached,
	}
}

// FindOperations find 工具的操作接口
type FindOperations interface {
	Exists(path string) bool
	Exec(path, name, typeFilter string) (string, error)
}

// DefaultFindOperations 默认的 find 操作
type DefaultFindOperations struct{}

func (o *DefaultFindOperations) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (o *DefaultFindOperations) Exec(path, name, typeFilter string) (string, error) {
	args := []string{path, "-type", typeFilter, "-name", name}
	cmd := exec.Command("find", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// FindTool Find 工具
type FindTool struct {
	cwd        string
	operations FindOperations
}

func NewFindTool(cwd string, options ...FindToolOption) agent.AgentTool {
	tool := &FindTool{
		cwd:        cwd,
		operations: &DefaultFindOperations{},
	}
	for _, opt := range options {
		if opt != nil { opt(tool) }
	}
	return agent.NewAgentTool(tool)
}

type FindToolOption func(*FindTool)

func WithFindOperations(ops FindOperations) FindToolOption {
	return func(t *FindTool) {
		t.operations = ops
	}
}

func (t *FindTool) GetName() string { return "find" }
func (t *FindTool) GetLabel() string { return "find" }
func (t *FindTool) GetDescription() string {
	return fmt.Sprintf("Find files by name and type. Output is truncated to %d lines or %dKB.",
		1000, DEFAULT_MAX_BYTES/1024)
}
func (t *FindTool) GetParameters() map[string]any {
	return map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to search (default: current directory)",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "File name pattern (e.g., *.go)",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "File type (f for file, d for directory)",
				},
			},
			"required": []string{"name"},
		}
}

func (t *FindTool) Execute(ctx context.Context, params map[string]any, onUpdate func(*agent.AgentToolResult)) (*agent.AgentToolResult, error) {
	input, err := ValidateToolParams[FindToolInput](params)
	if err != nil {
		return nil, err
	}
	path := input.Path
	if input.Path != "" {
		if !strings.HasPrefix(input.Path, "/") {
			path = strings.Join([]string{t.cwd, input.Path}, "/")
		} else {
			path = input.Path
		}
	}

	typeFilter := "f"
	if input.Type != "" {
		typeFilter = input.Type
	}

	// 执行 find
	output, err := t.operations.Exec(path, input.Name, typeFilter)
	resultCount := 0
	if err == nil {
		resultCount = len(strings.Split(output, "\n"))
	}
	
	// 处理输出
	truncation := TruncateHead(output)

	// 添加命令信息
	resultText := fmt.Sprintf("$ find %s -type %s -name '%s'\n\n%s", path, typeFilter, input.Name, truncation.Content)

	// 添加截断信息
	if truncation.Truncated {
		resultText += "\n\n[Output truncated. Use find options to limit output.]"
	}
	// 处理错误
	if err != nil && output == "" {
		resultText += fmt.Sprintf("\n\n[Error: %v]", err)
	}
	return &agent.AgentToolResult{
		Content: []ai.ContentBlock{
			ai.NewTextContentBlock(resultText),
		},
		Details: NewFindToolDetails(path, resultCount, &truncation),
	}, nil
}