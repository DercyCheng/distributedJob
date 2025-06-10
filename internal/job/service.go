package job

import (
	"context"
	"encoding/json"
	"fmt"
	"go-job/api/grpc"
	"go-job/internal/models"
	"go-job/pkg/database"
	"go-job/pkg/logger"
	"strconv"
	"strings"
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
		CreatedBy:     getUserFromContext(ctx), // 从上下文获取用户信息
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
	// 详细的 Cron 表达式验证
	if cronExpr == "" {
		return fmt.Errorf("cron 表达式不能为空")
	}

	// 使用 cron 库验证表达式
	// 支持标准的 5 字段格式: 分 时 日 月 周
	// 也支持扩展的 6 字段格式: 秒 分 时 日 月 周
	fields := strings.Fields(cronExpr)

	// 检查字段数量
	if len(fields) != 5 && len(fields) != 6 {
		return fmt.Errorf("cron 表达式必须包含 5 或 6 个字段，当前有 %d 个字段", len(fields))
	}

	// 基本字段验证
	for i, field := range fields {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("第 %d 个字段不能为空", i+1)
		}

		// 检查特殊字符
		if !isValidCronField(field) {
			return fmt.Errorf("第 %d 个字段包含无效字符: %s", i+1, field)
		}
	}

	// 基本范围验证
	if len(fields) == 5 {
		// 标准格式: 分 时 日 月 周
		ranges := []struct {
			min, max int
			name     string
		}{
			{0, 59, "分钟"},
			{0, 23, "小时"},
			{1, 31, "日"},
			{1, 12, "月"},
			{0, 7, "周"}, // 0 和 7 都表示周日
		}

		for i, r := range ranges {
			if err := validateCronFieldRange(fields[i], r.min, r.max, r.name); err != nil {
				return err
			}
		}
	} else {
		// 6字段格式: 秒 分 时 日 月 周
		ranges := []struct {
			min, max int
			name     string
		}{
			{0, 59, "秒"},
			{0, 59, "分钟"},
			{0, 23, "小时"},
			{1, 31, "日"},
			{1, 12, "月"},
			{0, 7, "周"}, // 0 和 7 都表示周日
		}

		for i, r := range ranges {
			if err := validateCronFieldRange(fields[i], r.min, r.max, r.name); err != nil {
				return err
			}
		}
	}

	return nil
}

// isValidCronField 检查字段是否包含有效的 cron 字符
func isValidCronField(field string) bool {
	// 允许的字符: 数字、*、?、-、,、/、L、W、#
	validChars := "0123456789*?-,/LW#"
	for _, char := range field {
		found := false
		for _, validChar := range validChars {
			if char == validChar {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// validateCronFieldRange 验证字段的数值范围
func validateCronFieldRange(field string, min, max int, fieldName string) error {
	// 跳过特殊字符
	if field == "*" || field == "?" {
		return nil
	}

	// 处理范围表达式 (如 1-5)
	if strings.Contains(field, "-") {
		parts := strings.Split(field, "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(parts[0])
			end, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil {
				return fmt.Errorf("%s字段范围格式错误: %s", fieldName, field)
			}
			if start < min || start > max || end < min || end > max {
				return fmt.Errorf("%s字段范围超出有效范围 [%d-%d]: %s", fieldName, min, max, field)
			}
			if start > end {
				return fmt.Errorf("%s字段范围开始值不能大于结束值: %s", fieldName, field)
			}
			return nil
		}
	}

	// 处理列表表达式 (如 1,3,5)
	if strings.Contains(field, ",") {
		values := strings.Split(field, ",")
		for _, value := range values {
			if err := validateSingleCronValue(strings.TrimSpace(value), min, max, fieldName); err != nil {
				return err
			}
		}
		return nil
	}

	// 处理步长表达式 (如 */5 或 0-30/5)
	if strings.Contains(field, "/") {
		parts := strings.Split(field, "/")
		if len(parts) == 2 {
			step, err := strconv.Atoi(parts[1])
			if err != nil || step <= 0 {
				return fmt.Errorf("%s字段步长值无效: %s", fieldName, field)
			}
			// 验证基础部分
			if parts[0] != "*" {
				return validateSingleCronValue(parts[0], min, max, fieldName)
			}
			return nil
		}
	}

	// 验证单个值
	return validateSingleCronValue(field, min, max, fieldName)
}

// validateSingleCronValue 验证单个数值
func validateSingleCronValue(value string, min, max int, fieldName string) error {
	// 跳过特殊字符
	if value == "*" || value == "?" || value == "L" || value == "W" {
		return nil
	}

	// 处理包含特殊字符的值 (如 L, W, #)
	if strings.Contains(value, "L") || strings.Contains(value, "W") || strings.Contains(value, "#") {
		// 这些特殊字符的验证比较复杂，这里只做基本检查
		return nil
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("%s字段包含非数字值: %s", fieldName, value)
	}

	if num < min || num > max {
		return fmt.Errorf("%s字段值超出有效范围 [%d-%d]: %d", fieldName, min, max, num)
	}

	return nil
}

// getUserFromContext 从上下文获取用户信息
func getUserFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		return userID
	}
	return "system"
}
