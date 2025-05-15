// Package service provides business logic services
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"distributedJob/internal/agent/core"
	"distributedJob/internal/agent/tools"
	"distributedJob/internal/agent/types"
	"distributedJob/internal/ai"
	"distributedJob/internal/config"
	"distributedJob/internal/mcp/protocol"
	"distributedJob/internal/model/entity"
	"distributedJob/internal/rag/document"
	"distributedJob/internal/rag/generator"
	"distributedJob/internal/rag/vectorstore"
)

// AIService 提供AI功能服务
type AIService struct {
	aiController *ai.Controller
	config       *config.AIConfig
}

// NewAIService 创建一个新的AI服务
func NewAIService(config *config.AIConfig) *AIService {
	return &AIService{
		aiController: ai.NewController(config),
		config:       config,
	}
}

// Initialize 初始化AI服务
func (s *AIService) Initialize(ctx context.Context) error {
	return s.aiController.Initialize(ctx)
}

// ---------- MCP (Model Context Protocol) 相关方法 ----------

// Chat 处理聊天请求
func (s *AIService) Chat(ctx context.Context, request *protocol.ChatRequest) (*protocol.ChatResponse, error) {
	content, usage, err := s.aiController.Chat(ctx, request.Messages, request.Model)
	if err != nil {
		return nil, err
	}
	return &protocol.ChatResponse{
		Content: content,
		Usage:   usage,
	}, nil
}

