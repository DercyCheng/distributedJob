// Package handler provides API handlers for the application's endpoints
package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"distributedJob/internal/mcp/protocol"
	"distributedJob/internal/model/entity"
)

// OldAIHandler 处理AI相关API请求(旧版本)
type OldAIHandler struct {
	aiService interface{} // Using interface{} to avoid explicit dependency on AIService
}

// NewOldAIHandler 创建一个新的旧版AI处理器
func NewOldAIHandler(aiService interface{}) *OldAIHandler {
	return &OldAIHandler{aiService: aiService}
}

// OldErrorResponse 表示错误响应(旧版本)
type OldErrorResponse struct {
	Error string `json:"error"`
}

// OldChatRequest 表示聊天请求(旧版本)
type OldChatRequest struct {
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Model       string  `json:"model,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float32 `json:"temperature,omitempty"`
	Stream      bool    `json:"stream,omitempty"`
}

// OldChatResponse 表示聊天响应(旧版本)
type OldChatResponse struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// OldQueryRequest 表示RAG查询请求(旧版本)
type OldQueryRequest struct {
	Query    string                 `json:"query"`
	Filters  map[string]interface{} `json:"filters,omitempty"`
	AgentID  string                 `json:"agent_id,omitempty"`
	RAGModel string                 `json:"rag_model,omitempty"`
}

// OldQueryResponse 表示RAG查询响应(旧版本)
type OldQueryResponse struct {
	Answer  string `json:"answer"`
	Sources []struct {
		Content  string                 `json:"content"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
		Score    float32                `json:"score"`
	} `json:"sources,omitempty"`
	Error string `json:"error,omitempty"`
}

