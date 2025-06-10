package http

import (
	"net/http"
	"strconv"

	"go-job/api/grpc"
	"go-job/internal/job"
	"go-job/pkg/logger"

	"github.com/gin-gonic/gin"
)

// JobHandler 任务处理器
type JobHandler struct {
	jobService *job.Service
}

// NewJobHandler 创建任务处理器
func NewJobHandler() *JobHandler {
	return &JobHandler{
		jobService: job.NewService(),
	}
}

// CreateJobRequest 创建任务请求
type CreateJobRequest struct {
	Name          string            `json:"name" binding:"required"`
	Description   string            `json:"description"`
	Cron          string            `json:"cron" binding:"required"`
	Command       string            `json:"command" binding:"required"`
	Params        map[string]string `json:"params"`
	RetryAttempts int32             `json:"retry_attempts"`
	Timeout       int32             `json:"timeout"`
}

// UpdateJobRequest 更新任务请求
type UpdateJobRequest struct {
	Name          string            `json:"name" binding:"required"`
	Description   string            `json:"description"`
	Cron          string            `json:"cron" binding:"required"`
	Command       string            `json:"command" binding:"required"`
	Params        map[string]string `json:"params"`
	Enabled       bool              `json:"enabled"`
	RetryAttempts int32             `json:"retry_attempts"`
	Timeout       int32             `json:"timeout"`
}

// CreateJob 创建任务
func (h *JobHandler) CreateJob(c *gin.Context) {
	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.CreateJobRequest{
		Name:          req.Name,
		Description:   req.Description,
		Cron:          req.Cron,
		Command:       req.Command,
		Params:        req.Params,
		RetryAttempts: req.RetryAttempts,
		Timeout:       req.Timeout,
	}

	resp, err := h.jobService.CreateJob(c.Request.Context(), grpcReq)
	if err != nil {
		logger.WithError(err).Error("创建任务失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": resp.Job})
}

// GetJob 获取任务
func (h *JobHandler) GetJob(c *gin.Context) {
	id := c.Param("id")

	grpcReq := &grpc.GetJobRequest{Id: id}
	resp, err := h.jobService.GetJob(c.Request.Context(), grpcReq)
	if err != nil {
		logger.WithError(err).Error("获取任务失败")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp.Job})
}

// ListJobs 获取任务列表
func (h *JobHandler) ListJobs(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	size, _ := strconv.ParseInt(c.DefaultQuery("size", "10"), 10, 32)
	keyword := c.Query("keyword")
	enabled, _ := strconv.ParseBool(c.DefaultQuery("enabled", "false"))

	grpcReq := &grpc.ListJobsRequest{
		Page:    int32(page),
		Size:    int32(size),
		Keyword: keyword,
		Enabled: enabled,
	}

	resp, err := h.jobService.ListJobs(c.Request.Context(), grpcReq)
	if err != nil {
		logger.WithError(err).Error("获取任务列表失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"jobs":  resp.Jobs,
			"total": resp.Total,
			"page":  page,
			"size":  size,
		},
	})
}

// UpdateJob 更新任务
func (h *JobHandler) UpdateJob(c *gin.Context) {
	id := c.Param("id")

	var req UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.UpdateJobRequest{
		Id:            id,
		Name:          req.Name,
		Description:   req.Description,
		Cron:          req.Cron,
		Command:       req.Command,
		Params:        req.Params,
		Enabled:       req.Enabled,
		RetryAttempts: req.RetryAttempts,
		Timeout:       req.Timeout,
	}

	resp, err := h.jobService.UpdateJob(c.Request.Context(), grpcReq)
	if err != nil {
		logger.WithError(err).Error("更新任务失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp.Job})
}

// DeleteJob 删除任务
func (h *JobHandler) DeleteJob(c *gin.Context) {
	id := c.Param("id")

	grpcReq := &grpc.DeleteJobRequest{Id: id}
	_, err := h.jobService.DeleteJob(c.Request.Context(), grpcReq)
	if err != nil {
		logger.WithError(err).Error("删除任务失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务删除成功"})
}

// TriggerJob 手动触发任务
func (h *JobHandler) TriggerJob(c *gin.Context) {
	id := c.Param("id")

	var params map[string]string
	c.ShouldBindJSON(&params)

	grpcReq := &grpc.TriggerJobRequest{
		Id:     id,
		Params: params,
	}

	resp, err := h.jobService.TriggerJob(c.Request.Context(), grpcReq)
	if err != nil {
		logger.WithError(err).Error("触发任务失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "任务触发成功",
		"execution_id": resp.ExecutionId,
	})
}

// GetJobExecutions 获取任务执行记录
func (h *JobHandler) GetJobExecutions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetJobExecutions - to be implemented"})
}
