package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jay-y/pi/ai"
	agent "github.com/jay-y/pi/ai-agent"
	"github.com/jay-y/pi/utils"
)

// GrepToolInput Grep 工具输入
type GrepToolInput struct {
	Pattern    string `json:"pattern"`
	Path       string `json:"path,omitempty"`
	Glob       string `json:"glob,omitempty"`
	IgnoreCase bool   `json:"ignoreCase,omitempty"`
	Literal    bool   `json:"literal,omitempty"`
	Context    int    `json:"context,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// GrepToolDetails Grep 工具详细信息
type GrepToolDetails struct {
	Summary           string            `json:"summary,omitempty"`
	Truncation        *TruncationResult `json:"truncation,omitempty"`
	MatchLimitReached int               `json:"matchLimitReached,omitempty"`
	LinesTruncated    bool              `json:"linesTruncated,omitempty"`
}

func NewGrepToolDetails(path string, resultCount int, truncation *TruncationResult, matchLimitReached int, linesTruncated bool) *GrepToolDetails {
	return &GrepToolDetails{
		Summary:           fmt.Sprintf("in %s (%d results)", path, resultCount),
		Truncation:        truncation,
		MatchLimitReached: matchLimitReached,
		LinesTruncated:    linesTruncated,
	}
}

// GrepOperations Grep 操作接口
type GrepOperations interface {
	IsDirectory(path string) (bool, error)
	ReadFile(path string) (string, error)
}

// DefaultGrepOperations 默认 Grep 操作
type DefaultGrepOperations struct{}

func (o *DefaultGrepOperations) IsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (o *DefaultGrepOperations) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GrepTool Grep 工具
type GrepTool struct {
	cwd        string
	operations GrepOperations
}

func NewGrepTool(cwd string, options ...GrepToolOption) agent.AgentTool {
	tool := &GrepTool{
		cwd:        cwd,
		operations: &DefaultGrepOperations{},
	}
	for _, opt := range options {
		if opt != nil {
			opt(tool)
		}
	}
	return agent.NewAgentTool(tool)
}

type GrepToolOption func(*GrepTool)

func WithGrepOperations(ops GrepOperations) GrepToolOption {
	return func(t *GrepTool) {
		t.operations = ops
	}
}

func (t *GrepTool) GetName() string  { return "grep" }
func (t *GrepTool) GetLabel() string { return "grep" }
func (t *GrepTool) GetDescription() string {
	return fmt.Sprintf("Search file contents for a pattern. Output truncated to %d matches or %dKB.",
		100, DEFAULT_MAX_BYTES/1024)
}
func (t *GrepTool) GetParameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Search pattern (regex or literal string)",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Directory or file to search",
			},
			"glob": map[string]any{
				"type":        "string",
				"description": "Filter files by glob pattern",
			},
			"ignoreCase": map[string]any{
				"type":        "boolean",
				"description": "Case-insensitive search",
			},
			"literal": map[string]any{
				"type":        "boolean",
				"description": "Treat pattern as literal string",
			},
			"context": map[string]any{
				"type":        "number",
				"description": "Number of context lines",
			},
			"limit": map[string]any{
				"type":        "number",
				"description": "Maximum number of matches",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepTool) Execute(ctx context.Context, params map[string]any, onUpdate func(*agent.AgentToolResult)) (*agent.AgentToolResult, error) {
	input, err := ValidateToolParams[GrepToolInput](params)
	if err != nil {
		return nil, err
	}
	pattern := input.Pattern
	searchDir := input.Path
	glob := input.Glob
	ignoreCase := input.IgnoreCase
	literal := input.Literal
	context := input.Context
	limit := input.Limit
	if limit == 0 {
		limit = 100
	}

	searchPath, err := resolvePath(searchDir, t.cwd)
	if err != nil {
		return nil, err
	}

	isDirectory, err := t.operations.IsDirectory(searchPath)
	if err != nil {
		return nil, fmt.Errorf("path not found: %s", searchPath)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	rgPath, err := exec.LookPath("rg")
	if err != nil {
		return nil, fmt.Errorf("ripgrep (rg) is not available: %w", err)
	}

	args := []string{"--json", "--line-number", "--color=never", "--hidden"}

	if ignoreCase {
		args = append(args, "--ignore-case")
	}

	if literal {
		args = append(args, "--fixed-strings")
	}

	if glob != "" {
		args = append(args, "--glob", glob)
	}

	args = append(args, pattern, searchPath)

	cmd := exec.CommandContext(ctx, rgPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	type MatchData struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		LineNumber int `json:"line_number"`
	}

	type MatchEvent struct {
		Type string    `json:"type"`
		Data MatchData `json:"data"`
	}

	matches := []struct {
		FilePath   string
		LineNumber int
	}{}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event MatchEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		if event.Type == "match" {
			matches = append(matches, struct {
				FilePath   string
				LineNumber int
			}{
				FilePath:   event.Data.Path.Text,
				LineNumber: event.Data.LineNumber,
			})

			if len(matches) >= limit {
				cmd.Process.Kill()
				break
			}
		}
	}

	cmd.Wait()

	if len(matches) == 0 {
		return &agent.AgentToolResult{
			Content: []ai.ContentBlock{ai.NewTextContentBlock("No matches found")},
		}, nil
	}

	fileCache := make(map[string][]string)
	getFileLines := func(filePath string) []string {
		if lines, ok := fileCache[filePath]; ok {
			return lines
		}

		content, err := t.operations.ReadFile(filePath)
		if err != nil {
			return []string{}
		}

		content = strings.ReplaceAll(content, "\r\n", "\n")
		content = strings.ReplaceAll(content, "\r", "\n")
		lines := strings.Split(content, "\n")
		fileCache[filePath] = lines
		return lines
	}

	formatPath := func(filePath string) string {
		if isDirectory {
			rel, err := filepath.Rel(searchPath, filePath)
			if err == nil && !strings.HasPrefix(rel, "..") {
				return filepath.ToSlash(rel)
			}
		}
		return filepath.Base(filePath)
	}

	outputLines := []string{}
	linesTruncated := false

	for _, match := range matches {
		relativePath := formatPath(match.FilePath)
		lines := getFileLines(match.FilePath)

		if len(lines) == 0 {
			outputLines = append(outputLines, fmt.Sprintf("%s:%d: (unable to read file)", relativePath, match.LineNumber))
			continue
		}

		start := match.LineNumber
		if context > 0 {
			start = max(1, match.LineNumber-context)
		}
		end := match.LineNumber
		if context > 0 {
			end = min(len(lines), match.LineNumber+context)
		}

		for current := start; current <= end; current++ {
			lineText := lines[current-1]
			lineText = strings.ReplaceAll(lineText, "\r", "")
			isMatchLine := current == match.LineNumber

			truncatedText, wasTruncated := TruncateLine(lineText)
			if wasTruncated {
				linesTruncated = true
			}

			if isMatchLine {
				outputLines = append(outputLines, fmt.Sprintf("%s:%d: %s", relativePath, current, truncatedText))
			} else {
				outputLines = append(outputLines, fmt.Sprintf("%s-%d- %s", relativePath, current, truncatedText))
			}
		}
	}

	rawOutput := strings.Join(outputLines, "\n")
	truncation := TruncateHead(rawOutput)

	output := truncation.Content
	notices := []string{}

	if len(matches) >= limit {
		notices = append(notices, fmt.Sprintf("%d matches limit reached. Use limit=%d for more, or refine pattern", limit, limit*2))
	}

	if truncation.Truncated {
		notices = append(notices, fmt.Sprintf("%s limit reached", utils.FormatSize(DEFAULT_MAX_BYTES)))
	}

	if linesTruncated {
		notices = append(notices, fmt.Sprintf("Some lines truncated to %d chars. Use read tool to see full lines", GREP_MAX_LINE_LENGTH))
	}

	if len(notices) > 0 {
		output += "\n\n[" + strings.Join(notices, ". ") + "]"
	}

	return &agent.AgentToolResult{
		Content: []ai.ContentBlock{ai.NewTextContentBlock(output)},
		Details: NewGrepToolDetails(searchDir, len(matches), &truncation, limit, linesTruncated),
	}, nil
}
