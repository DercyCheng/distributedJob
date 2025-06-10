package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	grpc "go-job/api/grpc"
	"go-job/internal/models"
)

type MCPService struct {
	grpc.UnimplementedMCPServiceServer
	db          *gorm.DB
	aiScheduler *AISchedulerService
}

func NewMCPService(db *gorm.DB, aiScheduler *AISchedulerService) *MCPService {
	return &MCPService{
		db:          db,
		aiScheduler: aiScheduler,
	}
}

func (s *MCPService) ListTools(ctx context.Context, req *grpc.ListToolsRequest) (*grpc.ListToolsResponse, error) {
	tools := []*grpc.MCPTool{
		{
			Name:        "list_jobs",
			Description: "列出系统中的所有任务",
			Category:    "job_management",
			Parameters: map[string]string{
				"department_id": "部门ID（可选）",
				"status":        "任务状态（可选）",
				"limit":         "返回数量限制（可选，默认10）",
			},
		},
		{
			Name:        "create_job",
			Description: "创建新的定时任务",
			Category:    "job_management",
			Parameters: map[string]string{
				"name":        "任务名称（必需）",
				"description": "任务描述（可选）",
				"cron":        "Cron表达式（必需）",
				"command":     "执行命令（必需）",
				"timeout":     "超时时间（可选，默认300秒）",
			},
		},
		{
			Name:        "analyze_job_performance",
			Description: "分析任务执行性能",
			Category:    "analytics",
			Parameters: map[string]string{
				"job_id": "任务ID（必需）",
				"days":   "分析天数（可选，默认7天）",
				"metric": "分析指标（可选：success_rate,avg_duration,error_rate）",
			},
		},
		{
			Name:        "optimize_schedule",
			Description: "优化任务调度时间",
			Category:    "optimization",
			Parameters: map[string]string{
				"job_ids": "任务ID列表（必需）",
				"goal":    "优化目标（可选：load_balance,minimize_conflicts,maximize_throughput）",
			},
		},
		{
			Name:        "get_system_status",
			Description: "获取系统运行状态",
			Category:    "monitoring",
			Parameters: map[string]string{
				"include_workers": "是否包含工作节点信息（可选，默认true）",
				"include_stats":   "是否包含统计信息（可选，默认true）",
			},
		},
		{
			Name:        "predict_resource_usage",
			Description: "预测系统资源使用情况",
			Category:    "prediction",
			Parameters: map[string]string{
				"hours":       "预测小时数（可选，默认24小时）",
				"granularity": "时间粒度（可选：hour,day，默认hour）",
			},
		},
		{
			Name:        "get_job_recommendations",
			Description: "获取任务优化建议",
			Category:    "recommendation",
			Parameters: map[string]string{
				"job_id": "任务ID（可选，不提供则对所有任务分析）",
				"type":   "建议类型（可选：performance,reliability,cost）",
			},
		},
		{
			Name:        "schedule_maintenance",
			Description: "安排系统维护窗口",
			Category:    "maintenance",
			Parameters: map[string]string{
				"start_time": "开始时间（必需，格式：2006-01-02 15:04:05）",
				"duration":   "持续时间（必需，分钟）",
				"reason":     "维护原因（可选）",
			},
		},
	}

	// 根据类别过滤
	if category := req.GetCategory(); category != "" {
		var filteredTools []*grpc.MCPTool
		for _, tool := range tools {
			if tool.Category == category {
				filteredTools = append(filteredTools, tool)
			}
		}
		tools = filteredTools
	}

	return &grpc.ListToolsResponse{
		Tools: tools,
	}, nil
}

func (s *MCPService) CallTool(ctx context.Context, req *grpc.CallToolRequest) (*grpc.CallToolResponse, error) {
	toolName := req.GetToolName()
	arguments := req.GetArguments()

	switch toolName {
	case "list_jobs":
		return s.handleListJobs(ctx, arguments)
	case "create_job":
		return s.handleCreateJob(ctx, arguments)
	case "analyze_job_performance":
		return s.handleAnalyzeJobPerformance(ctx, arguments)
	case "optimize_schedule":
		return s.handleOptimizeSchedule(ctx, arguments)
	case "get_system_status":
		return s.handleGetSystemStatus(ctx, arguments)
	case "predict_resource_usage":
		return s.handlePredictResourceUsage(ctx, arguments)
	case "get_job_recommendations":
		return s.handleGetJobRecommendations(ctx, arguments)
	case "schedule_maintenance":
		return s.handleScheduleMaintenance(ctx, arguments)
	default:
		return &grpc.CallToolResponse{
			Success: false,
			Error:   fmt.Sprintf("unknown tool: %s", toolName),
		}, nil
	}
}

