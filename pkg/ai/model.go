package ai

// Cost 成本
type Cost struct {
	Input      float64 `json:"input"` // 输入的token数的成本
	Output     float64 `json:"output"` // 输出的token数的成本
	CacheRead  float64 `json:"cacheRead"` // 缓存读取的token数的成本
	CacheWrite float64 `json:"cacheWrite"` // 缓存写入的token数的成本
	Total      float64 `json:"total"` // 总的成本
}

// Usage 用量
type Usage struct {
	Input       int  `json:"input"` // 输入的token数
	Output      int  `json:"output"` // 输出的token数
	CacheRead   int  `json:"cacheRead"` // 缓存读取的token数
	CacheWrite  int  `json:"cacheWrite"` // 缓存写入的token数
	TotalTokens int  `json:"totalTokens"` // 总的token数
	Cost        Cost `json:"cost"` // 成本
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
	GetAPI() ModelApi
	GetProvider() ModelProvider
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
	API           ModelApi          `json:"api"`
	APIKey        string            `json:"apiKey,omitempty"`
	Provider      ModelProvider     `json:"provider"`
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
func (m *BaseModel) GetAPI() ModelApi                   { return m.API }
func (m *BaseModel) GetProvider() ModelProvider         { return m.Provider }
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