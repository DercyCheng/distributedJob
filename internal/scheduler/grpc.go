package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"go-job/api/grpc"
	"go-job/internal/models"
	"go-job/pkg/logger"
	"time"

	"github.com/google/uuid"
)

// RegisterWorker 注册工作节点
func (s *Service) RegisterWorker(ctx context.Context, req *grpc.RegisterWorkerRequest) (*grpc.RegisterWorkerResponse, error) {
	logger.Infof("注册工作节点: %s", req.GetName())

	workerID := uuid.New().String()

	// 创建工作节点记录
	metadataJSON, _ := json.Marshal(req.GetMetadata())
	worker := &models.Worker{
		ID:          workerID,
		Name:        req.GetName(),
		IP:          req.GetIp(),
		Port:        int(req.GetPort()),
		Status:      models.WorkerStatusOnline,
		Capacity:    int(req.GetCapacity()),
		CurrentLoad: 0,
		Metadata:    string(metadataJSON),
	}

	if err := s.db.Create(worker).Error; err != nil {
		logger.WithError(err).Error("创建工作节点记录失败")
		return nil, fmt.Errorf("注册工作节点失败: %w", err)
	}

	// 添加到内存中
	s.workersMu.Lock()
	s.workers[workerID] = &WorkerInfo{
		ID:          workerID,
		Name:        req.GetName(),
		IP:          req.GetIp(),
		Port:        req.GetPort(),
		Status:      grpc.WorkerStatus_ONLINE,
		Capacity:    req.GetCapacity(),
		CurrentLoad: 0,
		LastSeen:    time.Now(),
		Metadata:    req.GetMetadata(),
	}
	s.workersMu.Unlock()

	logger.Infof("工作节点注册成功: %s (ID: %s)", req.GetName(), workerID)

	return &grpc.RegisterWorkerResponse{
		WorkerId: workerID,
	}, nil
}

// Heartbeat 心跳
func (s *Service) Heartbeat(ctx context.Context, req *grpc.HeartbeatRequest) (*grpc.HeartbeatResponse, error) {
	workerID := req.GetWorkerId()

	s.workersMu.Lock()
	worker, exists := s.workers[workerID]
	if !exists {
		s.workersMu.Unlock()
		return &grpc.HeartbeatResponse{Success: false}, nil
	}

	// 更新心跳时间和状态
	worker.LastSeen = time.Now()
	worker.CurrentLoad = req.GetCurrentLoad()
	worker.Status = req.GetStatus()
	s.workersMu.Unlock()

	// 更新数据库
	updates := map[string]interface{}{
		"last_heartbeat": time.Now(),
		"current_load":   req.GetCurrentLoad(),
		"status":         convertWorkerStatus(req.GetStatus()),
	}

	if err := s.db.Model(&models.Worker{}).Where("id = ?", workerID).Updates(updates).Error; err != nil {
		logger.WithError(err).Errorf("更新工作节点心跳失败: %s", workerID)
	}

	return &grpc.HeartbeatResponse{Success: true}, nil
}