// DocumentRequest 表示文档请求
type DocumentRequest struct {
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Source   string                 `json:"source,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentResponse 表示文档响应
type DocumentResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// RegisterRoutes 注册API路由
func (h *OldAIHandler) RegisterRoutes(router gin.IRouter) {
	aiGroup := router.Group("/ai")
	{
		aiGroup.POST("/chat", h.handleChat)
		aiGroup.POST("/stream-chat", h.handleStreamChat)
		aiGroup.POST("/query", h.handleQuery)
		aiGroup.POST("/query/stream", h.handleStreamQuery)
		aiGroup.POST("/document", h.handleIndexDocument)
		aiGroup.POST("/document/file", h.handleIndexFile)
	}
}

// handleChat 处理聊天请求
func (h *OldAIHandler) handleChat(c *gin.Context) {
	var req OldChatRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 转换消息格式
	var messages []protocol.Message
	for _, msg := range req.Messages {
		messages = append(messages, protocol.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 默认使用配置的模型或请求中指定的模型
	modelName := req.Model
	if modelName == "" {
		modelName = "default" // 默认模型名称，从配置中获取
	}

	// 使用AI服务处理请求
	chatRequest := &protocol.ChatRequest{
		Messages: messages,
		Model:    modelName,
	}
	response, err := h.aiService.Chat(c.Request.Context(), chatRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process chat: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, OldChatResponse{
		Content: response.Content,
	})
}

// handleStreamChat 处理流式聊天请求
func (h *OldAIHandler) handleStreamChat(c *gin.Context) {
	var req OldChatRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 转换消息格式
	var messages []protocol.Message
	for _, msg := range req.Messages {
		messages = append(messages, protocol.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 使用AI服务处理流式请求
	chatRequest := &protocol.ChatRequest{
		Messages: messages,
		Model:    req.Model,
		Stream:   true,
	}
	responseChan, errChan, err := h.aiService.StreamChat(c.Request.Context(), chatRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start stream: " + err.Error()})
		return
	}

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 使用writer的Flush方法确保数据发送至客户端
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming unsupported"})
		return
	}

	// 客户端关闭连接时取消请求
	clientGone := c.Writer.CloseNotify()

	// 发送SSE事件
	for {
		select {
		case <-clientGone:
			return
		case err, ok := <-errChan:
			if !ok {
				// 错误通道已关闭，但可能还有消息
				continue
			}
			data, _ := json.Marshal(OldChatResponse{Error: err.Error()})
			c.Writer.Write([]byte("data: " + string(data) + "\n\n"))
			flusher.Flush()
			return
		case resp, ok := <-responseChan:
			if !ok {
				// 响应通道已关闭，流结束
				c.Writer.Write([]byte("data: [DONE]\n\n"))
				flusher.Flush()
				return
			}
			data, _ := json.Marshal(OldChatResponse{Content: resp.Content})
			c.Writer.Write([]byte("data: " + string(data) + "\n\n"))
			flusher.Flush()
		}
	}
}

// handleQuery 处理RAG查询
func (h *OldAIHandler) handleQuery(c *gin.Context) {
	var req OldQueryRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取RAG系统ID，默认为default
	ragID := req.RAGModel
	if ragID == "" {
		ragID = "default"
	}

	// 获取Agent ID，可选
	agentID := req.AgentID

	var response string
	var err error

	// 如果指定了Agent，则使用Agent处理
	if agentID != "" {
		response, err = h.aiService.ProcessQuery(c.Request.Context(), req.Query, agentID, ragID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process query: " + err.Error()})
			return
		}
	} else {
		// 否则直接使用RAG系统
		answer, _, err := h.aiService.QueryWithRAG(c.Request.Context(), req.Query, 3, nil, "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query RAG: " + err.Error()})
			return
		}
		response = answer
	}

	c.JSON(http.StatusOK, OldQueryResponse{
		Answer: response,
	})
}

// handleStreamQuery 处理流式RAG查询
func (h *OldAIHandler) handleStreamQuery(c *gin.Context) {
	var req OldQueryRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取RAG系统ID，默认为default
	ragID := req.RAGModel
	if ragID == "" {
		ragID = "default"
	}

	// 使用RAG系统的流式查询
	responseChan, errChan, err := h.aiService.StreamRAG(c.Request.Context(), req.Query, ragID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start stream: " + err.Error()})
		return
	}

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 使用writer的Flush方法确保数据发送至客户端
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming unsupported"})
		return
	}

	// 客户端关闭连接时取消请求
	clientGone := c.Writer.CloseNotify()

	// 发送SSE事件
	for {
		select {
		case <-clientGone:
			return
		case err, ok := <-errChan:
			if !ok {
				continue
			}
			data, _ := json.Marshal(OldChatResponse{Error: err.Error()})
			c.Writer.Write([]byte("data: " + string(data) + "\n\n"))
			flusher.Flush()
			return
		case content, ok := <-responseChan:
			if !ok {
				c.Writer.Write([]byte("data: [DONE]\n\n"))
				flusher.Flush()
				return
			}
			data, _ := json.Marshal(OldChatResponse{Content: content})
			c.Writer.Write([]byte("data: " + string(data) + "\n\n"))
			flusher.Flush()
		}
	}
}

// handleIndexDocument 处理文档索引请求
func (h *OldAIHandler) handleIndexDocument(c *gin.Context) {
	var req DocumentRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取RAG系统ID，默认为default
	ragID := c.Query("rag_id")
	if ragID == "" {
		ragID = "default"
	}

	// 创建文档实体
	doc := entity.Document{
		Title:    req.Title,
		Content:  req.Content,
		Source:   req.Source,
		Metadata: req.Metadata,
	}

	// 索引文档
	err := h.aiService.IndexDocument(c.Request.Context(), doc, ragID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index document: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, DocumentResponse{
		Status: "success",
	})
}

// handleIndexFile 处理文件索引请求，使用multipart表单上传
func (h *OldAIHandler) handleIndexFile(c *gin.Context) {
	// 获取RAG系统ID，默认为default
	ragID := c.PostForm("rag_id")
	if ragID == "" {
		ragID = "default"
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file: " + err.Error()})
		return
	}
	defer file.Close()

	// 创建临时文件
	tempFile, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file: " + err.Error()})
		return
	}

	// 获取文件元数据
	metadata := make(map[string]interface{})
	metadata["filename"] = header.Filename
	metadata["size"] = header.Size
	metadata["content_type"] = header.Header.Get("Content-Type")

	// 创建文档实体
	doc := entity.Document{
		Title:    header.Filename,
		Content:  string(tempFile),
		Source:   "file_upload",
		Metadata: metadata,
	}

	// 索引文档
	err = h.aiService.IndexDocument(c.Request.Context(), doc, ragID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, DocumentResponse{
		Status: "success",
	})
}
