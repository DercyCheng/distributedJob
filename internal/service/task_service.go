package service

import (
	"errors"
	"time"

	"github.com/distributedJob/internal/job"
	"github.com/distributedJob/internal/model/entity"
	"github.com/distributedJob/internal/store"
)

// 定义错误
var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrInvalidParameters = errors.New("invalid parameters")
)

// 定义任务类型常量
const (
	TaskTypeHTTP = "HTTP"
	TaskTypeGRPC = "GRPC"
)

// 定义任务状态常量
const (
	TaskStatusEnabled  int8 = 1
	TaskStatusDisabled int8 = 2
)

// TaskStatistics 任务统计信息
type TaskStatistics struct {
	TaskCount        int                // 任务总数
	SuccessRate      float64            // 任务成功率
	AvgExecutionTime float64            // 平均执行时间(毫秒)
	ExecutionStats   map[string]float64 // 执行统计，可包含不同类型任务的统计数据
}

// TaskService 任务服务接口
type TaskService interface {
	// 任务相关
	GetTaskList(departmentID int64, page, size int) ([]*entity.Task, int64, error)
	GetTaskByID(id int64) (*entity.Task, error)
	CreateHTTPTask(task *entity.Task) (int64, error)
	CreateGRPCTask(task *entity.Task) (int64, error)
	UpdateHTTPTask(task *entity.Task) error
	UpdateGRPCTask(task *entity.Task) error
	DeleteTask(id int64) error
	UpdateTaskStatus(id int64, status int8) error

	// 执行记录相关
	GetRecordList(year, month int, taskID, departmentID *int64, success *int8, page, size int) ([]*entity.Record, int64, error)
	GetRecordByID(id int64, year, month int) (*entity.Record, error)
	GetRecordStats(year, month int, taskID, departmentID *int64) (map[string]interface{}, error)
	GetRecordListByTimeRange(year, month int, taskID, departmentID *int64, success *int8, page, size int, startTime, endTime time.Time) ([]*entity.Record, int64, error)

	// 为RPC服务添加的方法
	GetTaskRecords(taskID int64, startTime, endTime time.Time, limit, offset int) ([]*entity.Record, int64, error)
	GetTaskStatistics(departmentID int64, startTime, endTime time.Time) (*TaskStatistics, error)
}

// taskService 任务服务实现
type taskService struct {
	taskRepo  store.TaskRepository
	scheduler *job.Scheduler
}

// NewTaskService 创建任务服务
func NewTaskService(taskRepo store.TaskRepository, scheduler *job.Scheduler) TaskService {
	return &taskService{
		taskRepo:  taskRepo,
		scheduler: scheduler,
	}
}

// GetTaskList 获取任务列表
func (s *taskService) GetTaskList(departmentID int64, page, size int) ([]*entity.Task, int64, error) {
	return s.taskRepo.GetTasksByDepartmentID(departmentID, page, size)
}