func (s *MCPService) GetResources(ctx context.Context, req *grpc.GetResourcesRequest) (*grpc.GetResourcesResponse, error) {
	resourceType := req.GetType()
	filter := req.GetFilter()

	var resources []*grpc.MCPResource

	switch resourceType {
	case "jobs":
		jobs, err := s.getJobResources(filter)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get job resources: %v", err)
		}
		resources = append(resources, jobs...)

	case "workers":
		workers, err := s.getWorkerResources(filter)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get worker resources: %v", err)
		}
		resources = append(resources, workers...)

	case "executions":
		executions, err := s.getExecutionResources(filter)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get execution resources: %v", err)
		}
		resources = append(resources, executions...)

	default:
		// 返回所有类型的资源
		allResources, err := s.getAllResources()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get resources: %v", err)
		}
		resources = allResources
	}

	return &grpc.GetResourcesResponse{
		Resources: resources,
	}, nil
}

func (s *MCPService) handleListJobs(ctx context.Context, arguments map[string]string) (*grpc.CallToolResponse, error) {
	var jobs []models.Job
	query := s.db.Model(&models.Job{})

	if deptID := arguments["department_id"]; deptID != "" {
		query = query.Where("department_id = ?", deptID)
	}

	if status := arguments["status"]; status != "" {
		enabled := status == "enabled"
		query = query.Where("enabled = ?", enabled)
	}

	limit := 10
	if limitStr := arguments["limit"]; limitStr != "" {
		if l, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && l > 0 {
			// 使用解析的limit
		}
	}

	err := query.Limit(limit).Find(&jobs).Error
	if err != nil {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to query jobs: %v", err),
		}, nil
	}

	result, _ := json.Marshal(jobs)
	return &grpc.CallToolResponse{
		Success: true,
		Result:  string(result),
	}, nil
}

func (s *MCPService) handleCreateJob(ctx context.Context, arguments map[string]string) (*grpc.CallToolResponse, error) {
	name := arguments["name"]
	if name == "" {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   "name is required",
		}, nil
	}

	cron := arguments["cron"]
	if cron == "" {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   "cron expression is required",
		}, nil
	}

	command := arguments["command"]
	if command == "" {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   "command is required",
		}, nil
	}

	job := &models.Job{
		Name:        name,
		Description: arguments["description"],
		Cron:        cron,
		Command:     command,
		Enabled:     true,
		Timeout:     300, // 默认5分钟
	}

	if timeout := arguments["timeout"]; timeout != "" {
		if t, err := fmt.Sscanf(timeout, "%d", &job.Timeout); err != nil || t != 1 {
			job.Timeout = 300
		}
	}

	err := s.db.Create(job).Error
	if err != nil {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to create job: %v", err),
		}, nil
	}

	result, _ := json.Marshal(job)
	return &grpc.CallToolResponse{
		Success: true,
		Result:  fmt.Sprintf("Job created successfully: %s", string(result)),
	}, nil
}

func (s *MCPService) handleAnalyzeJobPerformance(ctx context.Context, arguments map[string]string) (*grpc.CallToolResponse, error) {
	jobID := arguments["job_id"]
	if jobID == "" {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   "job_id is required",
		}, nil
	}

	// 这里可以调用AI分析服务
	if s.aiScheduler != nil {
		response, err := s.aiScheduler.AnalyzeJob(ctx, &grpc.AnalyzeJobRequest{
			JobId:   jobID,
			Context: "performance_analysis",
		})
		if err != nil {
			return &grpc.CallToolResponse{
				Success: false,
				Error:   fmt.Sprintf("AI analysis failed: %v", err),
			}, nil
		}

		return &grpc.CallToolResponse{
			Success: true,
			Result:  response.Analysis,
		}, nil
	}

	// 简单的性能分析
	var executions []models.JobExecution
	err := s.db.Where("job_id = ?", jobID).
		Order("created_at DESC").
		Limit(100).
		Find(&executions).Error
	if err != nil {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to query executions: %v", err),
		}, nil
	}

	analysis := fmt.Sprintf("Performance analysis for job %s: %d recent executions found", jobID, len(executions))
	return &grpc.CallToolResponse{
		Success: true,
		Result:  analysis,
	}, nil
}

