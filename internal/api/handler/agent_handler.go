// Package handler provides API handlers for the application's endpoints
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"distributedJob/internal/agent/core"
	"distributedJob/internal/agent/types"
	"distributedJob/internal/service"
)

// AgentHandler 处理智能代理相关的API请求
type AgentHandler struct {
	aiService *service.AIService
}

// NewAgentHandler 创建一个新的智能代理处理器
func NewAgentHandler(aiService *service.AIService) *AgentHandler {
	return &AgentHandler{aiService: aiService}
}

// AgentCreateRequest 表示创建智能代理的请求
type AgentCreateRequest struct {
	Name         string             `json:"name" binding:"required"`
	Description  string             `json:"description" binding:"required"`
	Model        string             `json:"model" binding:"required"`
	SystemPrompt string             `json:"system_prompt" binding:"required"`
	Tools        []string           `json:"tools" binding:"required"`
	Memory       types.MemoryConfig `json:"memory"`
}

// AgentResponse 表示智能代理的响应
type AgentResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Model       string   `json:"model"`
	Tools       []string `json:"tools"`
}

// CreateAgent 创建新智能代理
// @Summary 创建新的智能代理
// @Description 创建一个新的智能代理实例
// @Tags agent
// @Accept json
// @Produce json
// @Param agent body AgentCreateRequest true "智能代理配置"
// @Success 201 {object} AgentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents [post]
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req AgentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的请求参数: " + err.Error()})
		return
	}

	agentConfig := core.AgentConfig{
		Name:         req.Name,
		Description:  req.Description,
		Model:        req.Model,
		SystemPrompt: req.SystemPrompt,
		Tools:        req.Tools,
		Memory:       req.Memory,
	}

	agent, err := h.aiService.CreateAgent(c.Request.Context(), agentConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "创建智能代理失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, AgentResponse{
		ID:          agent.GetID(),
		Name:        agent.GetName(),
		Description: agent.GetDescription(),
		Model:       agent.GetModel(),
		Tools:       agent.GetTools(),
	})
}

// GetAgent 获取智能代理详情
// @Summary 获取智能代理详情
// @Description 获取指定ID的智能代理详细信息
// @Tags agent
// @Produce json
// @Param id path string true "智能代理ID"
// @Success 200 {object} AgentResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/{id} [get]
func (h *AgentHandler) GetAgent(c *gin.Context) {
	id := c.Param("id")
	agent, err := h.aiService.GetAgent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "智能代理未找到: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, AgentResponse{
		ID:          agent.GetID(),
		Name:        agent.GetName(),
		Description: agent.GetDescription(),
		Model:       agent.GetModel(),
		Tools:       agent.GetTools(),
	})
}

// ListAgents 获取智能代理列表
// @Summary 获取智能代理列表
// @Description 获取所有可用的智能代理
// @Tags agent
// @Produce json
// @Success 200 {array} AgentResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents [get]
func (h *AgentHandler) ListAgents(c *gin.Context) {
	agents, err := h.aiService.ListAgents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取智能代理列表失败: " + err.Error()})
		return
	}

	var response []AgentResponse
	for _, agent := range agents {
		response = append(response, AgentResponse{
			ID:          agent.GetID(),
			Name:        agent.GetName(),
			Description: agent.GetDescription(),
			Model:       agent.GetModel(),
			Tools:       agent.GetTools(),
		})
	}

	c.JSON(http.StatusOK, response)
}

// ExecuteAgentRequest 表示执行智能代理的请求
type ExecuteAgentRequest struct {
	Input string `json:"input" binding:"required"`
}

// ExecuteAgentResponse 表示智能代理执行的响应
type ExecuteAgentResponse struct {
	Output string      `json:"output"`
	Steps  interface{} `json:"steps,omitempty"`
}

// ExecuteAgent 指派智能代理执行任务
// @Summary 执行智能代理任务
// @Description 指派智能代理执行任务
// @Tags agent
// @Accept json
// @Produce json
// @Param id path string true "智能代理ID"
// @Param request body ExecuteAgentRequest true "执行请求"
// @Success 200 {object} ExecuteAgentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/{id}/execute [post]
func (h *AgentHandler) ExecuteAgent(c *gin.Context) {
	id := c.Param("id")
	var req ExecuteAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的请求参数: " + err.Error()})
		return
	}

	output, steps, err := h.aiService.ExecuteAgentAction(c.Request.Context(), id, req.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "智能代理执行失败: " + err.Error()})
		return
	}

	response := ExecuteAgentResponse{
		Output: output,
		Steps:  steps, // 添加步骤信息到响应
	}

	c.JSON(http.StatusOK, response)
}

// DeleteAgent 删除智能代理
// @Summary 删除智能代理
// @Description 删除指定ID的智能代理
// @Tags agent
// @Param id path string true "智能代理ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/agents/{id} [delete]
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	id := c.Param("id")
	err := h.aiService.DeleteAgent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "删除智能代理失败: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// RegisterRoutes 注册Agent处理器的路由
func (h *AgentHandler) RegisterRoutes(router gin.IRouter) {
	agentGroup := router.Group("/api/v1/agents")
	{
		agentGroup.GET("", h.ListAgents)
		agentGroup.POST("", h.CreateAgent)
		agentGroup.GET("/:id", h.GetAgent)
		agentGroup.DELETE("/:id", h.DeleteAgent)
		agentGroup.POST("/:id/execute", h.ExecuteAgent)
	}
}
