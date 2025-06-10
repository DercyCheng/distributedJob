package scheduler

import (
	"context"
	"fmt"
	"go-job/api/grpc"
	"go-job/internal/models"
	"go-job/pkg/config"
	"go-job/pkg/database"
	"go-job/pkg/logger"
	"go-job/pkg/redis"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// Service 调度器服务
type Service struct {
	grpc.UnimplementedSchedulerServiceServer
	config    *config.Config
	cron      *cron.Cron
	workers   map[string]*WorkerInfo
	workersMu sync.RWMutex
	db        *gorm.DB
	taskQueue chan *models.JobSchedule
	quit      chan struct{}
}

// WorkerInfo 工作节点信息
type WorkerInfo struct {
	ID          string
	Name        string
	IP          string
	Port        int32
	Status      grpc.WorkerStatus
	Capacity    int32
	CurrentLoad int32
	LastSeen    time.Time
	Metadata    map[string]string
}

// NewService 创建调度器服务
func NewService(cfg *config.Config) *Service {
	location, _ := time.LoadLocation(cfg.Scheduler.Timezone)

	return &Service{
		config:    cfg,
		cron:      cron.New(cron.WithLocation(location)),
		workers:   make(map[string]*WorkerInfo),
		db:        database.GetDB(),
		taskQueue: make(chan *models.JobSchedule, 1000),
		quit:      make(chan struct{}),
	}
}

// Start 启动调度器
func (s *Service) Start(ctx context.Context) error {
	logger.Info("启动调度器服务")

	// 启动 cron 调度器
	s.cron.Start()

	// 加载现有任务
	if err := s.loadJobs(); err != nil {
		return fmt.Errorf("加载任务失败: %w", err)
	}

	// 启动工作节点监控
	go s.monitorWorkers(ctx)

	// 启动任务分发器
	go s.taskDispatcher(ctx)

	// 启动任务清理器
	go s.taskCleaner(ctx)

	<-ctx.Done()
	logger.Info("调度器服务已停止")
	return nil
}

// Stop 停止调度器
func (s *Service) Stop() {
	logger.Info("正在停止调度器服务")
	s.cron.Stop()
	close(s.quit)
}

// loadJobs 加载数据库中的任务
func (s *Service) loadJobs() error {
	var jobs []models.Job
	if err := s.db.Where("enabled = ?", true).Find(&jobs).Error; err != nil {
		return err
	}

	for _, job := range jobs {
		if err := s.addJobToCron(job); err != nil {
			logger.WithError(err).Errorf("添加任务到 cron 失败: %s", job.Name)
		}
	}

	logger.Infof("已加载 %d 个任务", len(jobs))
	return nil
}

// addJobToCron 添加任务到 cron 调度器
func (s *Service) addJobToCron(job models.Job) error {
	entryID, err := s.cron.AddFunc(job.Cron, func() {
		s.scheduleJob(job.ID)
	})
	if err != nil {
		return err
	}

	// 将任务 ID 和 cron entry ID 的映射存储到 Redis
	key := fmt.Sprintf("job_cron:%s", job.ID)
	return redis.Set(context.Background(), key, fmt.Sprintf("%d", entryID), 0)
}

// scheduleJob 调度任务
func (s *Service) scheduleJob(jobID string) {
	logger.Infof("调度任务: %s", jobID)

	schedule := &models.JobSchedule{
		ID:          uuid.New().String(),
		JobID:       jobID,
		ScheduledAt: time.Now(),
		Status:      models.ScheduleStatusPending,
	}

	if err := s.db.Create(schedule).Error; err != nil {
		logger.WithError(err).Errorf("创建任务调度记录失败: %s", jobID)
		return
	}

	// 将任务加入队列
	select {
	case s.taskQueue <- schedule:
		logger.Debugf("任务已加入队列: %s", jobID)
	default:
		logger.Warnf("任务队列已满，跳过任务: %s", jobID)
	}
}

// taskDispatcher 任务分发器
func (s *Service) taskDispatcher(ctx context.Context) {
	logger.Info("启动任务分发器")

	for {
		select {
		case <-ctx.Done():
			return
		case schedule := <-s.taskQueue:
			s.dispatchTask(schedule)
		}
	}
}

// dispatchTask 分发任务
func (s *Service) dispatchTask(schedule *models.JobSchedule) {
	// 查找可用的工作节点
	worker := s.findAvailableWorker()
	if worker == nil {
		logger.Warnf("没有可用的工作节点，任务将被重新调度: %s", schedule.JobID)
		// 重新调度
		time.AfterFunc(30*time.Second, func() {
			select {
			case s.taskQueue <- schedule:
			default:
			}
		})
		return
	}

	// 更新调度记录
	schedule.WorkerID = worker.ID
	schedule.Status = models.ScheduleStatusAssigned
	if err := s.db.Save(schedule).Error; err != nil {
		logger.WithError(err).Errorf("更新调度记录失败: %s", schedule.ID)
		return
	}

	// 创建执行记录
	execution := &models.JobExecution{
		ID:       uuid.New().String(),
		JobID:    schedule.JobID,
		WorkerID: worker.ID,
		Status:   models.ExecutionStatusPending,
	}

	if err := s.db.Create(execution).Error; err != nil {
		logger.WithError(err).Errorf("创建执行记录失败: %s", schedule.JobID)
		return
	}

	// 更新调度记录的执行 ID
	schedule.ExecutionID = execution.ID
	if err := s.db.Save(schedule).Error; err != nil {
		logger.WithError(err).Errorf("更新调度记录执行ID失败: %s", schedule.ID)
	}

	// 更新工作节点负载
	s.workersMu.Lock()
	worker.CurrentLoad++
	s.workersMu.Unlock()

	logger.Infof("任务 %s 已分配给工作节点 %s", schedule.JobID, worker.ID)
}

// findAvailableWorker 查找可用的工作节点
func (s *Service) findAvailableWorker() *WorkerInfo {
	s.workersMu.RLock()
	defer s.workersMu.RUnlock()

	var bestWorker *WorkerInfo
	minLoad := int32(1000)

	for _, worker := range s.workers {
		if worker.Status == grpc.WorkerStatus_ONLINE &&
			worker.CurrentLoad < worker.Capacity &&
			worker.CurrentLoad < minLoad {
			bestWorker = worker
			minLoad = worker.CurrentLoad
		}
	}

	return bestWorker
}

// monitorWorkers 监控工作节点
func (s *Service) monitorWorkers(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.config.Scheduler.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkWorkersHealth()
		}
	}
}

