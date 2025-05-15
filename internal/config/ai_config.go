package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"distributedJob/internal/agent/types"
	"distributedJob/internal/mcp/client"
	"distributedJob/internal/rag/embedding"
	"distributedJob/internal/rag/retriever"
)

// AIConfig 全局AI组件配置
type AIConfig struct {
	Enabled bool         `json:"enabled" yaml:"enabled"` // 是否启用AI功能
	MCP     MCPConfig    `json:"mcp" yaml:"mcp"`         // MCP相关配置
	RAG     RAGConfig    `json:"rag" yaml:"rag"`         // RAG相关配置
	Agent   AgentConfig  `json:"agent" yaml:"agent"`     // Agent相关配置
	Common  CommonConfig `json:"common" yaml:"common"`   // 通用配置
}

// MCPConfig MCP组件配置
type MCPConfig struct {
	Clients map[string]client.Config `json:"clients" yaml:"clients"` // MCP客户端配置
}

// RAGConfig RAG组件配置
type RAGConfig struct {
	Embeddings   map[string]embedding.Config       `json:"embeddings" yaml:"embeddings"`       // 嵌入模型配置
	VectorStores map[string]map[string]interface{} `json:"vector_stores" yaml:"vector_stores"` // 向量存储配置
	Retriever    map[string]retriever.Config       `json:"retriever" yaml:"retriever"`         // 检索器配置
	ChunkSize    int                               `json:"chunk_size" yaml:"chunk_size"`       // 默认块大小
	ChunkOverlap int                               `json:"chunk_overlap" yaml:"chunk_overlap"` // 默认块重叠大小
}

// AgentConfig 智能代理配置
type AgentConfig struct {
	Tools  map[string]bool      `json:"tools" yaml:"tools"`   // 启用的工具
	Memory types.MemoryConfig   `json:"memory" yaml:"memory"` // 记忆配置
	Agents map[string]AgentSpec `json:"agents"`               // 预定义的代理配置
}

// AgentSpec 代理规格配置
type AgentSpec struct {
	Name         string   `json:"name"`          // 代理名称
	Description  string   `json:"description"`   // 代理描述
	Model        string   `json:"model"`         // 使用的模型
	SystemPrompt string   `json:"system_prompt"` // 系统提示词
	Tools        []string `json:"tools"`         // 使用的工具
}

// CommonConfig 通用配置
type CommonConfig struct {
	LogLevel  string `json:"log_level" yaml:"log_level"`   // 日志级别
	Metrics   bool   `json:"metrics" yaml:"metrics"`       // 是否启用指标
	APISecret string `json:"api_secret" yaml:"api_secret"` // API密钥
	TempDir   string `json:"temp_dir" yaml:"temp_dir"`     // 临时目录
	MaxTokens int    `json:"max_tokens" yaml:"max_tokens"` // 最大令牌数
}

// LoadAIConfig 从配置文件加载AI配置
func LoadAIConfig(configPath string) (*AIConfig, error) {
	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	var config AIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// 设置默认值
	setAIDefaults(&config)

	return &config, nil
}

// setAIDefaults 为配置设置默认值
func setAIDefaults(config *AIConfig) {
	// MCP默认值
	if len(config.MCP.Clients) == 0 {
		config.MCP.Clients = map[string]client.Config{
			"default": {
				Provider: "openai",
				Model:    "gpt-3.5-turbo",
			},
		}
	}

	// RAG默认值
	if config.RAG.ChunkSize == 0 {
		config.RAG.ChunkSize = 1000
	}
	if config.RAG.ChunkOverlap == 0 {
		config.RAG.ChunkOverlap = config.RAG.ChunkSize / 10
	}

	// Agent默认值
	if config.Agent.Memory.Capacity == 0 {
		config.Agent.Memory.Capacity = 100
	}
	if config.Agent.Memory.Type == "" {
		config.Agent.Memory.Type = types.BufferMemory
	}

	// 通用默认值
	if config.Common.LogLevel == "" {
		config.Common.LogLevel = "info"
	}
}

// SaveAIConfig 将配置保存到文件
func SaveAIConfig(config *AIConfig, configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultAIConfig 返回默认配置
func GetDefaultAIConfig() *AIConfig {
	config := &AIConfig{
		MCP: MCPConfig{
			Clients: map[string]client.Config{
				"openai": {
					Provider:    "openai",
					Model:       "gpt-3.5-turbo",
					APIEndpoint: "https://api.openai.com/v1/chat/completions",
					APIKey:      os.Getenv("OPENAI_API_KEY"),
				},
				"local": {
					Provider:    "local",
					Model:       "llama2",
					APIEndpoint: "http://localhost:8000/v1/chat/completions",
				},
			},
		},
		RAG: RAGConfig{
			Embeddings: map[string]embedding.Config{
				"openai": {
					Type:        "openai",
					Model:       "text-embedding-3-small",
					APIEndpoint: "https://api.openai.com/v1/embeddings",
					APIKey:      os.Getenv("OPENAI_API_KEY"),
					Dimension:   1536,
				},
				"local": {
					Type:        "local",
					Model:       "bge-base-zh",
					APIEndpoint: "http://localhost:8000/v1/embeddings",
					Dimension:   768,
				},
			},
			VectorStores: map[string]map[string]interface{}{
				"memory": {
					"type": "memory",
				},
				"postgres": {
					"type":              "postgres",
					"connection_string": "postgresql://postgres:password@localhost:5432/vector_db?sslmode=disable",
					"table_name":        "vector_store",
				},
			},
			Retriever: map[string]retriever.Config{
				"default": {
					VectorStore:       "memory",
					VectorStoreConfig: map[string]interface{}{},
					EmbeddingConfig:   embedding.Config{Type: "openai"},
				},
			},
			ChunkSize:    1000,
			ChunkOverlap: 100,
		},
		Agent: AgentConfig{
			Tools: map[string]bool{
				"scheduler_tool": true,
				"data_tool":      true,
				"system_tool":    true,
			},
			Memory: types.MemoryConfig{
				Type:     types.BufferMemory,
				Capacity: 100,
			},
			Agents: map[string]AgentSpec{
				"assistant": {
					Name:         "General Assistant",
					Description:  "通用AI助手，可以回答问题并执行任务",
					Model:        "gpt-3.5-turbo",
					SystemPrompt: "你是一个可靠的AI助手，可以帮助用户解决各种问题。",
					Tools:        []string{"data_tool"},
				},
				"admin": {
					Name:         "System Administrator",
					Description:  "系统管理员助手，可以帮助管理系统和服务",
					Model:        "gpt-4",
					SystemPrompt: "你是一个系统管理员助手，可以帮助用户管理系统资源和服务。",
					Tools:        []string{"system_tool", "scheduler_tool"},
				},
			},
		},
		Common: CommonConfig{
			LogLevel:  "info",
			Metrics:   true,
			APISecret: os.Getenv("API_SECRET"),
		},
	}

	return config
}
