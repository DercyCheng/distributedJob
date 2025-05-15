// Package handler provides API handlers for the application's endpoints
package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"distributedJob/internal/model/entity"
	"distributedJob/internal/rag/document"
	"distributedJob/internal/service"
)

// RAGHandler 处理检索增强生成相关的API请求
type RAGHandler struct {
	aiService *service.AIService
}

// NewRAGHandler 创建一个新的RAG处理器
func NewRAGHandler(aiService *service.AIService) *RAGHandler {
	return &RAGHandler{aiService: aiService}
}

// convertDocumentToEntity 将document.Document转换为entity.Document
func convertDocumentToEntity(doc document.Document) entity.Document {
	return entity.Document{
		Title:     doc.Title,
		Content:   doc.Content,
		Source:    doc.Metadata["source"].(string),
		Metadata:  doc.Metadata,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: time.Now(),
	}
}

// DocumentInfo 表示文档信息
type DocumentInfo struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	ContentType string                 `json:"content_type"`
	Size        int                    `json:"size"`
	ChunkCount  int                    `json:"chunk_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   string                 `json:"created_at"`
}

// UploadDocumentResponse 表示上传文档的响应
type UploadDocumentResponse struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
	ChunkCount int    `json:"chunk_count"`
	Message    string `json:"message"`
}

// UploadDocument 上传并索引文档
// @Summary 上传并索引文档
// @Description 上传文档并将其添加到RAG索引中
// @Tags rag
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "要索引的文档"
// @Param title formData string false "文档标题"
// @Param metadata formData string false "文档元数据 (JSON)"
// @Success 201 {object} UploadDocumentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rag/documents [post]
func (h *RAGHandler) UploadDocument(c *gin.Context) {
	// 从表单获取文件
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "文件上传失败: " + err.Error()})
		return
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "读取文件失败: " + err.Error()})
		return
	}

	// 获取文档标题，如果未提供则使用文件名
	title := c.PostForm("title")
	if title == "" {
		title = fileHeader.Filename
	}

	// 获取可选的元数据
	metadata := make(map[string]interface{})
	metadataStr := c.PostForm("metadata")
	if metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "元数据格式无效: " + err.Error()})
			return
		}
	}

	// 检测文档类型
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}

	// 创建文档对象
	doc := document.CreateDocumentFromFile(fileHeader, content, title, metadata)

	// 转换为entity.Document类型并索引文档
	entityDoc := convertDocumentToEntity(doc)
	err = h.aiService.IndexDocument(c.Request.Context(), entityDoc, "default")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "文档索引失败: " + err.Error()})
		return
	}

	chunkCount := doc.ChunkCount

	c.JSON(http.StatusCreated, UploadDocumentResponse{
		DocumentID: doc.ID,
		Title:      doc.Title,
		ChunkCount: chunkCount,
		Message:    "文档已成功索引",
	})
}

// ListDocumentsResponse 表示文档列表响应
type ListDocumentsResponse struct {
	Documents []DocumentInfo `json:"documents"`
	Total     int            `json:"total"`
}

// ListDocuments 获取已索引文档列表
// @Summary 获取已索引文档列表
// @Description 获取所有已添加到RAG索引中的文档
// @Tags rag
// @Produce json
// @Param page query int false "页码，默认为1"
// @Param limit query int false "每页数量，默认为20"
// @Success 200 {object} ListDocumentsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rag/documents [get]
func (h *RAGHandler) ListDocuments(c *gin.Context) {
	page := 1
	limit := 20

	// 获取分页参数
	if pageParam := c.Query("page"); pageParam != "" {
		if val, err := strconv.Atoi(pageParam); err == nil && val > 0 {
			page = val
		}
	}
	if limitParam := c.Query("limit"); limitParam != "" {
		if val, err := strconv.Atoi(limitParam); err == nil && val > 0 {
			limit = val
		}
	}

	// 获取文档列表
	docs, total, err := h.aiService.ListDocuments(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "获取文档列表失败: " + err.Error()})
		return
	}

	// 构建响应
	var response ListDocumentsResponse
	response.Total = total
	for _, doc := range docs {
		response.Documents = append(response.Documents, DocumentInfo{
			ID:          doc.ID,
			Title:       doc.Title,
			ContentType: doc.ContentType,
			Size:        len(doc.Content),
			ChunkCount:  doc.ChunkCount,
			Metadata:    doc.Metadata,
			CreatedAt:   doc.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetDocument 获取索引文档详情
// @Summary 获取索引文档详情
// @Description 获取指定ID的索引文档详情
// @Tags rag
// @Produce json
// @Param id path string true "文档ID"
// @Success 200 {object} DocumentInfo
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rag/documents/{id} [get]
func (h *RAGHandler) GetDocument(c *gin.Context) {
	id := c.Param("id")
	doc, err := h.aiService.GetDocument(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "文档未找到: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, DocumentInfo{
		ID:          doc.ID,
		Title:       doc.Title,
		ContentType: doc.ContentType,
		Size:        len(doc.Content),
		ChunkCount:  doc.ChunkCount,
		Metadata:    doc.Metadata,
		CreatedAt:   doc.CreatedAt.Format(time.RFC3339),
	})
}

// DeleteDocument 删除索引文档
// @Summary 删除索引文档
// @Description 删除指定ID的索引文档
// @Tags rag
// @Param id path string true "文档ID"
// @Success 204 "No Content"
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rag/documents/{id} [delete]
func (h *RAGHandler) DeleteDocument(c *gin.Context) {
	id := c.Param("id")
	err := h.aiService.DeleteDocument(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "删除文档失败: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// QueryRequest 表示RAG查询请求
type RAGQueryRequest struct {
	Query       string                 `json:"query" binding:"required"`
	TopK        int                    `json:"top_k,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Temperature float32                `json:"temperature,omitempty"`
}

