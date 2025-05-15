// Package handler provides API handlers for the application's endpoints
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"distributedJob/internal/mcp/protocol"
	"distributedJob/internal/service"
)

// MCPHandler 处理Model Context Protocol相关的API请求
type MCPHandler struct {
	aiService *service.AIService
}

// NewMCPHandler 创建一个新的MCP处理器
func NewMCPHandler(aiService *service.AIService) *MCPHandler {
	return &MCPHandler{aiService: aiService}
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

// ListModelsResponse 表示可用模型列表响应
type ListModelsResponse struct {
	Models []ModelInfo `json:"models"`
}

// ListModels 获取可用模型列表
// @Summary 获取可用AI模型列表
// @Description 获取所有可用的AI模型信息
// @Tags mcp
// @Produce json
// @Success 200 {object} ListModelsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/mcp/models [get]
func (h *MCPHandler) ListModels(c *gin.Context) {
	models, err := h.aiService.ListModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取模型列表失败: " + err.Error()})
		return
	}

	var response ListModelsResponse
	for _, model := range models {
		response.Models = append(response.Models, ModelInfo{
			ID:          model.ID,
			Name:        model.Name,
			Provider:    model.Provider,
			Description: model.Description,
			MaxTokens:   model.MaxTokens,
			LocalModel:  model.LocalModel,
			Cost:        model.Cost,
		})
	}

	c.JSON(http.StatusOK, response)
}

// Message 表示聊天消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MCPChatRequest 表示聊天请求
type MCPChatRequest struct {
	Messages    []Message `json:"messages" binding:"required"`
	Model       string    `json:"model,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// MCPChatResponse 表示聊天响应
type MCPChatResponse struct {
	Content string       `json:"content"`
	Usage   UsageMetrics `json:"usage,omitempty"`
}

// UsageMetrics 表示使用指标
type UsageMetrics struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Chat 发送对话请求
// @Summary 发送对话请求
// @Description 向指定模型发送对话请求并获取回复
// @Tags mcp
// @Accept json
// @Produce json
// @Param request body ChatRequest true "对话请求"
// @Success 200 {object} ChatResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/mcp/chat [post]
func (h *MCPHandler) Chat(c *gin.Context) {
	var req MCPChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的请求参数: " + err.Error()})
		return
	}

	// 使用简化的ChatWithAI方法
	// 组合所有消息为一个字符串
	var messageText string
	for _, msg := range req.Messages {
		messageText += msg.Role + ": " + msg.Content + "\n"
	}

	// 如果没有指定模型，使用默认模型
	modelID := req.Model
	if modelID == "" {
		modelID = "default"
	}

	content, err := h.aiService.ChatWithAI(c.Request.Context(), messageText, modelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "聊天请求失败: " + err.Error()})
		return
	}

	// 模拟使用计数
	promptTokens := len(messageText) / 4 // 非精确估算
	completionTokens := len(content) / 4 // 非精确估算
	totalTokens := promptTokens + completionTokens

	c.JSON(http.StatusOK, MCPChatResponse{
		Content: content,
		Usage: UsageMetrics{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		},
	})
}

// StreamChat 处理流式聊天请求
// @Summary 流式对话
// @Description 向指定模型发送对话请求并获取流式回复
// @Tags mcp
// @Accept json
// @Produce text/event-stream
// @Param request body ChatRequest true "对话请求"
// @Success 200 {string} string "数据流"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/mcp/stream-chat [post]
func (h *MCPHandler) StreamChat(c *gin.Context) {
	var req MCPChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的请求参数: " + err.Error()})
		return
	}

	// 设置SSE头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 组合所有消息为一个字符串
	var messageText string
	for _, msg := range req.Messages {
		messageText += msg.Role + ": " + msg.Content + "\n"
	}

	// 如果没有指定模型，使用默认模型
	modelID := req.Model
	if modelID == "" {
		modelID = "default"
	}

	// 对于简化实现，我们将使用非流式接口并模拟流式返回
	content, err := h.aiService.ChatWithAI(c.Request.Context(), messageText, modelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "流式聊天请求失败: " + err.Error()})
		return
	}

	// 创建响应和错误通道
	responseCh := make(chan protocol.ChatResponse)
	errCh := make(chan error)

	// 启动一个goroutine来模拟流式响应
	go func() {
		defer close(responseCh)
		defer close(errCh)

		// 将内容分割为多个块进行流式传输
		chunkSize := 10
		for i := 0; i < len(content); i += chunkSize {
			end := i + chunkSize
			if end > len(content) {
				end = len(content)
			}
			chunk := content[i:end]

			// 发送块
			responseCh <- protocol.ChatResponse{Content: chunk}

			// 模拟网络延迟
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// 清除内部缓冲，确保实时数据传输
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	// 监听客户端连接关闭
	clientGone := c.Request.Context().Done()

	// 发送数据流
	for {
		select {
		case <-clientGone:
			// 客户端已断开连接
			return
		case response, ok := <-responseCh:
			if !ok {
				// 通道已关闭
				return
			}
			// 发送数据块
			c.SSEvent("message", response.Content)
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
		case err, ok := <-errCh:
			if !ok {
				// 错误通道已关闭
				return
			}
			// 发送错误
			c.SSEvent("error", err.Error())
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
			return
		}
	}
}

// UsageResponse 表示模型使用统计响应
type UsageResponse struct {
	TotalCalls  int                   `json:"total_calls"`
	TotalTokens int                   `json:"total_tokens"`
	ModelUsage  map[string]ModelUsage `json:"model_usage"`
}

// ModelUsage 表示单个模型的使用统计
type ModelUsage struct {
	Calls            int     `json:"calls"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	EstimatedCost    float64 `json:"estimated_cost"`
}

// GetUsage 获取模型使用统计
// @Summary 获取模型使用统计
// @Description 获取AI模型使用统计和成本估算
// @Tags mcp
// @Produce json
// @Success 200 {object} UsageResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/mcp/usage [get]
func (h *MCPHandler) GetUsage(c *gin.Context) {
	usage, err := h.aiService.GetModelUsage(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取使用统计失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// RegisterRoutes 注册MCP处理器的路由
func (h *MCPHandler) RegisterRoutes(router gin.IRouter) {
	mcpGroup := router.Group("/api/v1/mcp")
	{
		mcpGroup.GET("/models", h.ListModels)
		mcpGroup.POST("/chat", h.Chat)
		mcpGroup.POST("/stream-chat", h.StreamChat)
		mcpGroup.GET("/usage", h.GetUsage)
	}
}
