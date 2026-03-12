package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jay-y/pi/pkg/ai"
	"github.com/jay-y/pi/pkg/utils"
	"github.com/jay-y/pi/pkg/utils/lockfile"
)

// AuthStorageData 认证存储数据结构
type AuthStorageData map[string]*ai.OAuthCredentials

// LockResult 锁定操作结果
type LockResult struct {
	Result any
	Next   *string // 如果需要更新文件内容
}

// AuthStorageBackend 认证存储后端接口
type AuthStorageBackend interface {
	WithLock(fn func(current string) *LockResult) (*LockResult, error)
}

// FileAuthStorageBackend 基于文件的认证存储后端
type FileAuthStorageBackend struct {
	authPath string
}

// NewFileAuthStorageBackend 创建文件认证存储后端
func NewFileAuthStorageBackend(authPath string) *FileAuthStorageBackend {
	if authPath == "" {
		authPath = filepath.Join(utils.GetAgentDir(), "auth.json")
	}
	return &FileAuthStorageBackend{
		authPath: authPath,
	}
}

// ensureParentDir 确保父目录存在
func (b *FileAuthStorageBackend) ensureParentDir() error {
	dir := filepath.Dir(b.authPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0o700)
	}
	return nil
}

// ensureFileExists 确保文件存在
func (b *FileAuthStorageBackend) ensureFileExists() error {
	if _, err := os.Stat(b.authPath); os.IsNotExist(err) {
		if err := b.ensureParentDir(); err != nil {
			return err
		}
		return os.WriteFile(b.authPath, []byte("{}"), 0o600)
	}
	return nil
}

// WithLock 执行带锁的操作
func (b *FileAuthStorageBackend) WithLock(fn func(current string) *LockResult) (*LockResult, error) {
	if err := b.ensureParentDir(); err != nil {
		return nil, err
	}
	if err := b.ensureFileExists(); err != nil {
		return nil, err
	}

	// 创建锁文件路径
	lockPath := b.authPath + ".lock"

	// 创建锁文件
	lockfile, err := lockfile.NewLockfile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create lockfile: %w", err)
	}

	// 尝试获取锁（带重试）
	maxAttempts := 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := lockfile.TryLock()
		if err == nil {
			// 成功获取锁
			defer func() {
				_ = lockfile.Unlock()
				_ = os.Remove(lockPath)
			}()

			// 读取当前内容
			current, err := os.ReadFile(b.authPath)
			if err != nil && !os.IsNotExist(err) {
				return nil, err
			}

			currentStr := string(current)
			if len(current) == 0 {
				currentStr = "{}"
			}

			// 执行用户函数
			result := fn(currentStr)

			// 如果需要更新文件
			if result != nil && result.Next != nil {
				if err := b.ensureParentDir(); err != nil {
					return result, err
				}
				if err := os.WriteFile(b.authPath, []byte(*result.Next), 0o600); err != nil {
					return result, err
				}
			}

			return result, nil
		}

		// 如果是临时错误，重试
		if te, ok := err.(interface{ Temporary() bool }); ok && te.Temporary() {
			// 简单的退避策略
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
			continue
		}

		// 其他错误，直接返回
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return nil, fmt.Errorf("failed to acquire lock after %d attempts", maxAttempts)
}

// InMemoryAuthStorageBackend 内存认证存储后端
type InMemoryAuthStorageBackend struct {
	value string
	mu    sync.Mutex
}

// NewInMemoryAuthStorageBackend 创建内存认证存储后端
func NewInMemoryAuthStorageBackend() *InMemoryAuthStorageBackend {
	return &InMemoryAuthStorageBackend{
		value: "{}",
	}
}

// WithLock 执行带锁的操作（内存版本）
func (b *InMemoryAuthStorageBackend) WithLock(fn func(current string) *LockResult) (*LockResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := fn(b.value)

	if result != nil && result.Next != nil {
		b.value = *result.Next
	}

	return result, nil
}

// AuthStorage 认证存储
type AuthStorage struct {
	storage          AuthStorageBackend
	data             AuthStorageData
	runtimeOverrides map[string]string
	fallbackResolver func(provider string) *string
	loadError        error
	errors           []error
	mu               sync.RWMutex
}

// NewAuthStorage 创建认证存储
func NewAuthStorage(storage AuthStorageBackend) *AuthStorage {
	auth := &AuthStorage{
		storage:          storage,
		data:             make(AuthStorageData),
		runtimeOverrides: make(map[string]string),
		errors:           make([]error, 0),
	}
	auth.reload()
	return auth
}

// CreateAuthStorage 创建文件认证存储
func CreateAuthStorage(authPath string) *AuthStorage {
	return NewAuthStorage(NewFileAuthStorageBackend(authPath))
}

// CreateAuthStorageFromBackend 从后端创建认证存储
func CreateAuthStorageFromBackend(storage AuthStorageBackend) *AuthStorage {
	return NewAuthStorage(storage)
}