// StreamChat 处理流式聊天请求
func (s *AIService) StreamChat(ctx context.Context, request *protocol.ChatRequest) (<-chan protocol.ChatResponse, <-chan error, error) {
	responseChan, errChan, err := s.aiController.StreamChat(ctx, request.Messages, request.Model)
	if err != nil {
		return nil, nil, err
	}
	return responseChan, errChan, nil
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

// ListModels 获取可用模型列表
func (s *AIService) ListModels(ctx context.Context) ([]ModelInfo, error) {
	models, err := s.aiController.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	var result []ModelInfo
	for _, model := range models {
		result = append(result, ModelInfo{
			ID:          model.ID,
			Name:        model.Name,
			Provider:    model.Provider,
			Description: model.Description,
			MaxTokens:   model.MaxTokens,
			LocalModel:  model.LocalModel,
			Cost:        model.Cost,
		})
	}

	return result, nil
}

// GetModelUsage 获取模型使用统计
func (s *AIService) GetModelUsage(ctx context.Context) (map[string]interface{}, error) {
	return s.aiController.GetModelUsage(ctx)
}

// ProcessQuery 处理知识库查询
func (s *AIService) ProcessQuery(ctx context.Context, query string, agentName string, ragName string) (string, error) {
	return s.aiController.ProcessQuery(ctx, query, agentName, ragName)
}

// ProcessAgentTask 使用智能代理处理任务
func (s *AIService) ProcessAgentTask(ctx context.Context, input string, agentName string) (string, error) {
	agent, err := s.aiController.GetAgent(agentName)
	if err != nil {
		return "", fmt.Errorf("agent not found: %w", err)
	}

	return agent.Process(ctx, input)
}

// IndexDocument 索引文档到RAG系统
func (s *AIService) IndexDocument(ctx context.Context, document entity.Document, ragSystemName string) error {
	ragSystem, err := s.aiController.GetRAGSystem(ragSystemName)
	if err != nil {
		return fmt.Errorf("RAG system not found: %w", err)
	}

	// 转换为向量存储文档格式
	content := document.Content
	doc := vectorstore.Document{
		ID: strconv.FormatUint(uint64(document.ID), 10),
		Metadata: map[string]interface{}{
			"content":    content,
			"title":      document.Title,
			"source":     document.Source,
			"created_at": document.CreatedAt,
			"updated_at": document.UpdatedAt,
		},
	}

	// 如果有额外元数据，添加到文档
	if document.Metadata != nil {
		for k, v := range document.Metadata {
			doc.Metadata[k] = v
		}
	}

	// 添加到向量存储
	return ragSystem.VectorStore.Add(ctx, []vectorstore.Document{doc})
}

// IndexFile 将文件索引到RAG系统
func (s *AIService) IndexFile(ctx context.Context, filePath, ragSystemName string) error {
	ragSystem, err := s.aiController.GetRAGSystem(ragSystemName)
	if err != nil {
		return fmt.Errorf("RAG system not found: %w", err)
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// 创建基本元数据
	metadata := map[string]interface{}{
		"source":     filePath,
		"filename":   filepath.Base(filePath),
		"extension":  strings.TrimPrefix(filepath.Ext(filePath), "."),
		"indexed_at": time.Now(),
	}

	// 创建分块器
	chunker := document.NewRecursiveCharacterTextSplitter(1000, 100)

	// 创建文档
	doc := document.Document{
		ID:       filepath.Base(filePath),
		Content:  string(content),
		Metadata: metadata,
	}

	// 分块
	chunks, err := chunker.SplitDocument(doc)
	if err != nil {
		return fmt.Errorf("failed to split document: %w", err)
	}

	// 转换为向量存储文档格式并添加
	var docs []vectorstore.Document
	for i, chunk := range chunks {
		chunkDoc := vectorstore.Document{
			ID: fmt.Sprintf("%s_chunk_%d", doc.ID, i),
			Metadata: map[string]interface{}{
				"content":     chunk.Content,
				"chunk_index": i,
				"parent_id":   doc.ID,
			},
		}

		// 复制元数据
		for k, v := range metadata {
			chunkDoc.Metadata[k] = v
		}

		docs = append(docs, chunkDoc)
	}

	// 添加到向量存储
	return ragSystem.VectorStore.Add(ctx, docs)
}

// ListAvailableModels 列出可用模型
func (s *AIService) ListAvailableModels(ctx context.Context, provider string) ([]string, error) {
	client, err := s.aiController.GetMCPClient(provider)
	if err != nil {
		return nil, fmt.Errorf("MCP client not found: %w", err)
	}

	// 如果是本地客户端，尝试列出可用模型
	if client.GetProvider() == "local" {
		// 类型断言为本地客户端
		if localClient, ok := client.(interface {
			ListAvailableModels(context.Context) ([]string, error)
		}); ok {
			return localClient.ListAvailableModels(ctx)
		}
	}

	// 否则返回配置中的模型
	return []string{client.GetModel()}, nil
}

// QueryRAG 查询RAG系统
func (s *AIService) QueryRAG(ctx context.Context, query string, ragSystemName string) (*generator.Response, error) {
	ragSystem, err := s.aiController.GetRAGSystem(ragSystemName)
	if err != nil {
		return nil, fmt.Errorf("RAG system not found: %w", err)
	}

	// 使用生成器处理查询
	response, err := ragSystem.Generator.Generate(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}

// StreamRAG 流式查询RAG系统
func (s *AIService) StreamRAG(ctx context.Context, query string, ragSystemName string) (<-chan string, <-chan error, error) {
	ragSystem, err := s.aiController.GetRAGSystem(ragSystemName)
	if err != nil {
		return nil, nil, fmt.Errorf("RAG system not found: %w", err)
	}

	// 使用生成器处理流式查询
	return ragSystem.Generator.GenerateStream(ctx, query, nil)
}

// GetController 获取AI控制器
func (s *AIService) GetController() *ai.Controller {
	return s.aiController
}

// ---------- Agent 智能代理相关方法 ----------

// CreateAgent 创建一个新的智能代理
func (s *AIService) CreateAgent(ctx context.Context, config core.AgentConfig) (*core.Agent, error) {
	// 获取MCP客户端
	mcpClient, err := s.aiController.GetMCPClient(config.Model)
	if err != nil {
		mcpClient, err = s.aiController.GetMCPClient("default")
		if err != nil {
			return nil, fmt.Errorf("无法获取MCP客户端: %w", err)
		}
	}

	// 获取工具列表
	var agentTools []types.Tool
	for _, toolName := range config.Tools {
		tool, err := s.getToolByName(toolName)
		if err != nil {
			return nil, fmt.Errorf("无法获取工具 %s: %w", toolName, err)
		}
		agentTools = append(agentTools, tool)
	}

	// 创建Agent
	agent, err := core.NewAgent(config, mcpClient, agentTools)
	if err != nil {
		return nil, fmt.Errorf("创建智能代理失败: %w", err)
	}

	return agent, nil
}

// 内部方法：根据名称获取工具
func (s *AIService) getToolByName(name string) (types.Tool, error) {
	// 根据名称获取工具实例
	switch name {
	case "scheduler_tool":
		return tools.NewSchedulerTool(), nil
	case "data_tool":
		return tools.NewDataTool(), nil
	case "system_tool":
		return tools.NewSystemTool(), nil
	default:
		return nil, fmt.Errorf("未知的工具: %s", name)
	}
}

// GetAgent 获取智能代理
func (s *AIService) GetAgent(ctx context.Context, id string) (*core.Agent, error) {
	return s.aiController.GetAgent(id)
}

// ListAgents 获取所有智能代理
func (s *AIService) ListAgents(ctx context.Context) ([]*core.Agent, error) {
	// 获取所有Agent
	var agents []*core.Agent

	// 遍历Controller中的所有Agent
	for name := range s.config.Agent.Agents {
		agent, err := s.aiController.GetAgent(name)
		if err == nil {
			agents = append(agents, agent)
		}
	}

	return agents, nil
}

// ExecuteAgent 执行智能代理
func (s *AIService) ExecuteAgent(ctx context.Context, id string, input string) (string, error) {
	agent, err := s.aiController.GetAgent(id)
	if err != nil {
		return "", fmt.Errorf("智能代理未找到: %w", err)
	}

	return agent.Process(ctx, input)
}

// DeleteAgent 删除智能代理
func (s *AIService) DeleteAgent(ctx context.Context, id string) error {
	// 实现删除Agent的逻辑
	// 这里仅作为示例，不会真正删除

	return nil
}

// HandleAgentQuery 处理Agent查询
func (s *AIService) HandleAgentQuery(ctx context.Context, agentID string, query string) (string, interface{}, error) {
	agent, err := s.aiController.GetAgent(agentID)
	if err != nil {
		return "", nil, fmt.Errorf("智能代理未找到: %w", err)
	}

	// 执行Agent处理
	response, err := agent.Process(ctx, query)
	if err != nil {
		return "", nil, fmt.Errorf("智能代理处理失败: %w", err)
	}

	// 获取Agent步骤信息（实际实现需要从Agent获取执行步骤）
	steps := map[string]interface{}{
		"goal":        "回应用户查询",
		"description": "处理用户的问题并提供回答",
		"steps": []map[string]interface{}{
			{
				"id":          "step-1",
				"description": "分析用户查询",
				"tool":        "analysis",
				"completed":   true,
			},
			{
				"id":          "step-2",
				"description": "生成响应",
				"tool":        "generation",
				"completed":   true,
			},
		},
	}

	return response, steps, nil
}

// ---------- RAG 检索增强生成相关方法 ----------

// ProcessAndIndexDocument 处理并索引文档
func (s *AIService) ProcessAndIndexDocument(ctx context.Context, doc document.Document) (int, error) {
	// 从第一个可用的RAG系统中获取vectorstore
	var vectorStore vectorstore.VectorStore
	for name := range s.config.RAG.Retriever {
		ragSystem, err := s.aiController.GetRAGSystem(name)
		if err == nil {
			vectorStore = ragSystem.VectorStore
			break
		}
	}

	if vectorStore == nil {
		return 0, fmt.Errorf("没有可用的向量存储")
	}

	// 处理并索引文档
	processor := document.NewDefaultProcessor()
	processedDoc, err := processor.Process(doc)
	if err != nil {
		return 0, fmt.Errorf("文档处理失败: %w", err)
	}

	// 分割文档成chunks
	chunker := document.NewRecursiveCharacterTextSplitter(1000, 200)
	chunks, err := chunker.SplitDocument(processedDoc)
	if err != nil {
		return 0, fmt.Errorf("文档分块失败: %w", err)
	}

	// 索引所有分块
	for _, chunk := range chunks {
		// 转换为向量存储文档格式
		vecDoc := vectorstore.Document{
			ID:       chunk.ID,
			Content:  chunk.Content,
			Metadata: chunk.Metadata,
		}

		// 添加到向量存储
		err := vectorStore.Add(ctx, []vectorstore.Document{vecDoc})
		if err != nil {
			return 0, fmt.Errorf("文档索引失败: %w", err)
		}
	}

	// 更新文档的块计数
	doc.ChunkCount = len(chunks)

	return len(chunks), nil
}

// GetDocument 获取文档
func (s *AIService) GetDocument(ctx context.Context, id string) (*document.Document, error) {
	// 实现从存储中获取文档信息
	// 示例返回
	return &document.Document{
		ID:    id,
		Title: "示例文档",
		Metadata: map[string]interface{}{
			"author": "系统",
			"date":   time.Now(),
		},
		CreatedAt: time.Now(),
	}, nil
}

// ListDocuments 获取文档列表
func (s *AIService) ListDocuments(ctx context.Context, page, limit int) ([]document.Document, int, error) {
	// 实现获取文档列表
	// 示例返回
	docs := []document.Document{
		{
			ID:        "doc-1",
			Title:     "文档1",
			CreatedAt: time.Now(),
		},
		{
			ID:        "doc-2",
			Title:     "文档2",
			CreatedAt: time.Now(),
		},
	}

	return docs, len(docs), nil
}

// DeleteDocument 删除文档
func (s *AIService) DeleteDocument(ctx context.Context, id string) error {
	// 实现从向量存储中删除文档
	return nil
}

// QueryRAGWithConfig 执行带配置的RAG查询
func (s *AIService) QueryRAGWithConfig(ctx context.Context, query string, ragSystemName string, config map[string]interface{}) (*generator.Response, error) {
	ragSystem, err := s.aiController.GetRAGSystem(ragSystemName)
	if err != nil {
		return nil, fmt.Errorf("RAG system not found: %w", err)
	}

	// 使用生成器处理查询
	response, err := ragSystem.Generator.Generate(ctx, query, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}

// QueryWithRAG 执行带高级参数的RAG查询
func (s *AIService) QueryWithRAG(ctx context.Context, query string, topK int, filters map[string]interface{}, model string) (string, []entity.RAGSource, error) {
	// 使用默认RAG系统
	ragSystem, err := s.aiController.GetRAGSystem("default")
	if err != nil {
		return "", nil, fmt.Errorf("default RAG system not found: %w", err)
	}

	// 准备查询选项
	options := map[string]interface{}{}
	if topK > 0 {
		options["top_k"] = topK
	}
	if filters != nil && len(filters) > 0 {
		options["filters"] = filters
	}
	if model != "" {
		options["model"] = model
	}

	// 使用生成器处理查询
	response, err := ragSystem.Generator.Generate(ctx, query, options)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// 构造源文档列表
	var sources []entity.RAGSource
	for _, src := range response.Sources {
		sources = append(sources, entity.RAGSource{
			DocumentID:    src.Document.ID,
			DocumentTitle: src.Document.Metadata["title"].(string),
			Content:       src.Document.Content,
			Relevance:     float64(src.Score),
			Metadata:      src.Document.Metadata,
		})
	}

	return response.Answer, sources, nil
}

// ---------- 集成功能 ----------

// HandleIntegratedQuery 处理集成查询（结合Agent、RAG和MCP）
func (s *AIService) HandleIntegratedQuery(ctx context.Context, query string, agentID string, model string, filters map[string]interface{}) (string, []entity.RAGSource, interface{}, error) {
	// 1. 使用RAG检索相关信息
	answer, sources, err := s.QueryWithRAG(ctx, query, 3, filters, model)
	if err != nil {
		// 如果RAG失败，尝试使用Agent直接处理
		if agentID != "" {
			agentAnswer, agentSteps, agentErr := s.HandleAgentQuery(ctx, agentID, query)
			if agentErr == nil {
				return agentAnswer, nil, agentSteps, nil
			}
		}
		return "", nil, nil, fmt.Errorf("查询处理失败: %w", err)
	}

	// 2. 如果指定了Agent，使用Agent处理增强后的查询
	if agentID != "" {
		agent, err := s.aiController.GetAgent(agentID)
		if err == nil {
			// 构建增强上下文
			enhancedContext := fmt.Sprintf("用户查询: %s\n\n相关信息: %s", query, answer)

			// 让Agent处理增强后的上下文
			agentResponse, err := agent.Process(ctx, enhancedContext)
			if err == nil {
				// 获取Agent步骤信息
				steps := map[string]interface{}{
					"goal":        "处理RAG增强的查询",
					"description": "使用检索到的信息回答用户问题",
				}

				return agentResponse, sources, steps, nil
			}
		}
	}

	// 如果Agent处理失败或未指定Agent，返回RAG结果
	return answer, sources, nil, nil
}