// RAGQuerySource 表示查询结果的来源
type RAGQuerySource struct {
	DocumentID    string                 `json:"document_id"`
	DocumentTitle string                 `json:"document_title"`
	Content       string                 `json:"content"`
	Relevance     float64                `json:"relevance"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// RAGQueryResponse 表示RAG查询响应
type RAGQueryResponse struct {
	Answer  string           `json:"answer"`
	Sources []RAGQuerySource `json:"sources,omitempty"`
}

// Query 提交RAG查询请求
// @Summary 提交RAG查询请求
// @Description 使用RAG系统查询信息
// @Tags rag
// @Accept json
// @Produce json
// @Param request body QueryRequest true "查询请求"
// @Success 200 {object} QueryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/rag/query [post]
func (h *RAGHandler) Query(c *gin.Context) {
	var req RAGQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的请求参数: " + err.Error()})
		return
	}

	// 设置默认值
	if req.TopK <= 0 {
		req.TopK = 3
	}

	// 执行查询
	answer, sources, err := h.aiService.QueryWithRAG(c.Request.Context(), req.Query, req.TopK, req.Filters, req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "RAG查询失败: " + err.Error()})
		return
	}

	// 构建响应
	var sourcesResponse []RAGQuerySource
	for _, src := range sources {
		sourcesResponse = append(sourcesResponse, RAGQuerySource{
			DocumentID:    src.DocumentID,
			DocumentTitle: src.DocumentTitle,
			Content:       src.Content,
			Relevance:     src.Relevance,
			Metadata:      src.Metadata,
		})
	}

	c.JSON(http.StatusOK, RAGQueryResponse{
		Answer:  answer,
		Sources: sourcesResponse,
	})
}

// RegisterRoutes 注册RAG处理器的路由
func (h *RAGHandler) RegisterRoutes(router gin.IRouter) {
	ragGroup := router.Group("/api/v1/rag")
	{
		ragGroup.POST("/documents", h.UploadDocument)
		ragGroup.GET("/documents", h.ListDocuments)
		ragGroup.GET("/documents/:id", h.GetDocument)
		ragGroup.DELETE("/documents/:id", h.DeleteDocument)
		ragGroup.POST("/query", h.Query)
	}
}
