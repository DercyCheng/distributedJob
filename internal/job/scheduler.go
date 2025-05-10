package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"distributedJob/internal/config"
	"distributedJobodel/entity"
	"distributedJob"
	"github.com/robfig/cron/v3"
)

// Scheduler 管理定时任务的调度器
type Scheduler struct {
	cron        *cron.Cron
	config      *config.Config
	taskStore   TaskRepository
	jobs        map[int64]cron.EntryID
	jobsMutex   sync.RWMutex
	httpWorker  *HTTPWorker
	grpcWorker  *GRPCWorker
	isRunning   bool
	runningJobs chan *JobContext
	ctx         context.Context
	cancel      context.CancelFunc
}

// JobContext 表示一个任务执行上下文
type JobContext struct {
	Task       *entity.Task
	ExecutedAt time.Time
	Done       chan *JobResult
}

// JobResult 表示一个任务执行结果
type JobResult struct {
	Success      bool
	StatusCode   *int
	GrpcStatus   *int
	Response     string
	Error        error
	RetryTimes   int
	UseFallback  bool
	CostTime     int
	ActualURL    string // 实际执行的URL（可能是主URL或备用URL）
	ActualMethod string // 实际使用的HTTP方法
}

// TaskRepository 定义任务数据的存储接口
type TaskRepository interface {
	GetAllTasks() ([]*entity.Task, error)
	GetTaskByID(id int64) (*entity.Task, error)
	SaveTaskRecord(record *entity.Record) error
	CreateTask(task *entity.Task) (int64, error)
	UpdateTaskStatus(id int64, status int8) error
}

// NewScheduler 创建一个新的调度器
func NewScheduler(config *config.Config) (*Scheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建调度器实例
	s := &Scheduler{
		cron:        cron.New(cron.WithSeconds()), // 支持秒级调度
		config:      config,
		jobs:        make(map[int64]cron.EntryID),
		jobsMutex:   sync.RWMutex{},
		runningJobs: make(chan *JobContext, config.Job.QueueSize),
		ctx:         ctx,
		cancel:      cancel,
	}

	// 创建HTTP任务执行器
	s.httpWorker = NewHTTPWorker(config.Job.HttpWorkers, s.runningJobs)

	// 创建gRPC任务执行器
	s.grpcWorker = NewGRPCWorker(config.Job.GrpcWorkers, s.runningJobs)

	return s, nil
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	if s.isRunning {
		return nil
	}

	logger.Info("Starting job scheduler...")

	// 首先启动工作线程
	go s.httpWorker.Start(s.ctx)
	go s.grpcWorker.Start(s.ctx)

	// 启动调度器
	s.cron.Start()
	s.isRunning = true

	// 加载所有任务
	if s.taskStore != nil {
		if err := s.LoadAllTasks(); err != nil {
			logger.Errorf("Failed to load tasks: %v", err)
		}
	}

	// 启动结果处理协程
	go s.processResults()

	logger.Info("Job scheduler started successfully")
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	if !s.isRunning {
		return
	}

	logger.Info("Stopping job scheduler...")

	// 停止接收新任务
	s.cron.Stop()

	// 发送取消信号给所有工作线程
	s.cancel()

	s.isRunning = false
	logger.Info("Job scheduler stopped")
}

// LoadAllTasks 从存储加载所有任务并添加到调度器
func (s *Scheduler) LoadAllTasks() error {
	tasks, err := s.taskStore.GetAllTasks()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.Status == 1 { // 只加载启用状态的任务
			if err := s.AddTask(task); err != nil {
				logger.Errorf("Failed to add task %s (ID: %d): %v", task.Name, task.ID, err)
				continue
			}
		}
	}

	logger.Infof("Loaded %d tasks into scheduler", len(tasks))
	return nil
}