// checkWorkersHealth 检查工作节点健康状态
func (s *Service) checkWorkersHealth() {
	s.workersMu.Lock()
	defer s.workersMu.Unlock()

	timeout := time.Duration(s.config.Scheduler.HeartbeatInterval*2) * time.Second
	now := time.Now()

	for id, worker := range s.workers {
		if now.Sub(worker.LastSeen) > timeout {
			logger.Warnf("工作节点 %s 心跳超时，标记为离线", id)
			worker.Status = grpc.WorkerStatus_OFFLINE

			// 更新数据库中的工作节点状态
			s.db.Model(&models.Worker{}).Where("id = ?", id).Update("status", models.WorkerStatusOffline)
		}
	}
}

// taskCleaner 任务清理器
func (s *Service) taskCleaner(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanupOldTasks()
		}
	}
}

// cleanupOldTasks 清理旧的任务记录
func (s *Service) cleanupOldTasks() {
	// 清理 7 天前的执行记录
	cutoff := time.Now().AddDate(0, 0, -7)

	result := s.db.Where("created_at < ?", cutoff).Delete(&models.JobExecution{})
	if result.Error != nil {
		logger.WithError(result.Error).Error("清理旧执行记录失败")
	} else if result.RowsAffected > 0 {
		logger.Infof("已清理 %d 条旧执行记录", result.RowsAffected)
	}

	// 清理 30 天前的调度记录
	cutoff = time.Now().AddDate(0, 0, -30)
	result = s.db.Where("created_at < ?", cutoff).Delete(&models.JobSchedule{})
	if result.Error != nil {
		logger.WithError(result.Error).Error("清理旧调度记录失败")
	} else if result.RowsAffected > 0 {
		logger.Infof("已清理 %d 条旧调度记录", result.RowsAffected)
	}
}