// GetTask 获取任务
func (s *Service) GetTask(ctx context.Context, req *grpc.GetTaskRequest) (*grpc.GetTaskResponse, error) {
	workerID := req.GetWorkerId()
	capacity := req.GetCapacity()

	// 查找分配给该工作节点的待执行任务
	var schedules []models.JobSchedule
	err := s.db.Preload("Job").
		Where("worker_id = ? AND status = ?", workerID, models.ScheduleStatusAssigned).
		Limit(int(capacity)).
		Find(&schedules).Error

	if err != nil {
		logger.WithError(err).Errorf("查询工作节点任务失败: %s", workerID)
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	var tasks []*grpc.Task
	for _, schedule := range schedules {
		if schedule.Job.ID == "" {
			continue
		}

		// 解析任务参数
		var params map[string]string
		if schedule.Job.Params != "" {
			json.Unmarshal([]byte(schedule.Job.Params), &params)
		}

		task := &grpc.Task{
			Id:            schedule.ExecutionID,
			JobId:         schedule.JobID,
			Command:       schedule.Job.Command,
			Params:        params,
			Timeout:       int32(schedule.Job.Timeout),
			RetryAttempts: int32(schedule.Job.RetryAttempts),
		}

		tasks = append(tasks, task)

		// 更新调度状态为执行中
		schedule.Status = models.ScheduleStatusExecuting
		schedule.ExecutedAt = &time.Time{}
		*schedule.ExecutedAt = time.Now()
		s.db.Save(&schedule)
	}

	logger.Infof("为工作节点 %s 分配了 %d 个任务", workerID, len(tasks))

	return &grpc.GetTaskResponse{
		Tasks: tasks,
	}, nil
}

// ReportTaskResult 报告任务结果
func (s *Service) ReportTaskResult(ctx context.Context, req *grpc.ReportTaskResultRequest) (*grpc.ReportTaskResultResponse, error) {
	executionID := req.GetTaskId()
	workerID := req.GetWorkerId()

	logger.Infof("收到任务结果报告: %s", executionID)

	// 更新执行记录
	updates := map[string]interface{}{
		"status":    convertExecutionStatus(req.GetStatus()),
		"output":    req.GetOutput(),
		"error":     req.GetError(),
		"exit_code": req.GetExitCode(),
	}

	if req.GetStartedAt() != nil {
		startedAt := req.GetStartedAt().AsTime()
		updates["started_at"] = &startedAt
	}

	if req.GetFinishedAt() != nil {
		finishedAt := req.GetFinishedAt().AsTime()
		updates["finished_at"] = &finishedAt
	}

	if err := s.db.Model(&models.JobExecution{}).Where("id = ?", executionID).Updates(updates).Error; err != nil {
		logger.WithError(err).Errorf("更新执行记录失败: %s", executionID)
		return &grpc.ReportTaskResultResponse{Success: false}, nil
	}

	// 更新调度记录状态
	var scheduleStatus models.ScheduleStatus
	switch req.GetStatus() {
	case grpc.ExecutionStatus_SUCCESS:
		scheduleStatus = models.ScheduleStatusCompleted
	case grpc.ExecutionStatus_FAILED, grpc.ExecutionStatus_TIMEOUT, grpc.ExecutionStatus_CANCELLED:
		scheduleStatus = models.ScheduleStatusFailed
	default:
		scheduleStatus = models.ScheduleStatusExecuting
	}

	s.db.Model(&models.JobSchedule{}).
		Where("execution_id = ?", executionID).
		Update("status", scheduleStatus)

	// 减少工作节点负载
	s.workersMu.Lock()
	if worker, exists := s.workers[workerID]; exists && worker.CurrentLoad > 0 {
		worker.CurrentLoad--
	}
	s.workersMu.Unlock()

	// 如果任务失败且需要重试，创建重试任务
	if req.GetStatus() == grpc.ExecutionStatus_FAILED {
		s.handleTaskRetry(executionID)
	}

	logger.Infof("任务结果处理完成: %s", executionID)

	return &grpc.ReportTaskResultResponse{Success: true}, nil
}

// handleTaskRetry 处理任务重试
func (s *Service) handleTaskRetry(executionID string) {
	var execution models.JobExecution
	if err := s.db.Preload("Job").First(&execution, "id = ?", executionID).Error; err != nil {
		logger.WithError(err).Errorf("查询执行记录失败: %s", executionID)
		return
	}

	// 检查是否还有重试次数
	var retryCount int64
	s.db.Model(&models.JobExecution{}).
		Where("job_id = ? AND status = ?", execution.JobID, models.ExecutionStatusFailed).
		Count(&retryCount)

	if retryCount < int64(execution.Job.RetryAttempts) {
		logger.Infof("创建重试任务: %s (第%d次重试)", execution.JobID, retryCount+1)

		// 创建新的调度记录
		schedule := &models.JobSchedule{
			ID:          uuid.New().String(),
			JobID:       execution.JobID,
			ScheduledAt: time.Now().Add(30 * time.Second), // 30秒后重试
			Status:      models.ScheduleStatusPending,
		}

		if err := s.db.Create(schedule).Error; err != nil {
			logger.WithError(err).Errorf("创建重试调度记录失败: %s", execution.JobID)
			return
		}

		// 延迟加入队列
		time.AfterFunc(30*time.Second, func() {
			select {
			case s.taskQueue <- schedule:
				logger.Debugf("重试任务已加入队列: %s", execution.JobID)
			default:
				logger.Warnf("任务队列已满，重试任务被丢弃: %s", execution.JobID)
			}
		})
	}
}

// 状态转换函数
func convertWorkerStatus(status grpc.WorkerStatus) models.WorkerStatus {
	switch status {
	case grpc.WorkerStatus_ONLINE:
		return models.WorkerStatusOnline
	case grpc.WorkerStatus_BUSY:
		return models.WorkerStatusBusy
	case grpc.WorkerStatus_MAINTENANCE:
		return models.WorkerStatusMaintenance
	default:
		return models.WorkerStatusOffline
	}
}

func convertExecutionStatus(status grpc.ExecutionStatus) models.JobExecutionStatus {
	switch status {
	case grpc.ExecutionStatus_RUNNING:
		return models.ExecutionStatusRunning
	case grpc.ExecutionStatus_SUCCESS:
		return models.ExecutionStatusSuccess
	case grpc.ExecutionStatus_FAILED:
		return models.ExecutionStatusFailed
	case grpc.ExecutionStatus_TIMEOUT:
		return models.ExecutionStatusTimeout
	case grpc.ExecutionStatus_CANCELLED:
		return models.ExecutionStatusCancelled
	default:
		return models.ExecutionStatusPending
	}
}
