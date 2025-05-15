// Package handler provides API handlers for the application's endpoints
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"distributedJob/internal/mcp/protocol"
	"distributedJob/internal/model/entity"
	"distributedJob/internal/service"
)

// AIHandler 集成AI相关API请求的处理
type AIHandler struct {
	aiService *service.AIService

	// 子处理器
	agentHandler *AgentHandler
	mcpHandler   *MCPHandler
	ragHandler   *RAGHandler
}

// NewAIHandler 创建一个新的AI处理器
func NewAIHandler(aiService *service.AIService) *AIHandler {
	handler := &AIHandler{
		aiService: aiService,
	}

	// 初始化子处理器
	handler.agentHandler = NewAgentHandler(aiService)
	handler.mcpHandler = NewMCPHandler(aiService)
	handler.ragHandler = NewRAGHandler(aiService)

	return handler
}

// ErrorResponse 表示错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// 代理相关处理函数 - 委托给AgentHandler

// CreateAgent 创建新智能代理
func (h *AIHandler) CreateAgent(c *gin.Context) {
	h.agentHandler.CreateAgent(c)
}

// GetAgent 获取智能代理详情
func (h *AIHandler) GetAgent(c *gin.Context) {
	h.agentHandler.GetAgent(c)
}

// ListAgents 获取智能代理列表
func (h *AIHandler) ListAgents(c *gin.Context) {
	h.agentHandler.ListAgents(c)
}

// ExecuteAgent 指派智能代理执行任务
func (h *AIHandler) ExecuteAgent(c *gin.Context) {
	h.agentHandler.ExecuteAgent(c)
}

// DeleteAgent 删除智能代理
func (h *AIHandler) DeleteAgent(c *gin.Context) {
	h.agentHandler.DeleteAgent(c)
}

// MCP相关处理函数 - 委托给MCPHandler

// ListModels 获取可用模型列表
func (h *AIHandler) ListModels(c *gin.Context) {
	h.mcpHandler.ListModels(c)
}

// Chat 发送对话请求
func (h *AIHandler) Chat(c *gin.Context) {
	h.mcpHandler.Chat(c)
}

// StreamChat 处理流式聊天请求
func (h *AIHandler) StreamChat(c *gin.Context) {
	h.mcpHandler.StreamChat(c)
}

// GetUsage 获取模型使用统计
func (h *AIHandler) GetUsage(c *gin.Context) {
	h.mcpHandler.GetUsage(c)
}

// RAG相关处理函数 - 委托给RAGHandler

// UploadDocument 上传并索引文档
func (h *AIHandler) UploadDocument(c *gin.Context) {
	h.ragHandler.UploadDocument(c)
}

// ListDocuments 获取已索引文档列表
func (h *AIHandler) ListDocuments(c *gin.Context) {
	h.ragHandler.ListDocuments(c)
}

// GetDocument 获取索引文档详情
func (h *AIHandler) GetDocument(c *gin.Context) {
	h.ragHandler.GetDocument(c)
}

// DeleteDocument 删除索引文档
func (h *AIHandler) DeleteDocument(c *gin.Context) {
	h.ragHandler.DeleteDocument(c)
}

// Query 提交RAG查询请求
func (h *AIHandler) Query(c *gin.Context) {
	h.ragHandler.Query(c)
}

// 集成功能处理

