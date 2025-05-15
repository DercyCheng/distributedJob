// Package ai provides the integration of all AI capabilities
package ai

import (
	"context"
	"fmt"
	"sync"

	"distributedJob/internal/agent/core"
	"distributedJob/internal/agent/tools"
	"distributedJob/internal/agent/types"
	"distributedJob/internal/config"
	"distributedJob/internal/mcp/client"
	"distributedJob/internal/mcp/protocol"
	"distributedJob/internal/rag/embedding"
	"distributedJob/internal/rag/generator"
	"distributedJob/internal/rag/retriever"
	"distributedJob/internal/rag/vectorstore"
)

// Controller 是AI功能的控制器，整合Agent、MCP和RAG功能
type Controller struct {
	config      *config.AIConfig
	agents      map[string]*core.Agent
	mcpClients  map[string]client.Client
	ragSystems  map[string]*RAGSystem
	initialized bool
	mu          sync.Mutex
}

// RAGSystem 表示RAG系统的组合
type RAGSystem struct {
	VectorStore vectorstore.VectorStore
	Retriever   retriever.Engine
	Generator   *generator.Generator
}

// ModelInfo 表示模型信息
type ModelInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Provider    string  `json:"provider"`
	Description string  `json:"description,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	LocalModel  bool    `json:"local_model"`
	Cost        float64 `json:"cost_per_token,omitempty"`
}

// NewController 创建一个新的AI控制器
func NewController(cfg *config.AIConfig) *Controller {
	return &Controller{
		config:     cfg,
		agents:     make(map[string]*core.Agent),
		mcpClients: make(map[string]client.Client),
		ragSystems: make(map[string]*RAGSystem),
	}
}

// Initialize 初始化AI控制器，加载所有组件
func (c *Controller) Initialize(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	// 初始化MCP客户端
	for name, clientConfig := range c.config.MCP.Clients {
		c.mcpClients[name] = client.NewClient(clientConfig)
	}

	// 默认MCP客户端
	if _, ok := c.mcpClients["default"]; !ok {
		// 确保有默认客户端
		c.mcpClients["default"] = client.NewClient(client.Config{
			Provider: "local",
			Model:    "deepseekv3-7b",
		})
	}

	// 初始化Agent
	if err := c.initializeAgents(); err != nil {
		return fmt.Errorf("failed to initialize agents: %w", err)
	}

	// 初始化RAG系统
	if err := c.initializeRAG(ctx); err != nil {
		return fmt.Errorf("failed to initialize RAG: %w", err)
	}

	c.initialized = true
	return nil
}

// initializeAgents 初始化Agent系统
func (c *Controller) initializeAgents() error {
	// 创建工具集
	availableTools := map[string]types.Tool{
		"scheduler_tool": tools.NewSchedulerTool(),
		"data_tool":      tools.NewDataTool(),
		"system_tool":    tools.NewSystemTool(),
	}

	// 基于配置初始化Agent
	for name, agentSpec := range c.config.Agent.Agents {
		// 获取指定的MCP客户端
		mcpClient := c.mcpClients["default"]

		// 准备工具列表
		agentTools := []types.Tool{}
		for _, toolName := range agentSpec.Tools {
			if tool, ok := availableTools[toolName]; ok {
				agentTools = append(agentTools, tool)
			}
		}

		// 创建Agent配置
		agentConfig := core.AgentConfig{
			Name:         agentSpec.Name,
			Description:  agentSpec.Description,
			Model:        agentSpec.Model,
			SystemPrompt: agentSpec.SystemPrompt,
			Memory:       c.config.Agent.Memory,
			Tools:        agentSpec.Tools,
		}

		// 创建Agent
		agent, err := core.NewAgent(agentConfig, mcpClient, agentTools)
		if err != nil {
			return fmt.Errorf("failed to create agent %s: %w", name, err)
		}

		c.agents[name] = agent
	}

	return nil
}

// initializeRAG 初始化RAG系统
func (c *Controller) initializeRAG(_ context.Context) error {
	// 为每个配置的检索器创建RAG系统
	for name, retrieveConfig := range c.config.RAG.Retriever {
		// 获取嵌入提供者配置
		embeddingConfig, ok := c.config.RAG.Embeddings[retrieveConfig.EmbeddingConfig.Type]
		if !ok {
			return fmt.Errorf("embedding provider %s not found", retrieveConfig.EmbeddingConfig.Type)
		}

		// 创建嵌入提供者
		embeddingProvider, err := embedding.NewProvider(embeddingConfig)
		if err != nil {
			return fmt.Errorf("failed to create embedding provider: %w", err)
		}

		// 获取向量存储配置
		vectorStoreConfig, ok := c.config.RAG.VectorStores[retrieveConfig.VectorStore]
		if !ok {
			return fmt.Errorf("vector store %s not found", retrieveConfig.VectorStore)
		}

		// 创建向量存储
		vectorStore, err := vectorstore.New(vectorStoreConfig, embeddingProvider)
		if err != nil {
			return fmt.Errorf("failed to create vector store: %w", err)
		}

		// 创建检索引擎
		engine, err := retriever.NewEngine(vectorStore, retrieveConfig)
		if err != nil {
			return fmt.Errorf("failed to create retriever engine: %w", err)
		}

		// 使用默认MCP配置

		// 创建生成器
		gen, err := generator.NewGenerator(engine, generator.Config{
			SystemPrompt:  "你是一个有用的AI助手，会根据提供的信息回答用户问题，并保持客观准确。",
			MaxSourceDocs: 5,
			MCPConfig:     client.Config{Provider: "local", Model: "deepseekv3-7b"},
		})
		if err != nil {
			return fmt.Errorf("failed to create generator: %w", err)
		}

		// 保存RAG系统
		c.ragSystems[name] = &RAGSystem{
			VectorStore: vectorStore,
			Retriever:   engine,
			Generator:   gen,
		}
	}

	return nil
}

// GetAgent 获取指定名称的Agent
func (c *Controller) GetAgent(name string) (*core.Agent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	agent, ok := c.agents[name]
	if !ok {
		return nil, fmt.Errorf("agent %s not found", name)
	}
	return agent, nil
}

// GetMCPClient 获取指定名称的MCP客户端
func (c *Controller) GetMCPClient(name string) (client.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	mcpClient, ok := c.mcpClients[name]
	if !ok {
		return nil, fmt.Errorf("MCP client %s not found", name)
	}
	return mcpClient, nil
}

// GetRAGSystem 获取指定名称的RAG系统
func (c *Controller) GetRAGSystem(name string) (*RAGSystem, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	ragSystem, ok := c.ragSystems[name]
	if !ok {
		return nil, fmt.Errorf("RAG system %s not found", name)
	}
	return ragSystem, nil
}

// ProcessQuery 处理用户查询，整合RAG和Agent
func (c *Controller) ProcessQuery(ctx context.Context, query string, agentName string, ragName string) (string, error) {
	// 获取RAG系统
	ragSystem, err := c.GetRAGSystem(ragName)
	if err != nil {
		// 如果找不到指定的RAG系统，尝试使用默认系统
		ragSystem, err = c.GetRAGSystem("default")
		if err != nil {
			// 如果没有RAG系统，直接使用Agent处理
			agent, agentErr := c.GetAgent(agentName)
			if agentErr != nil {
				return "", fmt.Errorf("both RAG and Agent not found: %v, %v", err, agentErr)
			}
			return agent.Process(ctx, query)
		}
	}

	// 使用RAG生成增强回答
	ragResponse, err := ragSystem.Generator.Generate(ctx, query, nil)
	if err != nil {
		return "", fmt.Errorf("RAG generation failed: %w", err)
	}

	// 如果需要进一步处理，可以将RAG结果传递给Agent
	agent, err := c.GetAgent(agentName)
	if err == nil {
		// 构建增强上下文
		enhancedContext := fmt.Sprintf("用户查询: %s\n\n相关信息: %s", query, ragResponse.Answer)
		return agent.Process(ctx, enhancedContext)
	}

	// 否则直接返回RAG结果
	return ragResponse.Answer, nil
}

// Chat 简单的MCP聊天功能
func (c *Controller) Chat(ctx context.Context, messages []protocol.Message, clientName string) (string, protocol.Usage, error) {
	mcpClient, err := c.GetMCPClient(clientName)
	if err != nil {
		// 尝试使用默认客户端
		mcpClient, err = c.GetMCPClient("default")
		if err != nil {
			return "", protocol.Usage{}, fmt.Errorf("MCP client not available: %w", err)
		}
	}

	response, err := mcpClient.Chat(ctx, &protocol.ChatRequest{
		Messages: messages,
	})
	if err != nil {
		return "", protocol.Usage{}, err
	}

	return response.Content, response.Usage, nil
}

// StreamChat 流式MCP聊天功能
func (c *Controller) StreamChat(ctx context.Context, messages []protocol.Message, clientName string) (<-chan protocol.ChatResponse, <-chan error, error) {
	mcpClient, err := c.GetMCPClient(clientName)
	if err != nil {
		// 尝试使用默认客户端
		mcpClient, err = c.GetMCPClient("default")
		if err != nil {
			return nil, nil, fmt.Errorf("MCP client not available: %w", err)
		}
	}

	return mcpClient.StreamChat(ctx, &protocol.ChatRequest{
		Messages: messages,
		Stream:   true,
	})
}

// ListModels 获取可用模型列表
func (c *Controller) ListModels(ctx context.Context) ([]ModelInfo, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var models []ModelInfo

	// 遍历所有客户端，获取模型信息
	for _, mcpClient := range c.mcpClients {
		provider := mcpClient.GetProvider()
		modelName := mcpClient.GetModel()

		// 创建模型信息
		model := ModelInfo{
			ID:          modelName,
			Name:        modelName,
			Provider:    provider,
			Description: "MCP model: " + modelName,
			MaxTokens:   4096, // 默认值
			LocalModel:  provider == "local",
			Cost:        0.0, // 默认值
		}

		models = append(models, model)
	}

	return models, nil
}

// GetModelUsage 获取模型使用统计
func (c *Controller) GetModelUsage(ctx context.Context) (map[string]interface{}, error) {
	// 这里可以实现更复杂的使用统计
	// 当前实现仅返回简单的统计信息
	usage := map[string]interface{}{
		"total_requests": 0,
		"total_tokens":   0,
		"models": map[string]interface{}{
			"default": map[string]interface{}{
				"requests": 0,
				"tokens":   0,
			},
		},
	}

	return usage, nil
}
