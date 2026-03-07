package session

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GetAgentDir 获取代理目录
func GetAgentDir() string {
	// 从环境变量或默认位置获取
	if dir := os.Getenv("PI_AGENT_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pi")
}

// ResolvePath 解析路径，支持 ~ 展开（等价于 $HOME）
// 示例：
//   ~                -> /Users/creator
//   ~/pi             -> /Users/creator/pi
//   ./workspace      -> /current/working/directory/workspace
//   /absolute/path   -> /absolute/path
func ResolvePath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// 展开 ~ (等价于 home dir)
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[2:]), nil
	}

	// 确保绝对路径
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		return abs, nil
	}

	return path, nil
}

// EnsureDirExists 确保目录存在，若不存在则创建
// 返回实际目录路径（已做 abs 处理）和错误
func EnsureDirExists(dir string) (string, error) {
	resolved, err := ResolvePath(dir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(resolved, 0755); err != nil {
		return "", err
	}
	return resolved, nil
}

// EnsureFileExists 确保文件存在，若不存在则创建空文件
// 返回实际文件路径（已做 abs 处理）和错误
func EnsureFileExists(filePath string) (string, error) {
	resolved, err := ResolvePath(filePath)
	if err != nil {
		return "", err
	}
	// 确保父目录存在
	parentDir := filepath.Dir(resolved)
	if parentDir != "" && parentDir != "." {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return "", err
		}
	}
	// 创建文件（若不存在）
	if _, err := os.Stat(resolved); os.IsNotExist(err) {
		if err := os.WriteFile(resolved, []byte{}, 0644); err != nil {
			return "", err
		}
	}
	return resolved, nil
}

// ResolveConfigValue 解析配置值（支持环境变量）
func ResolveConfigValue(value string) string {
	// 简单的环境变量解析
	if len(value) > 0 && value[0] == '$' {
		envVar := value[1:]
		if envValue := os.Getenv(envVar); envValue != "" {
			return envValue
		}
	}
	return value
}

// ResolveHeaders 解析头部配置
func ResolveHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	resolved := make(map[string]string)
	for k, v := range headers {
		resolved[k] = ResolveConfigValue(v)
	}
	return resolved
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

func removeAtIndex(slice []string, index int) []string {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

func getCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}

// containsIgnoreCase 检查字符串是否包含子串（忽略大小写）
func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return contains(sLower, substrLower)
}

// toLower 转换为小写
func toLower(s string) string {
	result := make([]byte, len(s))
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			result[i] = byte(c + 32)
		} else {
			result[i] = byte(c)
		}
	}
	return string(result)
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// randomInt 生成随机整数（简化实现）
func randomInt(min, max int) int {
	if min >= max {
		return min
	}
	// 简化实现，实际应使用 crypto/rand
	return min + (int(getCurrentTimestamp())%1000)*(max-min)/1000
}