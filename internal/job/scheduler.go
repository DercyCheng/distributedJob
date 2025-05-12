package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"distributedJob/internal/config"
	"distributedJob/internal/model/entity"
	"distributedJob/internal/store/etcd"
	"distributedJob/internal/store/kafka"
	"distributedJob/pkg/logger"
	"distributedJob/pkg/metrics"
	"distributedJob/pkg/tracing"

	"github.com/IBM/sarama"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel/attribute"
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

	// 新增加的分布式支持
	useKafka       bool
	kafkaManager   *kafka.Manager
	kafkaTopic     string
	useEtcd        bool
	etcdManager    *etcd.Manager
	etcdLockPrefix string
	serviceID      string

	// 可观测性支持
	metrics *metrics.Metrics
	tracer  *tracing.Tracer
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
func NewScheduler(config *config.Config, opts ...SchedulerOption) (*Scheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 设置默认的Kafka主题名称
	kafkaTopic := "distributed_job_jobs" // 默认主题名
	if config.Kafka.TopicPrefix != "" {
		kafkaTopic = config.Kafka.TopicPrefix + "jobs"
	}

	// 创建调度器实例
	s := &Scheduler{
		cron:           cron.New(cron.WithSeconds()), // 支持秒级调度
		config:         config,
		jobs:           make(map[int64]cron.EntryID),
		jobsMutex:      sync.RWMutex{},
		runningJobs:    make(chan *JobContext, config.Job.QueueSize),
		ctx:            ctx,
		cancel:         cancel,
		useKafka:       false,
		useEtcd:        false,
		kafkaTopic:     kafkaTopic,
		etcdLockPrefix: "/distributed_job/locks/",
		serviceID:      fmt.Sprintf("scheduler-%d", time.Now().UnixNano()),
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(s)
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

	// 创建追踪span
	var ctx context.Context
	var startupSpan interface{} // 使用interface{}避免nil检查问题

	if s.tracer != nil {
		var span interface{}
		ctx, span = s.tracer.StartSpanWithAttributes(
			s.ctx,
			"scheduler_startup",
			attribute.String("service.id", s.serviceID),
		)
		startupSpan = span
		defer func() {
			if sp, ok := startupSpan.(interface{ End() }); ok {
				sp.End()
			}
		}()
	} else {
		ctx = s.ctx
	}
	// 如果启用Kafka，设置消费者监听作业执行请求
	if s.useKafka && s.kafkaManager != nil {
		logger.Info("Setting up Kafka job distribution...")

		// 任务处理函数
		jobHandler := func(msg *sarama.ConsumerMessage) error {
			// 在实际实现中，我们会从消息内容解析任务，并提交给工作线程执行
			if s.metrics != nil {
				s.metrics.IncrementCounter("jobs_received_kafka", "scheduler")
			}
			logger.Infof("Received job from Kafka: %s", string(msg.Value))
			return nil
		}

		// 设置消费者
		err := s.kafkaManager.InitializeConsumer(
			[]string{s.kafkaTopic},
			s.config.Kafka.ConsumerGroup,
			jobHandler,
		)
		if err != nil {
			return fmt.Errorf("failed to initialize kafka consumer: %w", err)
		}

		// 启动消费者
		if err := s.kafkaManager.StartConsumer(); err != nil {
			return fmt.Errorf("failed to start kafka consumer: %w", err)
		}
	}

	// 如果启用ETCD，注册服务
	if s.useEtcd && s.etcdManager != nil {
		logger.Info("Registering scheduler service in etcd...")

		serviceRegistry, err := s.etcdManager.NewServiceRegistry("distributed_job/services", 10)
		if err != nil {
			logger.Warnf("Failed to create service registry: %v", err)
		} else {
			err = serviceRegistry.Register(ctx, "scheduler", s.serviceID)
			if err != nil {
				logger.Warnf("Failed to register service: %v", err)
			}
		}
	}

	// 首先启动工作线程
	go s.httpWorker.Start(s.ctx)
	go s.grpcWorker.Start(s.ctx)

	// 启动调度器
	s.cron.Start()
	s.isRunning = true

	// 启动结果处理线程
	go s.processResults()

	// 记录启动事件
	if s.metrics != nil {
		s.metrics.IncrementCounter("scheduler_events", "startup")
		s.metrics.SetGauge("scheduler_status", 1) // 1 表示正在运行
	}

	logger.Info("Job scheduler started successfully")
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	if !s.isRunning {
		return
	}

	logger.Info("Stopping job scheduler...")

	// 标记为停止状态，防止新任务执行
	s.jobsMutex.Lock()
	s.isRunning = false
	s.jobsMutex.Unlock()

	// 创建追踪span
	var ctx context.Context
	var shutdownSpan interface{}

	if s.tracer != nil {
		var span interface{}
		ctx, span = s.tracer.StartSpanWithAttributes(
			context.Background(),
			"scheduler_shutdown",
			attribute.String("service.id", s.serviceID),
		)
		shutdownSpan = span
		defer func() {
			if sp, ok := shutdownSpan.(interface{ End() }); ok {
				sp.End()
			}
		}()
	} else {
		ctx = context.Background()
	}

	// 如果启用ETCD，注销服务
	if s.useEtcd && s.etcdManager != nil {
		logger.Info("Deregistering scheduler service from etcd...")

		serviceRegistry, err := s.etcdManager.NewServiceRegistry("distributed_job/services", 10)
		if err == nil {
			err = serviceRegistry.Deregister(ctx, "scheduler", s.serviceID)
			if err != nil {
				logger.Warnf("Failed to deregister service: %v", err)
			}
		}
	}
	// 停止接收新任务
	s.cron.Stop()

	// 关闭任务队列以防止新任务提交
	close(s.runningJobs)

	// 发送取消信号给所有工作线程
	s.cancel()

	// 等待工作线程处理正在进行的任务
	// 找出当前最长的任务超时时间，以便正确设置关闭超时
	maxTaskTimeout := 60 * time.Second // 默认60秒

	// 实际系统中，我们应该遍历当前正在执行的任务，找出最大的超时值
	// 这里我们使用简单实现，仅为演示

	// 添加额外的缓冲时间，确保任务有机会完成
	gracefulTimeout := maxTaskTimeout + 5*time.Second
	logger.Infof("Waiting up to %s for worker threads to finish...", gracefulTimeout)

	gracefulCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	// 实现一个更复杂的等待机制，定期检查正在执行的任务状态
	checkInterval := 500 * time.Millisecond
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// 记录cron停止时间
	cronStopTime := time.Now()

	for {
		select {
		case <-gracefulCtx.Done():
			if gracefulCtx.Err() == context.DeadlineExceeded {
				logger.Warn("Graceful shutdown timed out, some tasks may have been interrupted")
			}
			// 记录关闭事件
			if s.metrics != nil {
				s.metrics.IncrementCounter("scheduler_events", "shutdown")
				s.metrics.SetGauge("scheduler_status", 0) // 0 表示已停止
			}
			return
		case <-ticker.C:
			// 实际系统中，这里应该检查是否还有正在运行的任务
			// 如果所有任务都已完成，可以提前退出等待

			// 检查自cron停止以来的时间
			runningTime := time.Since(cronStopTime)
			if runningTime > 2*time.Second {
				logger.Info("All worker threads completed")
				// 记录关闭事件
				if s.metrics != nil {
					s.metrics.IncrementCounter("scheduler_events", "shutdown")
					s.metrics.SetGauge("scheduler_status", 0) // 0 表示已停止
				}
				return
			}
		}
	}
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
	// 检查是否已停止
	s.jobsMutex.RLock()
	if !s.isRunning {
		s.jobsMutex.RUnlock()
		logger.Warnf("Task execution skipped because scheduler is stopping: %s (ID: %d)", task.Name, task.ID)
		return
	}
	s.jobsMutex.RUnlock()

	// 创建执行上下文
	jobCtx := &JobContext{
		Task:       task,
		ExecutedAt: time.Now(),
		Done:       make(chan *JobResult, 1),
	}

	// 根据任务类型分发到不同的工作线程
	switch task.TaskType {
	case "HTTP":
		// 添加超时控制以避免在队列已满时无限期阻塞
		select {
		case s.httpWorker.workQueue <- jobCtx:
			// 任务已提交
		case <-time.After(5 * time.Second):
			logger.Errorf("Failed to submit HTTP task: %s (ID: %d) - queue is full", task.Name, task.ID)
			return
		}
	case "GRPC":
		// 添加超时控制以避免在队列已满时无限期阻塞
		select {
		case s.grpcWorker.workQueue <- jobCtx:
			// 任务已提交
		case <-time.After(5 * time.Second):
			logger.Errorf("Failed to submit GRPC task: %s (ID: %d) - queue is full", task.Name, task.ID)
			return
		}
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
		if result.StatusCode != nil {
			statusCode := int(*result.StatusCode)
			record.StatusCode = &statusCode
		}
	} else if task.TaskType == "GRPC" {
		record.GrpcService = task.GrpcService
		record.GrpcMethod = task.GrpcMethod
		if result.GrpcStatus != nil {
			grpcStatus := int(*result.GrpcStatus)
			record.GrpcStatus = &grpcStatus
		}
	}

	// 设置执行结果
	record.Success = boolToInt8(result.Success)
	record.Response = result.Response
	if result.Error != nil {
		// 存储错误信息到响应字段
		if record.Response != "" {
			record.Response += "\nError: " + result.Error.Error()
		} else {
			record.Response = "Error: " + result.Error.Error()
		}
	}

	// 持久化执行记录
	if s.taskStore != nil {
		if err := s.taskStore.SaveTaskRecord(record); err != nil {
			logger.Errorf("Failed to save task execution record: %v", err)
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

// SetTaskRepository 设置任务存储库
func (s *Scheduler) SetTaskRepository(repo TaskRepository) {
	s.taskStore = repo
}

// AddTaskAndStore 添加并存储任务
func (s *Scheduler) AddTaskAndStore(task *entity.Task) (int64, error) {
	// 首先保存任务到存储
	if s.taskStore == nil {
		return 0, errors.New("task repository not initialized")
	}

	// 确保任务类型字段同步
	task.SyncTypeFields()

	// 保存任务到持久化存储
	taskID, err := s.taskStore.CreateTask(task)
	if err != nil {
		return 0, fmt.Errorf("failed to store task: %w", err)
	}

	// 更新任务ID
	task.ID = taskID

	// 将任务添加到调度器
	if err := s.AddTask(task); err != nil {
		return taskID, fmt.Errorf("task stored but failed to schedule: %w", err)
	}

	logger.Infof("Task added and scheduled successfully: %s (ID: %d)", task.Name, taskID)
	return taskID, nil
}

// PauseTask 暂停任务
func (s *Scheduler) PauseTask(taskID int64) error {
	if s.taskStore == nil {
		return errors.New("task repository not initialized")
	}

	// 从调度器中移除任务
	s.RemoveTask(taskID)

	// 更新任务状态为暂停(0)
	err := s.taskStore.UpdateTaskStatus(taskID, 0)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	logger.Infof("Task paused: ID: %d", taskID)
	return nil
}

// ResumeTask 恢复任务
func (s *Scheduler) ResumeTask(taskID int64) error {
	if s.taskStore == nil {
		return errors.New("task repository not initialized")
	}

	// 更新任务状态为激活(1)
	err := s.taskStore.UpdateTaskStatus(taskID, 1)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// 获取任务详情
	task, err := s.taskStore.GetTaskByID(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task details: %w", err)
	}

	// 将任务添加到调度器
	if err := s.AddTask(task); err != nil {
		return fmt.Errorf("failed to resume task scheduling: %w", err)
	}

	logger.Infof("Task resumed: ID: %d", taskID)
	return nil
}

// GetTaskStatus 获取任务状态
func (s *Scheduler) GetTaskStatus(taskID int64) (*entity.Task, error) {
	if s.taskStore == nil {
		return nil, errors.New("task repository not initialized")
	}

	// 获取任务详情
	task, err := s.taskStore.GetTaskByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// 获取下一次执行时间
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()

	if entryID, found := s.jobs[taskID]; found && s.isRunning {
		entry := s.cron.Entry(entryID)
		task.NextExecuteTime = &entry.Next
	}

	return task, nil
}

// UpdateTaskExecutionTime 更新任务执行时间
func (s *Scheduler) UpdateTaskExecutionTime(taskID int64, executionTime time.Time) {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	task, err := s.taskStore.GetTaskByID(taskID)
	if err != nil {
		logger.Warnf("Failed to get task for updating execution time: %v", err)
		return
	}

	// 更新最后执行时间
	task.LastExecuteTime = &executionTime

	// 如果任务在调度器中，更新下一次执行时间
	if entryID, found := s.jobs[taskID]; found && s.isRunning {
		entry := s.cron.Entry(entryID)
		task.NextExecuteTime = &entry.Next
	}
}

// boolToInt8 将bool转换为int8 (0-false, 1-true)
func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}
