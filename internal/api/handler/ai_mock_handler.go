// Package handler provides API handlers for the application's endpoints
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AIMockHandler 处理AI相关API请求 - 作为模拟实现
type AIMockHandler struct{}

// NewAIMockHandler 创建一个新的AI处理器模拟实现
func NewAIMockHandler() *AIMockHandler {
	return &AIMockHandler{}
}

// RegisterRoutes 注册API路由
func (h *AIMockHandler) RegisterRoutes(router gin.IRouter) {
	aiGroup := router.Group("/ai")
	{
		aiGroup.POST("/chat", h.handleChat)
		aiGroup.POST("/query", h.handleQuery)
		aiGroup.GET("/status", h.handleStatus)
	}
}

// 简单的聊天请求和响应结构体
type simpleChatRequest struct {
	Query string `json:"query"`
}

type simpleChatResponse struct {
	Answer      string    `json:"answer"`
	ProcessedAt time.Time `json:"processed_at"`
	Model       string    `json:"model"`
}

// handleChat 处理聊天请求
func (h *AIMockHandler) handleChat(c *gin.Context) {
	var req simpleChatRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	// 模拟AI响应
	response := simpleChatResponse{
		Answer:      "这是一个模拟的AI回复，您的查询是: " + req.Query,
		ProcessedAt: time.Now(),
		Model:       "deepseekv3-7b-模拟",
	}

	c.JSON(http.StatusOK, response)
}

// handleQuery 处理查询请求
func (h *AIMockHandler) handleQuery(c *gin.Context) {
	var req simpleChatRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	// 模拟RAG查询响应
	response := map[string]interface{}{
		"answer": "这是来自RAG系统的模拟回复，您的查询是: " + req.Query,
		"sources": []map[string]interface{}{
			{
				"content": "这是一个示例文档内容，与您的查询'" + req.Query + "'相关。",
				"score":   0.92,
				"metadata": map[string]interface{}{
					"title":  "示例文档1",
					"author": "AI团队",
				},
			},
		},
		"processed_at": time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// handleStatus 返回AI系统状态
func (h *AIMockHandler) handleStatus(c *gin.Context) {
	status := map[string]interface{}{
		"status": "operational",
		"models": []map[string]interface{}{
			{
				"name":   "deepseekv3-7b",
				"status": "ready",
				"type":   "local",
			},
			{
				"name":   "qwen3-7b",
				"status": "ready",
				"type":   "local",
			},
		},
		"services": map[string]interface{}{
			"mcp":   "online",
			"rag":   "online",
			"agent": "online",
		},
		"last_check": time.Now(),
	}

	c.JSON(http.StatusOK, status)
}
