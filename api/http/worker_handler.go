package http

import (
	"net/http"
	"strconv"

	"go-job/internal/models"
	"go-job/pkg/database"
	"go-job/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WorkerHandler 工作节点处理器
type WorkerHandler struct {
	db *gorm.DB
}

// NewWorkerHandler 创建工作节点处理器
func NewWorkerHandler() *WorkerHandler {
	return &WorkerHandler{
		db: database.GetDB(),
	}
}

// ListWorkers 获取工作节点列表
func (h *WorkerHandler) ListWorkers(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	size, _ := strconv.ParseInt(c.DefaultQuery("size", "10"), 10, 32)
	status := c.Query("status")

	query := h.db.Model(&models.Worker{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.WithError(err).Error("查询工作节点总数失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 分页查询
	var workers []models.Worker
	offset := (page - 1) * size
	if err := query.Offset(int(offset)).Limit(int(size)).Order("created_at DESC").Find(&workers).Error; err != nil {
		logger.WithError(err).Error("查询工作节点失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"workers": workers,
			"total":   total,
			"page":    page,
			"size":    size,
		},
	})
}

// GetWorker 获取工作节点详情
func (h *WorkerHandler) GetWorker(c *gin.Context) {
	id := c.Param("id")

	var worker models.Worker
	if err := h.db.First(&worker, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "工作节点不存在"})
			return
		}
		logger.WithError(err).Error("查询工作节点失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": worker})
}

// UpdateWorkerStatusRequest 更新工作节点状态请求
type UpdateWorkerStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdateWorkerStatus 更新工作节点状态
func (h *WorkerHandler) UpdateWorkerStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateWorkerStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证状态值
	validStatuses := map[string]bool{
		"online":      true,
		"offline":     true,
		"busy":        true,
		"maintenance": true,
	}

	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态值"})
		return
	}

	var worker models.Worker
	if err := h.db.First(&worker, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "工作节点不存在"})
			return
		}
		logger.WithError(err).Error("查询工作节点失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 更新状态
	if err := h.db.Model(&worker).Update("status", req.Status).Error; err != nil {
		logger.WithError(err).Error("更新工作节点状态失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "工作节点状态更新成功",
		"data": gin.H{
			"id":     id,
			"status": req.Status,
		},
	})
}
