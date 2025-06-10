package http

import (
	"net/http"
	"strconv"
	"time"

	"go-job/internal/models"
	"go-job/pkg/database"
	"go-job/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ExecutionHandler 执行记录处理器
type ExecutionHandler struct {
	db *gorm.DB
}

// NewExecutionHandler 创建执行记录处理器
func NewExecutionHandler() *ExecutionHandler {
	return &ExecutionHandler{
		db: database.GetDB(),
	}
}

// ListExecutions 获取执行记录列表
func (h *ExecutionHandler) ListExecutions(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	size, _ := strconv.ParseInt(c.DefaultQuery("size", "10"), 10, 32)
	jobID := c.Query("job_id")
	status := c.Query("status")

	query := h.db.Model(&models.JobExecution{}).Preload("Job").Preload("Worker")

	if jobID != "" {
		query = query.Where("job_id = ?", jobID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.WithError(err).Error("查询执行记录总数失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 分页查询
	var executions []models.JobExecution
	offset := (page - 1) * size
	if err := query.Offset(int(offset)).Limit(int(size)).Order("created_at DESC").Find(&executions).Error; err != nil {
		logger.WithError(err).Error("查询执行记录失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"executions": executions,
			"total":      total,
			"page":       page,
			"size":       size,
		},
	})
}

// GetExecution 获取执行记录详情
func (h *ExecutionHandler) GetExecution(c *gin.Context) {
	id := c.Param("id")

	var execution models.JobExecution
	if err := h.db.Preload("Job").Preload("Worker").First(&execution, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "执行记录不存在"})
			return
		}
		logger.WithError(err).Error("查询执行记录失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": execution})
}

// CancelExecution 取消执行记录
func (h *ExecutionHandler) CancelExecution(c *gin.Context) {
	id := c.Param("id")

	var execution models.JobExecution
	if err := h.db.First(&execution, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "执行记录不存在"})
			return
		}
		logger.WithError(err).Error("查询执行记录失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查是否可以取消
	if execution.Status != models.ExecutionStatusPending && execution.Status != models.ExecutionStatusRunning {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能取消等待中或运行中的任务"})
		return
	}

	// 更新状态为已取消
	if err := h.db.Model(&execution).Updates(map[string]interface{}{
		"status":      models.ExecutionStatusCancelled,
		"finished_at": time.Now(),
		"error":       "用户取消执行",
	}).Error; err != nil {
		logger.WithError(err).Error("取消执行记录失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "执行记录已取消",
		"data":    gin.H{"id": id, "status": "cancelled"},
	})
}
