// Package handler provides API handlers for the application's endpoints
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"distributedJob/internal/service"
)

// IntegratedAIHandler 处理整合的AI相关API请求
type IntegratedAIHandler struct {
	aiService *service.AIService
}

// NewIntegratedAIHandler 创建一个新的整合AI处理器
func NewIntegratedAIHandler(aiService *service.AIService) *IntegratedAIHandler {
	return &IntegratedAIHandler{aiService: aiService}
}

// IntegratedQueryRequest 表示集成AI查询请求
type IntegratedQueryRequest struct {
	Query       string                 `json:"query" binding:"required"`
	AgentID     string                 `json:"agent_id,omitempty"`
	ModelID     string                 `json:"model_id,omitempty"`
	UseRAG      bool                   `json:"use_rag,omitempty"`
	UseAgent    bool                   `json:"use_agent,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Temperature float32                `json:"temperature,omitempty"`
}

// IntegratedQueryResponse 表示集成AI查询响应
type IntegratedQueryResponse struct {
	Answer      string      `json:"answer"`
	ProcessedBy string      `json:"processed_by"`
	AgentSteps  interface{} `json:"agent_steps,omitempty"`
	Sources     []Source    `json:"sources,omitempty"`
}

// Source 表示回答的来源
type Source struct {
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Relevance float64                `json:"relevance"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// HandleQuery 处理集成AI查询
// @Summary 处理集成AI查询
// @Description 使用Agent、RAG和MCP集成功能处理用户查询
// @Tags ai
// @Accept json
// @Produce json
// @Param request body IntegratedQueryRequest true "AI请求"
// @Success 200 {object} IntegratedQueryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/integrated-query [post]
func (h *IntegratedAIHandler) HandleQuery(c *gin.Context) {
	var req IntegratedQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的请求参数: " + err.Error()})
		return
	}

	// 使用集成方法处理查询
	answer, sources, steps, err := h.aiService.IntegrateQueryWithAll(
		c.Request.Context(),
		req.Query,
		req.UseRAG,
		req.UseAgent,
		req.AgentID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "处理查询失败: " + err.Error()})
		return
	}

	// 确定处理方法
	var processingMethod string
	if req.UseAgent && req.AgentID != "" {
		processingMethod = "Agent"
	} else if req.UseRAG {
		processingMethod = "RAG"
	} else {
		processingMethod = "MCP"
	}

	// 构建响应
	response := IntegratedQueryResponse{
		Answer:      answer,
		ProcessedBy: processingMethod,
		AgentSteps:  steps,
	}

	// 如果有来源，添加到响应
	if sources != nil && len(sources) > 0 {
		for _, src := range sources {
			response.Sources = append(response.Sources, Source{
				Title:     src.DocumentTitle,
				Content:   src.Content,
				Relevance: src.Relevance,
				Metadata:  src.Metadata,
			})
		}
	}

	c.JSON(http.StatusOK, response)
}

// RegisterRoutes 注册集成AI相关的路由
func (h *IntegratedAIHandler) RegisterRoutes(router gin.IRouter) {
	aiGroup := router.Group("/api/v1/ai")
	{
		aiGroup.POST("/integrated-query", h.HandleQuery)
	}
}