// CreateAuthStorageInMemory 创建内存认证存储
func CreateAuthStorageInMemory(data AuthStorageData) *AuthStorage {
	storage := NewInMemoryAuthStorageBackend()

	// 初始化数据
	if data != nil {
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		storage.value = string(jsonData)
	}

	return NewAuthStorage(storage)
}

// SetRuntimeApiKey 设置运行时 API Key 覆盖（不持久化）
func (a *AuthStorage) SetRuntimeApiKey(provider string, apiKey string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.runtimeOverrides[provider] = apiKey
}

// RemoveRuntimeApiKey 移除运行时 API Key 覆盖
func (a *AuthStorage) RemoveRuntimeApiKey(provider string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.runtimeOverrides, provider)
}

// SetFallbackResolver 设置回退解析器
func (a *AuthStorage) SetFallbackResolver(resolver func(provider string) *string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.fallbackResolver = resolver
}

// recordError 记录错误
func (a *AuthStorage) recordError(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errors = append(a.errors, err)
}

// parseStorageData 解析存储数据
func (a *AuthStorage) parseStorageData(content string) (AuthStorageData, error) {
	if content == "" || content == "{}" {
		return make(AuthStorageData), nil
	}

	var data AuthStorageData
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return nil, err
	}
	return data, nil
}

// reload 从存储重新加载凭证（内部方法）
func (a *AuthStorage) reload() {
	result, err := a.storage.WithLock(func(current string) *LockResult {
		return &LockResult{Result: current}
	})

	if err != nil {
		a.recordError(err)
		a.loadError = err
		return
	}

	content, ok := result.Result.(string)
	if !ok {
		content = "{}"
	}

	data, err := a.parseStorageData(content)
	if err != nil {
		a.recordError(err)
		a.loadError = err
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.data = data
	a.loadError = nil
}

// Reload 从存储重新加载凭证（公开方法）
func (a *AuthStorage) Reload() error {
	a.reload()
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.loadError
}

// persistProviderChange 持久化提供商凭证变更
func (a *AuthStorage) persistProviderChange(provider string, credential *ai.OAuthCredentials) error {
	a.mu.RLock()
	loadError := a.loadError
	a.mu.RUnlock()

	if loadError != nil {
		return nil // 如果加载失败，不持久化
	}

	_, err := a.storage.WithLock(func(current string) *LockResult {
		currentData, err := a.parseStorageData(current)
		if err != nil {
			return nil
		}

		merged := make(AuthStorageData)
		for k, v := range currentData {
			merged[k] = v
		}

		if credential != nil {
			merged[provider] = credential
		} else {
			delete(merged, provider)
		}

		jsonData, _ := json.MarshalIndent(merged, "", "  ")
		next := string(jsonData)
		return &LockResult{
			Result: "",
			Next:   &next,
		}
	})

	if err != nil {
		a.recordError(err)
		return err
	}

	return nil
}

// Get 获取提供商凭证
func (a *AuthStorage) Get(provider string) *ai.OAuthCredentials {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.data[provider]
}

// Set 设置提供商凭证
func (a *AuthStorage) Set(provider string, credential *ai.OAuthCredentials) {
	a.mu.Lock()
	a.data[provider] = credential
	a.mu.Unlock()

	a.persistProviderChange(provider, credential)
}

// Remove 移除提供商凭证
func (a *AuthStorage) Remove(provider string) {
	a.mu.Lock()
	delete(a.data, provider)
	a.mu.Unlock()

	a.persistProviderChange(provider, nil)
}

// List 列出所有有凭证的提供商
func (a *AuthStorage) List() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	providers := make([]string, 0, len(a.data))
	for provider := range a.data {
		providers = append(providers, provider)
	}
	return providers
}

// Has 检查提供商是否有凭证（仅在 auth.json 中）
func (a *AuthStorage) Has(provider string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	_, exists := a.data[provider]
	return exists
}

// HasAuth 检查是否有任何形式的认证（包括运行时覆盖、环境变量、回退解析器）
func (a *AuthStorage) HasAuth(provider string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 运行时覆盖
	if _, exists := a.runtimeOverrides[provider]; exists {
		return true
	}

	// auth.json 中的凭证
	if _, exists := a.data[provider]; exists {
		return true
	}

	// 环境变量
	if ai.GetEnvApiKey(provider) != "" {
		return true
	}

	// 回退解析器
	if a.fallbackResolver != nil {
		if key := a.fallbackResolver(provider); key != nil && *key != "" {
			return true
		}
	}

	return false
}

// GetAll 获取所有凭证
func (a *AuthStorage) GetAll() AuthStorageData {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(AuthStorageData)
	for k, v := range a.data {
		result[k] = v
	}
	return result
}

// DrainErrors 取出所有错误
func (a *AuthStorage) DrainErrors() []error {
	a.mu.Lock()
	defer a.mu.Unlock()

	errors := a.errors
	a.errors = make([]error, 0)
	return errors
}