// AddTask 添加一个任务到调度器
func (s *Scheduler) AddTask(task *entity.Task) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	// 如果任务已存在，先移除
	if entryID, found := s.jobs[task.ID]; found {
		s.cron.Remove(entryID)
		delete(s.jobs, task.ID)
	}

	// 创建一个任务副本以避免并发问题
	taskCopy := *task

	// 添加到调度器
	entryID, err := s.cron.AddFunc(task.Cron, func() {
		s.executeTask(&taskCopy)
	})
	if err != nil {
		return err
	}

	// 记录任务ID和对应的EntryID
	s.jobs[task.ID] = entryID
	logger.Infof("Added task to scheduler: (ID: %d, Cron: %s)", task.ID, task.Cron)
	return nil
}

// RemoveTask 从调度器中移除一个任务
func (s *Scheduler) RemoveTask(taskID int64) {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	if entryID, found := s.jobs[taskID]; found {
		s.cron.Remove(entryID)
		delete(s.jobs, taskID)
		logger.Infof("Removed task from scheduler: ID: %d", taskID)
	}
}

// executeTask 执行一个任务
func (s *Scheduler) executeTask(task *entity.Task) {
	// 创建执行上下文
	jobCtx := &JobContext{
		Task:       task,
		ExecutedAt: time.Now(),
		Done:       make(chan *JobResult, 1),
	}

	// 根据任务类型分发到不同的工作线程
	switch task.TaskType {
	case "HTTP":
		s.httpWorker.Submit(jobCtx)
	case "GRPC":
		s.grpcWorker.Submit(jobCtx)
	default:
		logger.Errorf("Unsupported task type: %s for task %d", task.TaskType, task.ID)
		return
	}

	// 不等待结果返回，由processResults协程处理结果
	logger.Infof("Task submitted for execution: %s (ID: %d, Type: %s)", task.Name, task.ID, task.TaskType)
}

// processResults 处理任务执行结果
func (s *Scheduler) processResults() {
	for {
		select {
		case <-s.ctx.Done():
			logger.Info("Result processor shutting down")
			return
		case result := <-s.httpWorker.Results():
			s.saveTaskResult(result.Task.Task, result.Result)
		case result := <-s.grpcWorker.Results():
			s.saveTaskResult(result.Task.Task, result.Result)
		}
	}
}

// saveTaskResult 保存任务执行结果
func (s *Scheduler) saveTaskResult(task *entity.Task, result *JobResult) {
	// Create record time
	now := time.Now()

	// Update task execution time tracking
	s.UpdateTaskExecutionTime(task.ID, now)

	// 创建执行记录
	record := &entity.Record{
		TaskID:       task.ID,
		TaskName:     task.Name,
		TaskType:     task.TaskType,
		DepartmentID: task.DepartmentID,
		RetryTimes:   result.RetryTimes,
		UseFallback:  boolToInt8(result.UseFallback),
		CostTime:     result.CostTime,
		CreateTime:   now,
	}

	// 根据任务类型设置记录字段
	if task.TaskType == "HTTP" {
		record.URL = result.ActualURL
		record.HTTPMethod = result.ActualMethod
		record.Body = task.Body
		record.Headers = task.Headers
		record.Response = result.Response
		record.StatusCode = result.StatusCode
		record.Success = boolToInt8(result.Success)
	} else if task.TaskType == "GRPC" {
		record.GrpcService = task.GrpcService
		record.GrpcMethod = task.GrpcMethod
		record.GrpcParams = task.GrpcParams
		record.Response = result.Response
		record.GrpcStatus = result.GrpcStatus
		record.Success = boolToInt8(result.Success)
	}

	// 保存记录到数据库
	if s.taskStore != nil {
		if err := s.taskStore.SaveTaskRecord(record); err != nil {
			logger.Errorf("Failed to save task record: %v", err)
		}
	}

	// 记录执行结果日志
	if result.Success {
		logger.Infof("Task executed successfully: %s (ID: %d, Retry: %d, UseFallback: %v, CostTime: %dms)",
			task.Name, task.ID, result.RetryTimes, result.UseFallback, result.CostTime)
	} else {
		logger.Errorf("Task execution failed: %s (ID: %d, Retry: %d, UseFallback: %v, Error: %v)",
			task.Name, task.ID, result.RetryTimes, result.UseFallback, result.Error)
	}
}

