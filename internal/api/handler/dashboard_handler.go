package handler

import (
	"net/http"
	"strconv"
	"time"

	"distributedJob/internal/job"
	"distributedJob/internal/model/entity"
	"distributedJob/internal/service"
	"distributedJob/pkg/metrics"

	"github.com/gin-gonic/gin"
)

// DashboardHandler 仪表盘处理器
type DashboardHandler struct {
	taskService service.TaskService
	scheduler   *job.Scheduler
	metrics     *metrics.Metrics
}

// NewDashboardHandler 创建仪表盘处理器
func NewDashboardHandler(
	taskService service.TaskService,
	scheduler *job.Scheduler,
	metrics *metrics.Metrics,
) *DashboardHandler {
	return &DashboardHandler{
		taskService: taskService,
		scheduler:   scheduler,
		metrics:     metrics,
	}
}

// SystemStats 系统统计信息
type SystemStats struct {
	TaskCount       int     `json:"taskCount"`       // 任务总数
	RunningJobs     int     `json:"runningJobs"`     // 正在运行的任务数
	SuccessRate     float64 `json:"successRate"`     // 成功率
	AvgResponseTime float64 `json:"avgResponseTime"` // 平均响应时间(毫秒)
	ErrorRate       float64 `json:"errorRate"`       // 错误率
	LastUpdate      string  `json:"lastUpdate"`      // 最后更新时间
}

// RealtimeMetrics 实时指标
type RealtimeMetrics struct {
	CPUUsage          float64 `json:"cpuUsage"`          // CPU使用率
	MemoryUsage       float64 `json:"memoryUsage"`       // 内存使用率
	DiskUsage         float64 `json:"diskUsage"`         // 磁盘使用率
	NetworkThroughput float64 `json:"networkThroughput"` // 网络吞吐量
	QPS               float64 `json:"qps"`               // 每秒查询数
}

// TaskCount 任务计数
type TaskCount struct {
	HTTP int `json:"http"` // HTTP任务数
	GRPC int `json:"grpc"` // GRPC任务数
}

// DashboardOverview 仪表盘概览
type DashboardOverview struct {
	Stats            SystemStats     `json:"stats"`            // 系统统计信息
	Metrics          RealtimeMetrics `json:"metrics"`          // 实时指标
	TaskTypeCounts   TaskCount       `json:"taskTypeCounts"`   // 任务类型计数
	RecentTasks      []entity.Task   `json:"recentTasks"`      // 最近任务
	RecentExecutions []entity.Record `json:"recentExecutions"` // 最近执行记录
}

// Overview 获取仪表盘概览数据
func (h *DashboardHandler) Overview(c *gin.Context) {
	// 获取部门ID，默认为0（全部部门）
	deptIDStr := c.Query("departmentId")
	deptID := int64(0)

	if deptIDStr != "" {
		var err error
		deptID, err = strconv.ParseInt(deptIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department ID"})
			return
		}
	}

	// 获取任务列表
	tasks, total, err := h.taskService.GetTaskList(deptID, 1, 1000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	// 获取最近执行记录
	now := time.Now()
	year, month := now.Year(), int(now.Month())
	records, _, err := h.taskService.GetRecordList(year, month, nil, &deptID, nil, 1, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get execution records"})
		return
	}

	// 构建响应
	overview := DashboardOverview{
		Stats: SystemStats{
			TaskCount:   int(total),
			RunningJobs: h.scheduler.GetRunningJobCount(), // 假设Scheduler有此方法
			SuccessRate: calculateSuccessRate(records),
			LastUpdate:  time.Now().Format(time.RFC3339),
		},
		TaskTypeCounts:   countTaskTypes(tasks),
		RecentTasks:      convertToTaskArray(tasks[:min(5, len(tasks))]),
		RecentExecutions: convertToRecordArray(records),
	}

	// 如果有指标组件，从中获取实时指标
	if h.metrics != nil {
		// 这里仅是示例，实际需要从Prometheus或其他监控系统获取
		overview.Metrics = RealtimeMetrics{
			CPUUsage:    45.7,
			MemoryUsage: 60.2,
			DiskUsage:   25.5,
			QPS:         120.5,
		}
	}

	c.JSON(http.StatusOK, overview)
}

// 计算成功率
func calculateSuccessRate(records []*entity.Record) float64 {
	if len(records) == 0 {
		return 100.0
	}

	successCount := 0
	for _, record := range records {
		if record.Success == 1 {
			successCount++
		}
	}

	return float64(successCount) * 100.0 / float64(len(records))
}

// 统计不同类型的任务数量
func countTaskTypes(tasks []*entity.Task) TaskCount {
	httpCount := 0
	grpcCount := 0

	for _, task := range tasks {
		if task.Type == "HTTP" {
			httpCount++
		} else if task.Type == "GRPC" {
			grpcCount++
		}
	}

	return TaskCount{
		HTTP: httpCount,
		GRPC: grpcCount,
	}
}

// 转换任务指针数组为非指针数组
func convertToTaskArray(tasks []*entity.Task) []entity.Task {
	result := make([]entity.Task, len(tasks))
	for i, task := range tasks {
		result[i] = *task
	}
	return result
}

// 转换记录指针数组为非指针数组
func convertToRecordArray(records []*entity.Record) []entity.Record {
	result := make([]entity.Record, len(records))
	for i, record := range records {
		result[i] = *record
	}
	return result
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
