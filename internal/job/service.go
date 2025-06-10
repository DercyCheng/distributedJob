package job

import (
	"context"
	"encoding/json"
	"fmt"
	"go-job/api/grpc"
	"go-job/internal/models"
	"go-job/pkg/database"
	"go-job/pkg/logger"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// Service 任务服务
type Service struct {
	grpc.UnimplementedJobServiceServer
	db *gorm.DB
}

// NewService 创建任务服务
func NewService() *Service {
	return &Service{
		db: database.GetDB(),
	}
}

// CreateJob 创建任务
func (s *Service) CreateJob(ctx context.Context, req *grpc.CreateJobRequest) (*grpc.CreateJobResponse, error) {
	logger.Infof("创建任务: %s", req.GetName())

	// 验证 Cron 表达式
	if err := validateCron(req.GetCron()); err != nil {
		return nil, fmt.Errorf("无效的 Cron 表达式: %w", err)
	}

	// 转换参数为 JSON
	paramsJSON, _ := json.Marshal(req.GetParams())

	job := &models.Job{
		ID:            uuid.New().String(),
		Name:          req.GetName(),
		Description:   req.GetDescription(),
		Cron:          req.GetCron(),
		Command:       req.GetCommand(),
		Params:        string(paramsJSON),
		Enabled:       true,
		RetryAttempts: int(req.GetRetryAttempts()),
		Timeout:       int(req.GetTimeout()),
		CreatedBy:     "system", // TODO: 从上下文获取用户信息
	}

	if err := s.db.Create(job).Error; err != nil {
		logger.WithError(err).Error("创建任务失败")
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	// 转换为 gRPC 消息
	grpcJob := s.modelToGrpc(job)

	logger.Infof("任务创建成功: %s (ID: %s)", job.Name, job.ID)

	return &grpc.CreateJobResponse{
		Job: grpcJob,
	}, nil
}

// GetJob 获取任务
func (s *Service) GetJob(ctx context.Context, req *grpc.GetJobRequest) (*grpc.GetJobResponse, error) {
	var job models.Job
	if err := s.db.First(&job, "id = ?", req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("任务不存在: %s", req.GetId())
		}
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	grpcJob := s.modelToGrpc(&job)

	return &grpc.GetJobResponse{
		Job: grpcJob,
	}, nil
}

// ListJobs 获取任务列表
func (s *Service) ListJobs(ctx context.Context, req *grpc.ListJobsRequest) (*grpc.ListJobsResponse, error) {
	page := req.GetPage()
	size := req.GetSize()
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	query := s.db.Model(&models.Job{})

	// 关键字搜索
	if keyword := req.GetKeyword(); keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 启用状态过滤
	if req.GetEnabled() {
		query = query.Where("enabled = ?", true)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("查询任务总数失败: %w", err)
	}

	// 分页查询
	var jobs []models.Job
	offset := (page - 1) * size
	if err := query.Offset(int(offset)).Limit(int(size)).Order("created_at DESC").Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("查询任务列表失败: %w", err)
	}

	// 转换为 gRPC 消息
	var grpcJobs []*grpc.Job
	for _, job := range jobs {
		grpcJobs = append(grpcJobs, s.modelToGrpc(&job))
	}

	return &grpc.ListJobsResponse{
		Jobs:  grpcJobs,
		Total: total,
	}, nil
}

// UpdateJob 更新任务
func (s *Service) UpdateJob(ctx context.Context, req *grpc.UpdateJobRequest) (*grpc.UpdateJobResponse, error) {
	logger.Infof("更新任务: %s", req.GetId())

	// 验证 Cron 表达式
	if err := validateCron(req.GetCron()); err != nil {
		return nil, fmt.Errorf("无效的 Cron 表达式: %w", err)
	}

	// 查找任务
	var job models.Job
	if err := s.db.First(&job, "id = ?", req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("任务不存在: %s", req.GetId())
		}
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	// 转换参数为 JSON
	paramsJSON, _ := json.Marshal(req.GetParams())

	// 更新字段
	updates := map[string]interface{}{
		"name":           req.GetName(),
		"description":    req.GetDescription(),
		"cron":           req.GetCron(),
		"command":        req.GetCommand(),
		"params":         string(paramsJSON),
		"enabled":        req.GetEnabled(),
		"retry_attempts": req.GetRetryAttempts(),
		"timeout":        req.GetTimeout(),
		"updated_at":     time.Now(),
	}

	if err := s.db.Model(&job).Updates(updates).Error; err != nil {
		logger.WithError(err).Error("更新任务失败")
		return nil, fmt.Errorf("更新任务失败: %w", err)
	}

	// 重新查询更新后的任务
	if err := s.db.First(&job, "id = ?", req.GetId()).Error; err != nil {
		return nil, fmt.Errorf("查询更新后的任务失败: %w", err)
	}

	grpcJob := s.modelToGrpc(&job)

	logger.Infof("任务更新成功: %s", job.ID)

	return &grpc.UpdateJobResponse{
		Job: grpcJob,
	}, nil
}

// DeleteJob 删除任务
func (s *Service) DeleteJob(ctx context.Context, req *grpc.DeleteJobRequest) (*grpc.DeleteJobResponse, error) {
	logger.Infof("删除任务: %s", req.GetId())

	// 软删除任务
	result := s.db.Delete(&models.Job{}, "id = ?", req.GetId())
	if result.Error != nil {
		logger.WithError(result.Error).Error("删除任务失败")
		return nil, fmt.Errorf("删除任务失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("任务不存在: %s", req.GetId())
	}

	logger.Infof("任务删除成功: %s", req.GetId())

	return &grpc.DeleteJobResponse{
		Success: true,
	}, nil
}

// TriggerJob 手动触发任务
func (s *Service) TriggerJob(ctx context.Context, req *grpc.TriggerJobRequest) (*grpc.TriggerJobResponse, error) {
	logger.Infof("手动触发任务: %s", req.GetId())

	// 查找任务
	var job models.Job
	if err := s.db.First(&job, "id = ? AND enabled = ?", req.GetId(), true).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("任务不存在或已禁用: %s", req.GetId())
		}
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}

	// 创建执行记录
	execution := &models.JobExecution{
		ID:     uuid.New().String(),
		JobID:  job.ID,
		Status: models.ExecutionStatusPending,
	}

	if err := s.db.Create(execution).Error; err != nil {
		logger.WithError(err).Error("创建执行记录失败")
		return nil, fmt.Errorf("创建执行记录失败: %w", err)
	}

	// 创建调度记录
	schedule := &models.JobSchedule{
		ID:          uuid.New().String(),
		JobID:       job.ID,
		ScheduledAt: time.Now(),
		Status:      models.ScheduleStatusPending,
		ExecutionID: execution.ID,
	}

	if err := s.db.Create(schedule).Error; err != nil {
		logger.WithError(err).Error("创建调度记录失败")
		return nil, fmt.Errorf("创建调度记录失败: %w", err)
	}

	logger.Infof("任务手动触发成功: %s (执行ID: %s)", job.ID, execution.ID)

	return &grpc.TriggerJobResponse{
		ExecutionId: execution.ID,
	}, nil
}

// modelToGrpc 将模型转换为 gRPC 消息
func (s *Service) modelToGrpc(job *models.Job) *grpc.Job {
	var params map[string]string
	if job.Params != "" {
		json.Unmarshal([]byte(job.Params), &params)
	}

	return &grpc.Job{
		Id:            job.ID,
		Name:          job.Name,
		Description:   job.Description,
		Cron:          job.Cron,
		Command:       job.Command,
		Params:        params,
		Enabled:       job.Enabled,
		RetryAttempts: int32(job.RetryAttempts),
		Timeout:       int32(job.Timeout),
		CreatedAt:     timestamppb.New(job.CreatedAt),
		UpdatedAt:     timestamppb.New(job.UpdatedAt),
		CreatedBy:     job.CreatedBy,
	}
}

// validateCron 验证 Cron 表达式
func validateCron(cronExpr string) error {
	// 这里可以使用 cron 库验证表达式
	// 简单验证，实际项目中应该使用更严格的验证
	if cronExpr == "" {
		return fmt.Errorf("Cron 表达式不能为空")
	}
	// TODO: 添加更详细的 Cron 表达式验证
	return nil
}
