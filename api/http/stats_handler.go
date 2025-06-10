package http

import (
	"net/http"
	"time"

	"go-job/internal/models"
	"go-job/pkg/database"
	"go-job/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StatsHandler 统计处理器
type StatsHandler struct {
	db *gorm.DB
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler() *StatsHandler {
	return &StatsHandler{
		db: database.GetDB(),
	}
}

// DashboardStats 仪表板统计数据
type DashboardStats struct {
	TotalJobs        int64                  `json:"total_jobs"`
	ActiveJobs       int64                  `json:"active_jobs"`
	TotalWorkers     int64                  `json:"total_workers"`
	OnlineWorkers    int64                  `json:"online_workers"`
	TotalExecutions  int64                  `json:"total_executions"`
	TodayExecutions  int64                  `json:"today_executions"`
	SuccessRate      float64                `json:"success_rate"`
	RecentJobs       []models.Job           `json:"recent_jobs"`
	RecentExecutions []models.JobExecution  `json:"recent_executions"`
	ExecutionStats   []ExecutionStatsByDate `json:"execution_stats"`
}

// ExecutionStatsByDate 按日期统计执行数据
type ExecutionStatsByDate struct {
	Date    string `json:"date"`
	Success int64  `json:"success"`
	Failed  int64  `json:"failed"`
	Total   int64  `json:"total"`
}

// JobStats 任务统计数据
type JobStats struct {
	JobID           string                 `json:"job_id"`
	JobName         string                 `json:"job_name"`
	TotalExecutions int64                  `json:"total_executions"`
	SuccessCount    int64                  `json:"success_count"`
	FailedCount     int64                  `json:"failed_count"`
	SuccessRate     float64                `json:"success_rate"`
	AvgDuration     float64                `json:"avg_duration"`
	LastExecution   *models.JobExecution   `json:"last_execution"`
	ExecutionStats  []ExecutionStatsByDate `json:"execution_stats"`
}

// GetDashboard 获取仪表板数据
func (h *StatsHandler) GetDashboard(c *gin.Context) {
	stats := DashboardStats{}

	// 总任务数
	h.db.Model(&models.Job{}).Count(&stats.TotalJobs)

	// 活跃任务数
	h.db.Model(&models.Job{}).Where("enabled = ?", true).Count(&stats.ActiveJobs)

	// 总工作节点数
	h.db.Model(&models.Worker{}).Count(&stats.TotalWorkers)

	// 在线工作节点数
	h.db.Model(&models.Worker{}).Where("status = ?", models.WorkerStatusOnline).Count(&stats.OnlineWorkers)

	// 总执行数
	h.db.Model(&models.JobExecution{}).Count(&stats.TotalExecutions)

	// 今日执行数
	today := time.Now().Format("2006-01-02")
	h.db.Model(&models.JobExecution{}).Where("DATE(created_at) = ?", today).Count(&stats.TodayExecutions)

	// 成功率
	var successCount int64
	h.db.Model(&models.JobExecution{}).Where("status = ?", models.ExecutionStatusSuccess).Count(&successCount)
	if stats.TotalExecutions > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalExecutions) * 100
	}

	// 最近的任务
	h.db.Order("created_at DESC").Limit(5).Find(&stats.RecentJobs)

	// 最近的执行记录
	h.db.Preload("Job").Order("created_at DESC").Limit(10).Find(&stats.RecentExecutions)

	// 最近7天的执行统计
	stats.ExecutionStats = h.getExecutionStatsByDays(7)

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetJobStats 获取任务统计
func (h *StatsHandler) GetJobStats(c *gin.Context) {
	jobID := c.Param("id")

	var job models.Job
	if err := h.db.First(&job, "id = ?", jobID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
			return
		}
		logger.WithError(err).Error("查询任务失败")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats := JobStats{
		JobID:   job.ID,
		JobName: job.Name,
	}

	// 总执行数
	h.db.Model(&models.JobExecution{}).Where("job_id = ?", jobID).Count(&stats.TotalExecutions)

	// 成功数
	h.db.Model(&models.JobExecution{}).Where("job_id = ? AND status = ?", jobID, models.ExecutionStatusSuccess).Count(&stats.SuccessCount)

	// 失败数
	h.db.Model(&models.JobExecution{}).Where("job_id = ? AND status = ?", jobID, models.ExecutionStatusFailed).Count(&stats.FailedCount)

	// 成功率
	if stats.TotalExecutions > 0 {
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalExecutions) * 100
	}

	// 平均执行时长
	type DurationResult struct {
		AvgDuration float64 `json:"avg_duration"`
	}
	var result DurationResult
	h.db.Model(&models.JobExecution{}).
		Select("AVG(TIMESTAMPDIFF(SECOND, started_at, finished_at)) as avg_duration").
		Where("job_id = ? AND started_at IS NOT NULL AND finished_at IS NOT NULL", jobID).
		Scan(&result)
	stats.AvgDuration = result.AvgDuration

	// 最后一次执行
	var lastExecution models.JobExecution
	if err := h.db.Where("job_id = ?", jobID).Order("created_at DESC").First(&lastExecution).Error; err == nil {
		stats.LastExecution = &lastExecution
	}

	// 最近30天的执行统计
	stats.ExecutionStats = h.getJobExecutionStatsByDays(jobID, 30)

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetWorkerStats 获取工作节点统计
func (h *StatsHandler) GetWorkerStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetWorkerStats - to be implemented"})
}

// GetExecutionStats 获取执行统计
func (h *StatsHandler) GetExecutionStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetExecutionStats - to be implemented"})
}

// getExecutionStatsByDays 获取最近N天的执行统计
func (h *StatsHandler) getExecutionStatsByDays(days int) []ExecutionStatsByDate {
	var stats []ExecutionStatsByDate

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		var total, success, failed int64
		h.db.Model(&models.JobExecution{}).Where("DATE(created_at) = ?", date).Count(&total)
		h.db.Model(&models.JobExecution{}).Where("DATE(created_at) = ? AND status = ?", date, models.ExecutionStatusSuccess).Count(&success)
		h.db.Model(&models.JobExecution{}).Where("DATE(created_at) = ? AND status = ?", date, models.ExecutionStatusFailed).Count(&failed)

		stats = append(stats, ExecutionStatsByDate{
			Date:    date,
			Success: success,
			Failed:  failed,
			Total:   total,
		})
	}

	return stats
}

// getJobExecutionStatsByDays 获取特定任务最近N天的执行统计
func (h *StatsHandler) getJobExecutionStatsByDays(jobID string, days int) []ExecutionStatsByDate {
	var stats []ExecutionStatsByDate

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		var total, success, failed int64
		h.db.Model(&models.JobExecution{}).Where("job_id = ? AND DATE(created_at) = ?", jobID, date).Count(&total)
		h.db.Model(&models.JobExecution{}).Where("job_id = ? AND DATE(created_at) = ? AND status = ?", jobID, date, models.ExecutionStatusSuccess).Count(&success)
		h.db.Model(&models.JobExecution{}).Where("job_id = ? AND DATE(created_at) = ? AND status = ?", jobID, date, models.ExecutionStatusFailed).Count(&failed)

		stats = append(stats, ExecutionStatsByDate{
			Date:    date,
			Success: success,
			Failed:  failed,
			Total:   total,
		})
	}

	return stats
}
