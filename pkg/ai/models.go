package ai

import (
	"strings"
)

// Cost 成本
type Cost struct {
	Input      float64 `json:"input"`      // 输入的token数的成本
	Output     float64 `json:"output"`     // 输出的token数的成本
	CacheRead  float64 `json:"cacheRead"`  // 缓存读取的token数的成本
	CacheWrite float64 `json:"cacheWrite"` // 缓存写入的token数的成本
	Total      float64 `json:"total"`      // 总的成本
}

// Usage 用量
type Usage struct {
	Input       int  `json:"input"`       // 输入的token数
	Output      int  `json:"output"`      // 输出的token数
	CacheRead   int  `json:"cacheRead"`   // 缓存读取的token数
	CacheWrite  int  `json:"cacheWrite"`  // 缓存写入的token数
	TotalTokens int  `json:"totalTokens"` // 总的token数
	Cost        Cost `json:"cost"`        // 成本
}

// ModelCost 模型成本
type ModelCost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
}

// Model 模型接口
type Model interface {
	GetID() string
	GetName() string
	GetAPI() string
	GetProvider() string
	GetBaseURL() string
	GetReasoning() bool
	GetInput() []string
	GetCost() ModelCost
	GetContextWindow() int
	GetMaxTokens() int
	GetHeaders() map[string]string
	GetCompat() any
	GetAPIKey() string
}

// BaseModel 默认模型实现
type BaseModel struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	API           string            `json:"api"`
	APIKey        string            `json:"apiKey,omitempty"`
	Provider      string            `json:"provider"`
	BaseURL       string            `json:"baseUrl"`
	Reasoning     bool              `json:"reasoning"`
	Input         []string          `json:"input"`
	Cost          ModelCost         `json:"cost"`
	ContextWindow int               `json:"contextWindow"`
	MaxTokens     int               `json:"maxTokens"`
	Headers       map[string]string `json:"headers,omitempty"`
	Compat        any               `json:"compat,omitempty"`
}

func (m *BaseModel) GetID() string                 { return m.ID }
func (m *BaseModel) GetName() string               { return m.Name }
func (m *BaseModel) GetAPI() string                { return m.API }
func (m *BaseModel) GetProvider() string           { return m.Provider }
func (m *BaseModel) GetBaseURL() string            { return m.BaseURL }
func (m *BaseModel) GetReasoning() bool            { return m.Reasoning }
func (m *BaseModel) GetInput() []string            { return m.Input }
func (m *BaseModel) GetCost() ModelCost            { return m.Cost }
func (m *BaseModel) GetContextWindow() int         { return m.ContextWindow }
func (m *BaseModel) GetMaxTokens() int             { return m.MaxTokens }
func (m *BaseModel) GetHeaders() map[string]string { return m.Headers }
func (m *BaseModel) GetCompat() any                { return m.Compat }
func (m *BaseModel) GetAPIKey() string             { return m.APIKey }

// 确保实现了 Model 接口
var _ Model = (*BaseModel)(nil)

// ModelRegistry 模型注册表
type ModelRegistry struct {
	providers map[string]map[string]Model
}

// NewModelRegistry 创建模型注册表
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		providers: make(map[string]map[string]Model),
	}
}

// Register 注册模型
func (r *ModelRegistry) Register(model Model) {
	provider := model.GetProvider()
	if _, exists := r.providers[provider]; !exists {
		r.providers[provider] = make(map[string]Model)
	}
	r.providers[provider][model.GetID()] = model
}

// Get 获取模型
func (r *ModelRegistry) Get(provider string, modelID string) Model {
	if providerModels, exists := r.providers[provider]; exists {
		return providerModels[modelID]
	}
	return nil
}

// GetProviders 获取所有提供商
func (r *ModelRegistry) GetProviders() []string {
	providers := make([]string, 0, len(r.providers))
	for provider := range r.providers {
		providers = append(providers, provider)
	}
	return providers
}

// GetModels 获取特定提供商的所有模型
func (r *ModelRegistry) GetModels(provider string) []Model {
	if providerModels, exists := r.providers[provider]; exists {
		models := make([]Model, 0, len(providerModels))
		for _, model := range providerModels {
			models = append(models, model)
		}
		return models
	}
	return []Model{}
}

// 全局模型注册表
var modelRegistry = NewModelRegistry()

// RegisterModel 注册模型到全局注册表
func RegisterModel(model Model) {
	modelRegistry.Register(model)
}

// GetModel 获取模型
func GetModel(provider string, modelID string) Model {
	return modelRegistry.Get(provider, modelID)
}

// GetProviders 获取所有提供商
func GetProviders() []string {
	return modelRegistry.GetProviders()
}

// GetModels 获取特定提供商的所有模型
func GetModels(provider string) []Model {
	return modelRegistry.GetModels(provider)
}

// CalculateCost 计算使用成本
func CalculateCost(model Model, usage *Usage) Cost {
	if usage == nil {
		return Cost{
			Input:      0,
			Output:     0,
			CacheRead:  0,
			CacheWrite: 0,
			Total:      0,
		}
	}
	usage.Cost.Input = (model.GetCost().Input / 1000000) * float64(usage.Input)
	usage.Cost.Output = (model.GetCost().Output / 1000000) * float64(usage.Output)
	usage.Cost.CacheRead = (model.GetCost().CacheRead / 1000000) * float64(usage.CacheRead)
	usage.Cost.CacheWrite = (model.GetCost().CacheWrite / 1000000) * float64(usage.CacheWrite)
	usage.Cost.Total = usage.Cost.Input + usage.Cost.Output + usage.Cost.CacheRead + usage.Cost.CacheWrite
	return usage.Cost
}

// SupportsXhigh 检查模型是否支持 xhigh 思考级别
// 支持的模型：
// - GPT-5.2 / GPT-5.3 / GPT-5.4 模型系列
// - Anthropic Messages API Opus 4.6 模型（xhigh 映射到自适应努力 "max"）
func SupportsXhigh(model Model) bool {
	modelID := model.GetID()
	api := model.GetAPI()

	if strings.Contains(modelID, "gpt-5.2") || strings.Contains(modelID, "gpt-5.3") || strings.Contains(modelID, "gpt-5.4") {
		return true
	}

	if api == "anthropic-messages" {
		return strings.Contains(modelID, "opus-4-6") || strings.Contains(modelID, "opus-4.6")
	}

	return false
}

// ModelsAreEqual 比较两个模型是否相等
// 通过比较 id 和 provider 来判断
// 如果任一模型为 nil，返回 false
func ModelsAreEqual(a Model, b Model) bool {
	if a == nil || b == nil {
		return false
	}
	return a.GetID() == b.GetID() && a.GetProvider() == b.GetProvider()
}