// AIAssistantRequest 表示AI助手请求
type AIAssistantRequest struct {
	Query       string                 `json:"query" binding:"required"`
	UseRAG      bool                   `json:"use_rag"`
	UseAgent    bool                   `json:"use_agent"`
	AgentID     string                 `json:"agent_id,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Temperature float32                `json:"temperature,omitempty"`
}

// AIAssistantResponse 表示AI助手响应
type AIAssistantResponse struct {
	Answer      string        `json:"answer"`
	Sources     []QuerySource `json:"sources,omitempty"`
	AgentSteps  interface{}   `json:"agent_steps,omitempty"`
	ProcessTime int64         `json:"process_time_ms"`
}

// QuerySource 表示查询结果的来源信息
type QuerySource struct {
	DocumentID    string                 `json:"document_id"`
	DocumentTitle string                 `json:"document_title"`
	Content       string                 `json:"content"`
	Relevance     float64                `json:"relevance"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// HandleQuery 处理综合AI助手查询
// @Summary 综合AI助手查询
// @Description 使用集成的AI能力(Agent+RAG+MCP)处理查询
// @Tags ai
// @Accept json
// @Produce json
// @Param request body AIAssistantRequest true "助手请求"
// @Success 200 {object} AIAssistantResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/assistant [post]
func (h *AIHandler) HandleQuery(c *gin.Context) {
	var req AIAssistantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的请求参数: " + err.Error()})
		return
	}

	startTime := time.Now()

	ctx := c.Request.Context()
	var answer string
	var sources []entity.RAGSource
	var agentSteps interface{}
	var err error

	// 根据请求配置决定使用哪些功能
	if req.UseRAG && req.UseAgent {
		// 使用完整集成处理
		answer, sources, agentSteps, err = h.aiService.HandleIntegratedQuery(ctx, req.Query, req.AgentID, req.Model, req.Filters)
	} else if req.UseRAG {
		// 仅使用RAG
		answer, sources, err = h.aiService.QueryWithRAG(ctx, req.Query, 3, req.Filters, req.Model)
	} else if req.UseAgent {
		// 仅使用Agent
		if req.AgentID == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "使用Agent时必须提供agent_id"})
			return
		}
		answer, agentSteps, err = h.aiService.HandleAgentQuery(ctx, req.AgentID, req.Query)
	} else {
		// 仅使用MCP
		chatReq := &protocol.ChatRequest{
			Messages: []protocol.Message{
				{Role: "user", Content: req.Query},
			},
			Temperature: req.Temperature,
			Model:       req.Model,
		}

		response, err := h.aiService.Chat(ctx, chatReq)
		if err == nil {
			answer = response.Content
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "处理查询失败: " + err.Error()})
		return
	}

	// 构建响应
	var sourcesResponse []QuerySource
	for _, src := range sources {
		sourcesResponse = append(sourcesResponse, QuerySource{
			DocumentID:    src.DocumentID,
			DocumentTitle: src.DocumentTitle,
			Content:       src.Content,
			Relevance:     src.Relevance,
			Metadata:      src.Metadata,
		})
	}

	processingTime := time.Since(startTime).Milliseconds()

	c.JSON(http.StatusOK, AIAssistantResponse{
		Answer:      answer,
		Sources:     sourcesResponse,
		AgentSteps:  agentSteps,
		ProcessTime: processingTime,
	})
}

// RegisterRoutes 注册API路由
func (h *AIHandler) RegisterRoutes(router gin.IRouter) {
	// AI集成接口
	aiGroup := router.Group("/api/v1/ai")
	{
		// AI助手接口 - 集成了Agent、MCP和RAG功能
		aiGroup.POST("/assistant", h.HandleQuery)
	}

	// 智能代理相关API
	agentGroup := router.Group("/api/v1/agents")
	{
		agentGroup.POST("", h.CreateAgent)
		agentGroup.GET("", h.ListAgents)
		agentGroup.GET("/:id", h.GetAgent)
		agentGroup.DELETE("/:id", h.DeleteAgent)
		agentGroup.POST("/:id/execute", h.ExecuteAgent)
	}

	// MCP相关API
	mcpGroup := router.Group("/api/v1/mcp")
	{
		mcpGroup.GET("/models", h.ListModels)
		mcpGroup.POST("/chat", h.Chat)
		mcpGroup.POST("/stream-chat", h.StreamChat)
	}

	// RAG相关API
	ragGroup := router.Group("/api/v1/rag")
	{
		ragGroup.POST("/documents", h.UploadDocument)
		ragGroup.GET("/documents", h.ListDocuments)
		ragGroup.GET("/documents/:id", h.GetDocument)
		ragGroup.DELETE("/documents/:id", h.DeleteDocument)
		ragGroup.POST("/query", h.Query)
	}
}