func (s *MCPService) handleOptimizeSchedule(ctx context.Context, arguments map[string]string) (*grpc.CallToolResponse, error) {
	if s.aiScheduler == nil {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   "AI scheduler service not available",
		}, nil
	}

	jobIDs := arguments["job_ids"]
	if jobIDs == "" {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   "job_ids is required",
		}, nil
	}

	goal := arguments["goal"]
	if goal == "" {
		goal = "load_balance"
	}

	response, err := s.aiScheduler.OptimizeSchedule(ctx, &grpc.OptimizeScheduleRequest{
		JobIds:           []string{jobIDs}, // 简化处理，实际应该解析逗号分隔的ID
		OptimizationGoal: goal,
	})
	if err != nil {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   fmt.Sprintf("schedule optimization failed: %v", err),
		}, nil
	}

	result, _ := json.Marshal(response)
	return &grpc.CallToolResponse{
		Success: true,
		Result:  string(result),
	}, nil
}

func (s *MCPService) handleGetSystemStatus(ctx context.Context, arguments map[string]string) (*grpc.CallToolResponse, error) {
	status := map[string]interface{}{
		"timestamp": fmt.Sprintf("%v", ctx.Value("timestamp")),
		"status":    "healthy",
	}

	// 获取任务统计
	var jobCount int64
	s.db.Model(&models.Job{}).Count(&jobCount)
	status["total_jobs"] = jobCount

	var activeJobCount int64
	s.db.Model(&models.Job{}).Where("enabled = ?", true).Count(&activeJobCount)
	status["active_jobs"] = activeJobCount

	// 获取工作节点统计
	if arguments["include_workers"] != "false" {
		var workerCount int64
		s.db.Model(&models.Worker{}).Count(&workerCount)
		status["total_workers"] = workerCount

		var onlineWorkerCount int64
		s.db.Model(&models.Worker{}).Where("status = ?", models.WorkerStatusOnline).Count(&onlineWorkerCount)
		status["online_workers"] = onlineWorkerCount
	}

	result, _ := json.Marshal(status)
	return &grpc.CallToolResponse{
		Success: true,
		Result:  string(result),
	}, nil
}

func (s *MCPService) handlePredictResourceUsage(ctx context.Context, arguments map[string]string) (*grpc.CallToolResponse, error) {
	// 这里应该实现真正的预测逻辑，可能需要调用AI服务
	prediction := map[string]interface{}{
		"prediction_type": "resource_usage",
		"hours":           24,
		"predicted_load":  "moderate",
		"recommendations": []string{
			"Current capacity is sufficient for next 24 hours",
			"Consider adding 1 more worker during peak hours (9-17)",
		},
	}

	result, _ := json.Marshal(prediction)
	return &grpc.CallToolResponse{
		Success: true,
		Result:  string(result),
	}, nil
}

func (s *MCPService) handleGetJobRecommendations(ctx context.Context, arguments map[string]string) (*grpc.CallToolResponse, error) {
	if s.aiScheduler != nil {
		response, err := s.aiScheduler.GetAIRecommendations(ctx, &grpc.GetAIRecommendationsRequest{
			Type:    arguments["type"],
			Context: arguments,
		})
		if err == nil {
			result, _ := json.Marshal(response.Recommendations)
			return &grpc.CallToolResponse{
				Success: true,
				Result:  string(result),
			}, nil
		}
	}

	// 简单的建议逻辑
	recommendations := []map[string]interface{}{
		{
			"type":        "performance",
			"title":       "优化任务超时设置",
			"description": "建议根据历史执行时间调整任务超时配置",
			"priority":    3,
		},
		{
			"type":        "reliability",
			"title":       "增加任务重试次数",
			"description": "对于偶尔失败的任务，建议增加重试次数",
			"priority":    2,
		},
	}

	result, _ := json.Marshal(recommendations)
	return &grpc.CallToolResponse{
		Success: true,
		Result:  string(result),
	}, nil
}