// SetTaskRepository 设置任务数据存储库
func (s *Scheduler) SetTaskRepository(repo TaskRepository) {
	s.taskStore = repo
}

// boolToInt8 将bool转换为int8 (0-false, 1-true)
func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

// AddTaskAndStore adds a task to the scheduler and stores it in the repository
func (s *Scheduler) AddTaskAndStore(task *entity.Task) (int64, error) {
	// Store task in repository first
	if s.taskStore == nil {
		return 0, errors.New("task repository not initialized")
	}

	taskID, err := s.taskStore.CreateTask(task)
	if err != nil {
		return 0, err
	}

	// Set the ID from the repository
	task.ID = taskID

	// Add task to scheduler
	if err := s.AddTask(task); err != nil {
		return taskID, err
	}

	return taskID, nil
}

// PauseTask pauses a running task
func (s *Scheduler) PauseTask(taskID int64) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	// Check if task exists in scheduler
	if _, found := s.jobs[taskID]; !found {
		return fmt.Errorf("task with ID %d not found in scheduler", taskID)
	}

	// Remove from cron scheduler
	s.RemoveTask(taskID)

	// Update task status in database
	if s.taskStore != nil {
		if err := s.taskStore.UpdateTaskStatus(taskID, int8(0)); err != nil {
			return err
		}
	}

	logger.Infof("Task paused: ID %d", taskID)
	return nil
}

// ResumeTask resumes a paused task
func (s *Scheduler) ResumeTask(taskID int64) error {
	// Get the task from the repository
	if s.taskStore == nil {
		return errors.New("task repository not initialized")
	}

	task, err := s.taskStore.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	// Update status in database
	if err := s.taskStore.UpdateTaskStatus(taskID, int8(1)); err != nil {
		return err
	}

	// Add task back to scheduler
	if err := s.AddTask(task); err != nil {
		return err
	}

	logger.Infof("Task resumed: ID %d", taskID)
	return nil
}

// GetTaskStatus retrieves the current status of a task
func (s *Scheduler) GetTaskStatus(taskID int64) (*entity.Task, error) {
	// Get the task from the repository
	if s.taskStore == nil {
		return nil, errors.New("task repository not initialized")
	}

	task, err := s.taskStore.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}

	// Check if task is currently scheduled
	s.jobsMutex.RLock()
	_, isScheduled := s.jobs[taskID]
	s.jobsMutex.RUnlock()

	// Update task with scheduler status
	if isScheduled {
		// Get the cron entry to determine next execution time
		entryID := s.jobs[taskID]
		entry := s.cron.Entry(entryID)
		nextTime := entry.Next
		task.NextExecuteTime = &nextTime
	}

	return task, nil
}

// UpdateTaskExecutionTime updates the last execution time of a task
func (s *Scheduler) UpdateTaskExecutionTime(taskID int64, executionTime time.Time) error {
	// Get the task from the repository
	if s.taskStore == nil {
		return errors.New("task repository not initialized")
	}

	task, err := s.taskStore.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	// Update the last execution time
	task.LastExecuteTime = &executionTime

	// Calculate next execution time based on cron expression
	s.jobsMutex.RLock()
	entryID, found := s.jobs[taskID]
	s.jobsMutex.RUnlock()

	if found {
		entry := s.cron.Entry(entryID)
		nextTime := entry.Next
		task.NextExecuteTime = &nextTime
	}

	// Update task in database if there's a specialized method for it
	// For now, we'll just log the update
	logger.Infof("Updated task execution time: %s (ID: %d, Last: %v, Next: %v)",
		task.Name, task.ID, executionTime, task.NextExecuteTime)

	return nil
}
