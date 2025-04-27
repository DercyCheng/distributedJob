package service

import (
	"errors"

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
