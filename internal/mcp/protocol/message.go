package protocol

// Message 表示与AI模型交互的基本消息单元
type Message struct {
	Role    string `json:"role"`           // 可以是 "system", "user", "assistant" 或 "function"
	Content string `json:"content"`        // 消息内容
	Name    string `json:"name,omitempty"` // 函数名称，仅当角色是 "function" 时使用
}

// ChatRequest 表示发送到AI模型的请求
type ChatRequest struct {
	Messages    []Message                `json:"messages"`              // 对话历史
	MaxTokens   int                      `json:"max_tokens,omitempty"`  // 最大生成令牌数
	Temperature float32                  `json:"temperature,omitempty"` // 温度参数，控制创造性
	Tools       []map[string]interface{} `json:"tools,omitempty"`       // 可用工具列表
	Stream      bool                     `json:"stream,omitempty"`      // 是否使用流式响应
	Model       string                   `json:"model,omitempty"`       // 模型名称
}

// ChatResponse 表示AI模型的响应
type ChatResponse struct {
	Content      string                   `json:"content"`                 // 响应内容
	ToolCalls    []map[string]interface{} `json:"tool_calls,omitempty"`    // 工具调用结果
	FinishReason string                   `json:"finish_reason,omitempty"` // 完成原因
	Usage        Usage                    `json:"usage,omitempty"`         // 令牌使用统计
}

// Usage 表示令牌使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`     // 提示使用的令牌数
	CompletionTokens int `json:"completion_tokens"` // 补全使用的令牌数
	TotalTokens      int `json:"total_tokens"`      // 总令牌数
}
