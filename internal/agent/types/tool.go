package types

// Tool 表示Agent可用的工具接口
type Tool interface {
	// Name 返回工具名称
	Name() string

	// Description 返回工具描述
	Description() string

	// Execute 执行工具功能
	Execute(args map[string]interface{}) (interface{}, error)

	// Parameters 返回工具参数定义
	Parameters() map[string]Parameter
}

// Parameter 表示工具参数定义
type Parameter struct {
	Type        string      `json:"type"`              // 参数类型，如 "string", "number", "boolean", "array", "object"
	Description string      `json:"description"`       // 参数描述
	Required    bool        `json:"required"`          // 是否必需
	Default     interface{} `json:"default,omitempty"` // 默认值
}

// ToolResult 表示工具执行结果
type ToolResult struct {
	Success bool        `json:"success"`           // 执行是否成功
	Result  interface{} `json:"result,omitempty"`  // 执行结果
	Error   string      `json:"error,omitempty"`   // 错误信息
	ToolID  string      `json:"tool_id,omitempty"` // 工具ID
}

// Memory 表示Agent记忆接口
type Memory interface {
	// Add 添加记忆项
	Add(key string, value interface{}) error

	// Get 获取记忆项
	Get(key string) (interface{}, bool)

	// GetRecent 获取最近的n条记忆
	GetRecent(n int) []MemoryItem

	// Clear 清除所有记忆
	Clear()
}

// MemoryItem 表示记忆项
type MemoryItem struct {
	Key      string                 `json:"key"`
	Value    interface{}            `json:"value"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Message 表示Agent的消息
type Message struct {
	Role     string                 `json:"role"`               // 可以是 "user", "agent", "system"
	Content  string                 `json:"content"`            // 消息内容
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 元数据
}

// Plan 表示Agent的执行计划
type Plan struct {
	Steps       []PlanStep `json:"steps"`       // 计划步骤
	Goal        string     `json:"goal"`        // 计划目标
	Description string     `json:"description"` // 计划描述
}

// PlanStep 表示执行计划的一个步骤
type PlanStep struct {
	ID          string                 `json:"id"`               // 步骤ID
	Description string                 `json:"description"`      // 步骤描述
	Tool        string                 `json:"tool"`             // 使用的工具名称
	Parameters  map[string]interface{} `json:"parameters"`       // 工具参数
	Completed   bool                   `json:"completed"`        // 是否已完成
	Result      *ToolResult            `json:"result,omitempty"` // 执行结果
}
