package client

import (
	"context"

	"distributedJob/internal/mcp/protocol"
)

// Config 表示MCP客户端配置
type Config struct {
	Provider    string      `json:"provider"`               // 提供者名称："openai", "anthropic", "local"
	Model       string      `json:"model"`                  // 模型名称
	APIKey      string      `json:"api_key,omitempty"`      // API密钥
	APIEndpoint string      `json:"api_endpoint,omitempty"` // API端点地址
	LocalConfig LocalConfig `json:"local_config,omitempty"` // 本地模型配置
}

// LocalConfig 表示本地模型服务配置
type LocalConfig struct {
	APIServerURL string             `json:"api_server_url"` // 本地模型服务器URL
	Models       []LocalModelConfig `json:"models"`         // 可用本地模型配置
}

// LocalModelConfig 表示单个本地模型配置
type LocalModelConfig struct {
	Name string `json:"name"` // 模型名称
	Path string `json:"path"` // 模型文件路径
}

// Client 定义MCP客户端接口
type Client interface {
	// Chat 向模型发送聊天请求并获取响应
	Chat(ctx context.Context, request *protocol.ChatRequest) (*protocol.ChatResponse, error)

	// StreamChat 向模型发送流式聊天请求
	StreamChat(ctx context.Context, request *protocol.ChatRequest) (<-chan protocol.ChatResponse, <-chan error, error)

	// GetProvider 获取当前使用的提供者名称
	GetProvider() string

	// GetModel 获取当前使用的模型名称
	GetModel() string
}

// NewClient 根据配置创建适当类型的MCP客户端
func NewClient(config Config) Client {
	switch config.Provider {
	case "openai":
		return NewOpenAIClient(config)
	case "anthropic":
		return NewAnthropicClient(config)
	case "local":
		return NewLocalClient(config)
	default:
		// 默认使用OpenAI客户端
		return NewOpenAIClient(config)
	}
}