// refreshOAuthTokenWithLock 带锁刷新 OAuth 令牌
func (a *AuthStorage) refreshOAuthTokenWithLock(providerID string, oauthProvider ai.OAuthProvider) (*ai.OAuthCredentials, error) {
	result, err := a.storage.WithLock(func(current string) *LockResult {
		currentData, err := a.parseStorageData(current)
		if err != nil {
			return nil
		}

		// 更新内存数据
		a.mu.Lock()
		a.data = currentData
		a.loadError = nil
		a.mu.Unlock()

		cred, exists := currentData[providerID]
		if !exists || cred.Type != ai.AuthCredentialTypeOAuth {
			return nil
		}

		// 检查令牌是否过期
		if time.Now().UnixMilli() < cred.ExpiresAt {
			return &LockResult{
				Result: cred,
			}
		}

		// 刷新令牌
		newCred, err := oauthProvider.RefreshToken(*cred)
		if err != nil {
			return nil
		}

		// 合并数据
		merged := make(AuthStorageData)
		for k, v := range currentData {
			merged[k] = v
		}
		merged[providerID] = newCred

		// 更新内存
		a.mu.Lock()
		a.data = merged
		a.loadError = nil
		a.mu.Unlock()

		jsonData, _ := json.MarshalIndent(merged, "", "  ")
		next := string(jsonData)
		return &LockResult{
			Result: newCred,
			Next:   &next,
		}
	})

	if err != nil {
		return nil, err
	}

	if result == nil || result.Result == nil {
		return nil, nil
	}

	cred, ok := result.Result.(*ai.OAuthCredentials)
	if !ok {
		return nil, nil
	}

	return cred, nil
}

// GetApiKey 获取提供商的 API Key
// 优先级：
// 1. 运行时覆盖（CLI --api-key）
// 2. auth.json 中的 API Key
// 3. auth.json 中的 OAuth 令牌（自动刷新）
// 4. 环境变量
// 5. 回退解析器（models.json 自定义提供商）
func (a *AuthStorage) GetApiKey(provider string) (string, error) {
	a.mu.RLock()
	runtimeKey, exists := a.runtimeOverrides[provider]
	a.mu.RUnlock()

	// 1. 运行时覆盖
	if exists {
		return runtimeKey, nil
	}

	a.mu.RLock()
	cred := a.data[provider]
	a.mu.RUnlock()

	// 2. API Key 凭证
	if cred != nil && cred.Type == ai.AuthCredentialTypeAPIKey {
		return utils.ResolveConfigValue(cred.Key), nil
	}

	// 3. OAuth 凭证
	if cred != nil && cred.Type == ai.AuthCredentialTypeOAuth {
		oauthProvider := ai.GetOAuthProvider(provider)
		if oauthProvider == nil {
			// Unknown OAuth provider, can't get API key
			return "", nil
		}

		// 检查令牌是否过期
		if time.Now().UnixMilli() >= cred.ExpiresAt {
			// 令牌已过期，使用带锁刷新防止竞争条件
			newCred, err := a.refreshOAuthTokenWithLock(provider, oauthProvider)
			if err != nil {
				a.recordError(err)
				// 刷新失败 - 重新读取文件检查其他实例是否成功
				a.reload()
				updatedCred := a.data[provider]

				if updatedCred != nil && updatedCred.Type == ai.AuthCredentialTypeOAuth && time.Now().UnixMilli() < updatedCred.ExpiresAt {
					// 其他实例刷新成功，使用那些凭证
					return oauthProvider.GetApiKey(*updatedCred), nil
				}

				// 刷新确实失败 - 返回空字符串让模型发现跳过此提供商
				// 用户可以 /login 重新认证（凭证保留用于重试）
				return "", nil
			}

			if newCred != nil {
				return oauthProvider.GetApiKey(*newCred), nil
			}
		} else {
			// 令牌未过期，使用当前访问令牌
			return oauthProvider.GetApiKey(*cred), nil
		}
	}

	// 4. 环境变量
	if envKey := ai.GetEnvApiKey(provider); envKey != "" {
		return envKey, nil
	}

	// 5. 回退解析器
	a.mu.RLock()
	resolver := a.fallbackResolver
	a.mu.RUnlock()

	if resolver != nil {
		if key := resolver(provider); key != nil {
			return *key, nil
		}
	}

	return "", nil
}

// Login OAuth 登录
func (a *AuthStorage) Login(providerID string, oauthProvider ai.OAuthProvider) error {
	// 这里需要实现 OAuth 登录流程
	// 实际使用时需要调用 oauthProvider 的登录方法
	return fmt.Errorf("OAuth login not implemented")
}

// Logout 登出提供商
func (a *AuthStorage) Logout(provider string) {
	a.Remove(provider)
}

// GetOAuthProviders 获取所有注册的 OAuth 提供商
func (a *AuthStorage) GetOAuthProviders() []ai.OAuthProvider {
	return ai.GetOAuthProviders()
}
