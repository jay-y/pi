package ai

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	AuthCredentialTypeOAuth  string = "oauth"
	AuthCredentialTypeAPIKey string = "api_key"
)

// OAuthProviderInterface OAuth 提供商接口
type OAuthProviderInterface interface {
	GetID() string
	ModifyModels(models []Model, cred *OAuthCredentials) []Model
	GetAccessToken(cred *OAuthCredentials) string
	IsTokenExpired(cred *OAuthCredentials) bool
	RefreshToken(cred *OAuthCredentials) (*OAuthCredentials, error)
}

// OAuthCredentials OAuth 凭证
type OAuthCredentials struct {
	Type string `json:"type"` // "api_key" | "oauth"
	// API Key 凭证
	Key string `json:"key,omitempty"`
	// OAuth 凭证
	RefreshToken string `json:"refresh_token,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	ExpiresAt    int64  `json:"expires,omitempty"` // Unix 时间戳（毫秒）
}

// OAuthProviderId OAuth 提供商 ID
type OAuthProviderId string

// OAuthPrompt OAuth 登录提示
type OAuthPrompt struct {
	Message     string `json:"message"`
	Placeholder string `json:"placeholder,omitempty"`
	AllowEmpty  bool   `json:"allowEmpty,omitempty"`
}

// OAuthAuthInfo OAuth 认证信息
type OAuthAuthInfo struct {
	URL          string `json:"url"`
	Instructions string `json:"instructions,omitempty"`
}

// OAuthLoginCallbacks OAuth 登录回调
type OAuthLoginCallbacks struct {
	OnAuth            func(info OAuthAuthInfo) error
	OnPrompt          func(prompt OAuthPrompt) (string, error)
	OnProgress        func(message string) error
	OnManualCodeInput func() (string, error)
	Signal            chan struct{} // 用于取消操作
}

// OAuthProvider OAuth 提供商接口
type OAuthProvider interface {
	GetID() string
	GetName() string
	Login(callbacks OAuthLoginCallbacks) (*OAuthCredentials, error)
	UsesCallbackServer() bool
	RefreshToken(cred OAuthCredentials) (*OAuthCredentials, error)
	GetApiKey(cred OAuthCredentials) string
	ModifyModels(models []Model, cred OAuthCredentials) []Model
}

// OAuthProviderRegistry OAuth 提供商注册表
type OAuthProviderRegistry struct {
	providers map[string]OAuthProvider
	BuiltIns  []OAuthProvider
}

// NewOAuthProviderRegistry 创建 OAuth 提供商注册表
func NewOAuthProviderRegistry() *OAuthProviderRegistry {
	return &OAuthProviderRegistry{
		providers: make(map[string]OAuthProvider),
		BuiltIns:  []OAuthProvider{},
	}
}

// Register 注册 OAuth 提供商
func (r *OAuthProviderRegistry) Register(provider OAuthProvider) {
	r.providers[provider.GetID()] = provider
}

// Get 获取 OAuth 提供商
func (r *OAuthProviderRegistry) Get(id string) OAuthProvider {
	return r.providers[id]
}

// GetAll 获取所有 OAuth 提供商
func (r *OAuthProviderRegistry) GetAll() []OAuthProvider {
	providers := make([]OAuthProvider, 0, len(r.providers))
	for _, provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}

// Unregister 取消注册 OAuth 提供商
func (r *OAuthProviderRegistry) Unregister(id string) {
	// 检查是否是内置提供商
	for _, provider := range r.BuiltIns {
		if provider.GetID() == id {
			// 恢复内置实现
			r.providers[id] = provider
			return
		}
	}
	// 移除自定义提供商
	delete(r.providers, id)
}

// Reset 重置为内置提供商
func (r *OAuthProviderRegistry) Reset() {
	r.providers = make(map[string]OAuthProvider)
	for _, provider := range r.BuiltIns {
		r.providers[provider.GetID()] = provider
	}
}

// 全局 OAuth 提供商注册表
var oauthProviderRegistry = NewOAuthProviderRegistry()

// RegisterOAuthProvider 注册 OAuth 提供商
func RegisterOAuthProvider(provider OAuthProvider) {
	oauthProviderRegistry.Register(provider)
}

// GetOAuthProvider 获取 OAuth 提供商
func GetOAuthProvider(id string) OAuthProvider {
	return oauthProviderRegistry.Get(id)
}

// GetOAuthProviders 获取所有 OAuth 提供商
func GetOAuthProviders() []OAuthProvider {
	return oauthProviderRegistry.GetAll()
}

// UnregisterOAuthProvider 取消注册 OAuth 提供商
func UnregisterOAuthProvider(id string) {
	oauthProviderRegistry.Unregister(id)
}

// ResetOAuthProviders 重置为内置提供商
func ResetOAuthProviders() {
	oauthProviderRegistry.Reset()
}

// GetOAuthApiKey 获取 OAuth API Key（自动刷新）
func GetOAuthApiKey(providerID string, credentials map[string]OAuthCredentials) (*struct {
	ApiKey         string
	NewCredentials OAuthCredentials
}, error) {
	provider := GetOAuthProvider(providerID)
	if provider == nil {
		return nil, fmt.Errorf("unknown OAuth provider: %s", providerID)
	}

	// 获取该提供商的凭证
	cred, exists := credentials[providerID]
	if !exists {
		return nil, nil
	}

	// 检查令牌是否过期
	if time.Now().UnixMilli() >= cred.ExpiresAt {
		// 令牌已过期，刷新令牌
		newCred, err := provider.RefreshToken(cred)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh OAuth token for %s: %w", providerID, err)
		}
		cred = *newCred
	}

	apiKey := provider.GetApiKey(cred)
	return &struct {
		ApiKey         string
		NewCredentials OAuthCredentials
	}{
		ApiKey:         apiKey,
		NewCredentials: cred,
	}, nil
}

// RefreshOAuthToken 刷新 OAuth 令牌
func RefreshOAuthToken(providerID string, credentials OAuthCredentials) (*OAuthCredentials, error) {
	provider := GetOAuthProvider(providerID)
	if provider == nil {
		return nil, fmt.Errorf("unknown OAuth provider: %s", providerID)
	}

	return provider.RefreshToken(credentials)
}

//TODO 内置 OAuth 提供商实现

// GitHubCopilotOAuthProvider GitHub Copilot OAuth 提供商
type GitHubCopilotOAuthProvider struct{}

// GetID 获取提供商 ID
func (p *GitHubCopilotOAuthProvider) GetID() string {
	return "github-copilot"
}

// GetName 获取提供商名称
func (p *GitHubCopilotOAuthProvider) GetName() string {
	return "GitHub Copilot"
}

// Login 登录
func (p *GitHubCopilotOAuthProvider) Login(callbacks OAuthLoginCallbacks) (*OAuthCredentials, error) {
	// 这里需要实现 GitHub Copilot 的 OAuth 登录流程
	// 实际实现时需要打开浏览器进行授权
	return nil, fmt.Errorf("GitHub Copilot OAuth login not implemented")
}

// UsesCallbackServer 是否使用回调服务器
func (p *GitHubCopilotOAuthProvider) UsesCallbackServer() bool {
	return true
}

// RefreshToken 刷新令牌
func (p *GitHubCopilotOAuthProvider) RefreshToken(cred OAuthCredentials) (*OAuthCredentials, error) {
	// 这里需要实现 GitHub Copilot 的令牌刷新逻辑
	return nil, fmt.Errorf("GitHub Copilot OAuth token refresh not implemented")
}

// GetApiKey 获取 API Key
func (p *GitHubCopilotOAuthProvider) GetApiKey(cred OAuthCredentials) string {
	return cred.AccessToken
}

// ModifyModels 修改模型
func (p *GitHubCopilotOAuthProvider) ModifyModels(models []Model, cred OAuthCredentials) []Model {
	// 可以在这里修改模型配置
	return models
}

// AnthropicOAuthProvider Anthropic OAuth 提供商
type AnthropicOAuthProvider struct{}

// GetID 获取提供商 ID
func (p *AnthropicOAuthProvider) GetID() string {
	return "anthropic"
}

// GetName 获取提供商名称
func (p *AnthropicOAuthProvider) GetName() string {
	return "Anthropic"
}

// Login 登录
func (p *AnthropicOAuthProvider) Login(callbacks OAuthLoginCallbacks) (*OAuthCredentials, error) {
	// 这里需要实现 Anthropic 的 OAuth 登录流程
	return nil, fmt.Errorf("Anthropic OAuth login not implemented")
}

// UsesCallbackServer 是否使用回调服务器
func (p *AnthropicOAuthProvider) UsesCallbackServer() bool {
	return true
}

// RefreshToken 刷新令牌
func (p *AnthropicOAuthProvider) RefreshToken(cred OAuthCredentials) (*OAuthCredentials, error) {
	// 这里需要实现 Anthropic 的令牌刷新逻辑
	return nil, fmt.Errorf("Anthropic OAuth token refresh not implemented")
}

// GetApiKey 获取 API Key
func (p *AnthropicOAuthProvider) GetApiKey(cred OAuthCredentials) string {
	return cred.AccessToken
}

// ModifyModels 修改模型
func (p *AnthropicOAuthProvider) ModifyModels(models []Model, cred OAuthCredentials) []Model {
	// 可以在这里修改模型配置
	return models
}

// GoogleGeminiCliOAuthProvider Google Gemini CLI OAuth 提供商
type GoogleGeminiCliOAuthProvider struct{}

// GetID 获取提供商 ID
func (p *GoogleGeminiCliOAuthProvider) GetID() string {
	return "google-gemini-cli"
}

// GetName 获取提供商名称
func (p *GoogleGeminiCliOAuthProvider) GetName() string {
	return "Google Gemini CLI"
}

// Login 登录
func (p *GoogleGeminiCliOAuthProvider) Login(callbacks OAuthLoginCallbacks) (*OAuthCredentials, error) {
	// 这里需要实现 Google Gemini CLI 的 OAuth 登录流程
	return nil, fmt.Errorf("Google Gemini CLI OAuth login not implemented")
}

// UsesCallbackServer 是否使用回调服务器
func (p *GoogleGeminiCliOAuthProvider) UsesCallbackServer() bool {
	return true
}

// RefreshToken 刷新令牌
func (p *GoogleGeminiCliOAuthProvider) RefreshToken(cred OAuthCredentials) (*OAuthCredentials, error) {
	// 这里需要实现 Google Gemini CLI 的令牌刷新逻辑
	return nil, fmt.Errorf("Google Gemini CLI OAuth token refresh not implemented")
}

// GetApiKey 获取 API Key
func (p *GoogleGeminiCliOAuthProvider) GetApiKey(cred OAuthCredentials) string {
	return cred.AccessToken
}

// ModifyModels 修改模型
func (p *GoogleGeminiCliOAuthProvider) ModifyModels(models []Model, cred OAuthCredentials) []Model {
	// 可以在这里修改模型配置
	return models
}

// GoogleAntigravityOAuthProvider Google Antigravity OAuth 提供商
type GoogleAntigravityOAuthProvider struct{}

// GetID 获取提供商 ID
func (p *GoogleAntigravityOAuthProvider) GetID() string {
	return "google-antigravity"
}

// GetName 获取提供商名称
func (p *GoogleAntigravityOAuthProvider) GetName() string {
	return "Google Antigravity"
}

// Login 登录
func (p *GoogleAntigravityOAuthProvider) Login(callbacks OAuthLoginCallbacks) (*OAuthCredentials, error) {
	// 这里需要实现 Google Antigravity 的 OAuth 登录流程
	return nil, fmt.Errorf("Google Antigravity OAuth login not implemented")
}

// UsesCallbackServer 是否使用回调服务器
func (p *GoogleAntigravityOAuthProvider) UsesCallbackServer() bool {
	return true
}

// RefreshToken 刷新令牌
func (p *GoogleAntigravityOAuthProvider) RefreshToken(cred OAuthCredentials) (*OAuthCredentials, error) {
	// 这里需要实现 Google Antigravity 的令牌刷新逻辑
	return nil, fmt.Errorf("Google Antigravity OAuth token refresh not implemented")
}

// GetApiKey 获取 API Key
func (p *GoogleAntigravityOAuthProvider) GetApiKey(cred OAuthCredentials) string {
	return cred.AccessToken
}

// ModifyModels 修改模型
func (p *GoogleAntigravityOAuthProvider) ModifyModels(models []Model, cred OAuthCredentials) []Model {
	// 可以在这里修改模型配置
	return models
}

// OpenAICodexOAuthProvider OpenAI Codex OAuth 提供商
type OpenAICodexOAuthProvider struct{}

// GetID 获取提供商 ID
func (p *OpenAICodexOAuthProvider) GetID() string {
	return "openai-codex"
}

// GetName 获取提供商名称
func (p *OpenAICodexOAuthProvider) GetName() string {
	return "OpenAI Codex"
}

// Login 登录
func (p *OpenAICodexOAuthProvider) Login(callbacks OAuthLoginCallbacks) (*OAuthCredentials, error) {
	// 这里需要实现 OpenAI Codex 的 OAuth 登录流程
	return nil, fmt.Errorf("OpenAI Codex OAuth login not implemented")
}

// UsesCallbackServer 是否使用回调服务器
func (p *OpenAICodexOAuthProvider) UsesCallbackServer() bool {
	return true
}

// RefreshToken 刷新令牌
func (p *OpenAICodexOAuthProvider) RefreshToken(cred OAuthCredentials) (*OAuthCredentials, error) {
	// 这里需要实现 OpenAI Codex 的令牌刷新逻辑
	return nil, fmt.Errorf("OpenAI Codex OAuth token refresh not implemented")
}

// GetApiKey 获取 API Key
func (p *OpenAICodexOAuthProvider) GetApiKey(cred OAuthCredentials) string {
	return cred.AccessToken
}

// ModifyModels 修改模型
func (p *OpenAICodexOAuthProvider) ModifyModels(models []Model, cred OAuthCredentials) []Model {
	// 可以在这里修改模型配置
	return models
}

// RegisterBuiltInOAuthProviders 注册内置 OAuth 提供商
func RegisterBuiltInOAuthProviders() {
	builtIns := []OAuthProvider{
		&GitHubCopilotOAuthProvider{},
		&AnthropicOAuthProvider{},
		&GoogleGeminiCliOAuthProvider{},
		&GoogleAntigravityOAuthProvider{},
		&OpenAICodexOAuthProvider{},
	}

	for _, provider := range builtIns {
		RegisterOAuthProvider(provider)
	}

	// 保存内置提供商列表
	oauthProviderRegistry.BuiltIns = builtIns
}

// OAuthCredentialsFromJSON 从 JSON 字符串解析 OAuth 凭证
func OAuthCredentialsFromJSON(jsonStr string) (OAuthCredentials, error) {
	var cred OAuthCredentials
	if err := json.Unmarshal([]byte(jsonStr), &cred); err != nil {
		return OAuthCredentials{}, err
	}
	return cred, nil
}

// OAuthCredentialsToJSON 将 OAuth 凭证转换为 JSON 字符串
func OAuthCredentialsToJSON(cred OAuthCredentials) (string, error) {
	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsOAuthProvider 检查是否是 OAuth 提供商
func IsOAuthProvider(provider string) bool {
	return GetOAuthProvider(provider) != nil
}