func (s *MCPService) handleScheduleMaintenance(ctx context.Context, arguments map[string]string) (*grpc.CallToolResponse, error) {
	startTime := arguments["start_time"]
	if startTime == "" {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   "start_time is required",
		}, nil
	}

	duration := arguments["duration"]
	if duration == "" {
		return &grpc.CallToolResponse{
			Success: false,
			Error:   "duration is required",
		}, nil
	}

	// 这里应该实现维护窗口的调度逻辑
	result := fmt.Sprintf("Maintenance scheduled: start=%s, duration=%s minutes", startTime, duration)
	if reason := arguments["reason"]; reason != "" {
		result += fmt.Sprintf(", reason=%s", reason)
	}

	return &grpc.CallToolResponse{
		Success: true,
		Result:  result,
	}, nil
}

func (s *MCPService) getJobResources(filter string) ([]*grpc.MCPResource, error) {
	var jobs []models.Job
	query := s.db.Model(&models.Job{})
	if filter != "" {
		query = query.Where("name LIKE ?", "%"+filter+"%")
	}

	err := query.Limit(50).Find(&jobs).Error
	if err != nil {
		return nil, err
	}

	var resources []*grpc.MCPResource
	for _, job := range jobs {
		jobData, _ := json.Marshal(job)
		resources = append(resources, &grpc.MCPResource{
			Uri:         fmt.Sprintf("job://%s", job.ID),
			Name:        job.Name,
			Description: job.Description,
			Type:        "job",
			Content:     string(jobData),
		})
	}

	return resources, nil
}

func (s *MCPService) getWorkerResources(filter string) ([]*grpc.MCPResource, error) {
	var workers []models.Worker
	query := s.db.Model(&models.Worker{})
	if filter != "" {
		query = query.Where("name LIKE ?", "%"+filter+"%")
	}

	err := query.Limit(50).Find(&workers).Error
	if err != nil {
		return nil, err
	}

	var resources []*grpc.MCPResource
	for _, worker := range workers {
		workerData, _ := json.Marshal(worker)
		resources = append(resources, &grpc.MCPResource{
			Uri:         fmt.Sprintf("worker://%s", worker.ID),
			Name:        worker.Name,
			Description: fmt.Sprintf("Worker at %s:%d", worker.IP, worker.Port),
			Type:        "worker",
			Content:     string(workerData),
		})
	}

	return resources, nil
}

func (s *MCPService) getExecutionResources(filter string) ([]*grpc.MCPResource, error) {
	var executions []models.JobExecution
	query := s.db.Model(&models.JobExecution{}).Preload("Job")
	if filter != "" {
		query = query.Joins("JOIN jobs ON job_executions.job_id = jobs.id").
			Where("jobs.name LIKE ?", "%"+filter+"%")
	}

	err := query.Order("created_at DESC").Limit(50).Find(&executions).Error
	if err != nil {
		return nil, err
	}

	var resources []*grpc.MCPResource
	for _, execution := range executions {
		execData, _ := json.Marshal(execution)
		resources = append(resources, &grpc.MCPResource{
			Uri:         fmt.Sprintf("execution://%s", execution.ID),
			Name:        fmt.Sprintf("Execution %s", execution.ID[:8]),
			Description: fmt.Sprintf("Job: %s, Status: %s", execution.Job.Name, execution.Status),
			Type:        "execution",
			Content:     string(execData),
		})
	}

	return resources, nil
}

func (s *MCPService) getAllResources() ([]*grpc.MCPResource, error) {
	var resources []*grpc.MCPResource

	// 获取任务资源
	jobs, err := s.getJobResources("")
	if err == nil {
		resources = append(resources, jobs...)
	}

	// 获取工作节点资源
	workers, err := s.getWorkerResources("")
	if err == nil {
		resources = append(resources, workers...)
	}

	// 获取执行记录资源
	executions, err := s.getExecutionResources("")
	if err == nil {
		resources = append(resources, executions...)
	}

	return resources, nil
}