// GetTaskByID 获取任务详情
func (s *taskService) GetTaskByID(id int64) (*entity.Task, error) {
	task, err := s.taskRepo.GetTaskByID(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// CreateHTTPTask 创建HTTP任务
func (s *taskService) CreateHTTPTask(task *entity.Task) (int64, error) {
	// 设置任务类型为HTTP
	task.TaskType = TaskTypeHTTP

	// 校验任务参数
	if err := s.validateHTTPTask(task); err != nil {
		return 0, err
	}

	// 创建任务
	id, err := s.taskRepo.CreateTask(task)
	if err != nil {
		return 0, err
	}

	// 同步到调度器
	if task.Status == TaskStatusEnabled {
		task.ID = id
		s.scheduler.AddTask(task)
	}

	return id, nil
}

// CreateGRPCTask 创建GRPC任务
func (s *taskService) CreateGRPCTask(task *entity.Task) (int64, error) {
	// 设置任务类型为GRPC
	task.TaskType = TaskTypeGRPC

	// 校验任务参数
	if err := s.validateGRPCTask(task); err != nil {
		return 0, err
	}

	// 创建任务
	id, err := s.taskRepo.CreateTask(task)
	if err != nil {
		return 0, err
	}

	// 同步到调度器
	if task.Status == TaskStatusEnabled {
		task.ID = id
		s.scheduler.AddTask(task)
	}

	return id, nil
}

// UpdateHTTPTask 更新HTTP任务
func (s *taskService) UpdateHTTPTask(task *entity.Task) error {
	// 检查任务是否存在
	oldTask, err := s.taskRepo.GetTaskByID(task.ID)
	if err != nil {
		return err
	}
	if oldTask == nil {
		return ErrTaskNotFound
	}

	// 设置任务类型为HTTP
	task.TaskType = TaskTypeHTTP

	// 校验任务参数
	if err := s.validateHTTPTask(task); err != nil {
		return err
	}

	// 更新任务
	if err := s.taskRepo.UpdateTask(task); err != nil {
		return err
	}

	// 同步到调度器
	if task.Status == TaskStatusEnabled {
		s.scheduler.AddTask(task) // 简化为重新添加任务而不是更新
	} else {
		s.scheduler.RemoveTask(task.ID)
	}

	return nil
}

// UpdateGRPCTask 更新GRPC任务
func (s *taskService) UpdateGRPCTask(task *entity.Task) error {
	// 检查任务是否存在
	oldTask, err := s.taskRepo.GetTaskByID(task.ID)
	if err != nil {
		return err
	}
	if oldTask == nil {
		return ErrTaskNotFound
	}

	// 设置任务类型为GRPC
	task.TaskType = TaskTypeGRPC

	// 校验任务参数
	if err := s.validateGRPCTask(task); err != nil {
		return err
	}

	// 更新任务
	if err := s.taskRepo.UpdateTask(task); err != nil {
		return err
	}

	// 同步到调度器
	if task.Status == TaskStatusEnabled {
		s.scheduler.AddTask(task) // 简化为重新添加任务而不是更新
	} else {
		s.scheduler.RemoveTask(task.ID)
	}

	return nil
}

// DeleteTask 删除任务
func (s *taskService) DeleteTask(id int64) error {
	// 检查任务是否存在
	task, err := s.taskRepo.GetTaskByID(id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	// 删除任务
	if err := s.taskRepo.DeleteTask(id); err != nil {
		return err
	}

	// 同步到调度器
	s.scheduler.RemoveTask(id)

	return nil
}

// UpdateTaskStatus 更新任务状态
func (s *taskService) UpdateTaskStatus(id int64, status int8) error {
	// 检查任务是否存在
	task, err := s.taskRepo.GetTaskByID(id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	// 检查状态是否有效
	if status != TaskStatusEnabled && status != TaskStatusDisabled {
		return ErrInvalidParameters
	}

	// 更新任务状态
	if err := s.taskRepo.UpdateTaskStatus(id, status); err != nil {
		return err
	}

	// 同步到调度器
	if status == TaskStatusEnabled {
		s.scheduler.AddTask(task)
	} else {
		s.scheduler.RemoveTask(id)
	}

	return nil
}

// GetRecordList 获取执行记录列表
func (s *taskService) GetRecordList(year, month int, taskID, departmentID *int64, success *int8, page, size int) ([]*entity.Record, int64, error) {
	return s.taskRepo.GetRecords(year, month, taskID, departmentID, success, page, size)
}

// GetRecordByID 获取执行记录详情
func (s *taskService) GetRecordByID(id int64, year, month int) (*entity.Record, error) {
	record, err := s.taskRepo.GetRecordByID(id, year, month)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, errors.New("record not found")
	}
	return record, nil
}

// GetRecordStats 获取执行记录统计
func (s *taskService) GetRecordStats(year, month int, taskID, departmentID *int64) (map[string]interface{}, error) {
	return s.taskRepo.GetRecordStats(year, month, taskID, departmentID)
}

// GetRecordListByTimeRange gets records within a time range
func (s *taskService) GetRecordListByTimeRange(year, month int, taskID, departmentID *int64, success *int8, page, size int, startTime, endTime time.Time) ([]*entity.Record, int64, error) {
	return s.taskRepo.GetRecordsByTimeRange(year, month, taskID, departmentID, success, page, size, startTime, endTime)
}

// GetTaskRecords gets task execution records by time range for the gRPC service
func (s *taskService) GetTaskRecords(taskID int64, startTime, endTime time.Time, limit, offset int) ([]*entity.Record, int64, error) {
	// Determine the year and month for partitioning
	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())

	// Start with current month/year and query backwards if needed
	var allRecords []*entity.Record
	var total int64 = 0

	// Start by querying the current month/year
	taskIDPtr := &taskID
	records, count, err := s.GetRecordListByTimeRange(
		currentYear,
		currentMonth,
		taskIDPtr,
		nil,  // no department filter
		nil,  // no success filter
		1,    // page
		1000, // large size to get all records
		startTime,
		endTime,
	)

	if err != nil {
		return nil, 0, err
	}

	allRecords = append(allRecords, records...)
	total += count

	// If the start time is from a previous month, we need to query those months too
	startYear := startTime.Year()
	startMonth := int(startTime.Month())

	// Query previous months if needed
	for (currentYear > startYear) || (currentYear == startYear && currentMonth > startMonth) {
		// Move to previous month
		currentMonth--
		if currentMonth == 0 {
			currentMonth = 12
			currentYear--
		}

		// Don't go beyond the start date
		if currentYear < startYear || (currentYear == startYear && currentMonth < startMonth) {
			break
		}

		// Query this month
		records, count, err := s.GetRecordListByTimeRange(
			currentYear,
			currentMonth,
			taskIDPtr,
			nil,  // no department filter
			nil,  // no success filter
			1,    // page
			1000, // large size to get all records
			startTime,
			endTime,
		)

		if err != nil {
			// Log error but continue with what we have
			continue
		}

		allRecords = append(allRecords, records...)
		total += count
	}

	// Apply pagination to the aggregated results
	start := offset
	end := offset + limit
	if start >= len(allRecords) {
		return []*entity.Record{}, total, nil
	}
	if end > len(allRecords) {
		end = len(allRecords)
	}

	return allRecords[start:end], total, nil
}

// GetTaskStatistics gets task statistics for the specified time range and department
func (s *taskService) GetTaskStatistics(departmentID int64, startTime, endTime time.Time) (*TaskStatistics, error) {
	// Determine the year and month for partitioning
	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())

	// Initialize statistics
	stats := &TaskStatistics{
		TaskCount:        0,
		SuccessRate:      0,
		AvgExecutionTime: 0,
		ExecutionStats:   make(map[string]float64),
	}

	// Get department tasks count
	deptIDPtr := &departmentID
	_, count, err := s.GetTaskList(departmentID, 1, 1000)
	if err != nil {
		return nil, err
	}

	// Set task count
	stats.TaskCount = int(count)

	// Get record statistics for the current month
	statsMap, err := s.GetRecordStats(currentYear, currentMonth, nil, deptIDPtr)
	if err != nil {
		return nil, err
	}

	// Extract success rate
	if successRate, ok := statsMap["success_rate"].(float64); ok {
		stats.SuccessRate = successRate
	}

	// Extract average execution time
	if avgExecTime, ok := statsMap["avg_execution_time"].(float64); ok {
		stats.AvgExecutionTime = avgExecTime
	}

	// Extract type-specific statistics
	if httpSuccessRate, ok := statsMap["http_success_rate"].(float64); ok {
		stats.ExecutionStats["http_success_rate"] = httpSuccessRate
	}

	if grpcSuccessRate, ok := statsMap["grpc_success_rate"].(float64); ok {
		stats.ExecutionStats["grpc_success_rate"] = grpcSuccessRate
	}

	if httpAvgTime, ok := statsMap["http_avg_time"].(float64); ok {
		stats.ExecutionStats["http_avg_time"] = httpAvgTime
	}

	if grpcAvgTime, ok := statsMap["grpc_avg_time"].(float64); ok {
		stats.ExecutionStats["grpc_avg_time"] = grpcAvgTime
	}

	if totalTasks, ok := statsMap["total_tasks"].(int); ok {
		stats.ExecutionStats["total_tasks"] = float64(totalTasks)
	}

	if totalSuccess, ok := statsMap["total_success"].(int); ok {
		stats.ExecutionStats["total_success"] = float64(totalSuccess)
	}

	if totalFailed, ok := statsMap["total_failed"].(int); ok {
		stats.ExecutionStats["total_failed"] = float64(totalFailed)
	}

	return stats, nil
}

// validateHTTPTask 验证HTTP任务
func (s *taskService) validateHTTPTask(task *entity.Task) error {
	if task.Name == "" {
		return errors.New("task name cannot be empty")
	}

	if task.Cron == "" {
		return errors.New("cron expression cannot be empty")
	}

	if task.URL == "" {
		return errors.New("HTTP URL cannot be empty")
	}

	if task.HTTPMethod == "" {
		return errors.New("HTTP method cannot be empty")
	}

	return nil
}

// validateGRPCTask 验证GRPC任务
func (s *taskService) validateGRPCTask(task *entity.Task) error {
	if task.Name == "" {
		return errors.New("task name cannot be empty")
	}

	if task.Cron == "" {
		return errors.New("cron expression cannot be empty")
	}

	if task.GrpcService == "" {
		return errors.New("GRPC service cannot be empty")
	}

	if task.GrpcMethod == "" {
		return errors.New("GRPC method cannot be empty")
	}

	return nil
}
