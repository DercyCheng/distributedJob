package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	grpc "go-job/api/grpc"
	"go-job/internal/models"
	"go-job/pkg/config"
	"go-job/pkg/logger"
)

type AISchedulerService struct {
	grpc.UnimplementedAISchedulerServiceServer
	db     *gorm.DB
	config *config.Config
	client *http.Client
}

type DashScopeRequest struct {
	Model      string              `json:"model"`
	Input      DashScopeInput      `json:"input"`
	Parameters DashScopeParameters `json:"parameters"`
}

type DashScopeInput struct {
	Messages []DashScopeMessage `json:"messages"`
}

type DashScopeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DashScopeParameters struct {
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

type DashScopeResponse struct {
	Output struct {
		Text    string `json:"text"`
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
}

func NewAISchedulerService(db *gorm.DB, config *config.Config) *AISchedulerService {
	return &AISchedulerService{
		db:     db,
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *AISchedulerService) AnalyzeJob(ctx context.Context, req *grpc.AnalyzeJobRequest) (*grpc.AnalyzeJobResponse, error) {
	// 获取任务信息
	var job models.Job
	err := s.db.Where("id = ?", req.GetJobId()).First(&job).Error
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "job not found: %v", err)
	}

	// 获取任务执行历史
	var executions []models.JobExecution
	s.db.Where("job_id = ?", job.ID).
		Order("created_at DESC").
		Limit(20).
		Find(&executions)

	// 构建分析上下文
	context := s.buildJobAnalysisContext(&job, executions, req.GetContext())

	// 调用阿里云百炼API
	aiResponse, err := s.callDashScopeAPI(context, "分析这个定时任务的性能、可靠性和优化建议")
	if err != nil {
		logger.WithError(err).Error("Failed to call DashScope API")
		return nil, status.Errorf(codes.Internal, "AI analysis failed: %v", err)
	}

	// 保存AI分析结果
	aiSchedule := &models.AISchedule{
		JobID:          job.ID,
		PromptTemplate: "job_analysis",
		AIResponse:     aiResponse,
		Strategy:       "performance_analysis",
		Priority:       s.calculatePriority(aiResponse),
		Context:        context,
	}
	s.db.Create(aiSchedule)

	// 解析AI响应，提取建议
	recommendations := s.extractRecommendations(aiResponse)

	return &grpc.AnalyzeJobResponse{
		Analysis:        aiResponse,
		Strategy:        "ai_optimized",
		Priority:        int32(aiSchedule.Priority),
		Recommendations: recommendations,
	}, nil
}

func (s *AISchedulerService) OptimizeSchedule(ctx context.Context, req *grpc.OptimizeScheduleRequest) (*grpc.OptimizeScheduleResponse, error) {
	jobIDs := req.GetJobIds()
	goal := req.GetOptimizationGoal()

	if len(jobIDs) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "job_ids cannot be empty")
	}

	// 获取任务信息
	var jobs []models.Job
	err := s.db.Where("id IN ?", jobIDs).Find(&jobs).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch jobs: %v", err)
	}

	// 获取系统负载信息
	systemContext := s.buildSystemContext()

	// 构建优化提示
	prompt := s.buildOptimizationPrompt(jobs, goal, systemContext)

	// 调用AI服务
	aiResponse, err := s.callDashScopeAPI(prompt, "基于系统负载和任务特性，优化这些任务的调度时间")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "AI optimization failed: %v", err)
	}

	// 解析AI建议
	optimizations := s.parseOptimizationResponse(aiResponse, jobs)

	return &grpc.OptimizeScheduleResponse{
		Optimizations: optimizations,
		Summary:       aiResponse,
	}, nil
}

func (s *AISchedulerService) GetAIRecommendations(ctx context.Context, req *grpc.GetAIRecommendationsRequest) (*grpc.GetAIRecommendationsResponse, error) {
	recType := req.GetType()
	context := req.GetContext()

	var prompt string
	switch recType {
	case "performance":
		prompt = s.buildPerformanceRecommendationPrompt(context)
	case "reliability":
		prompt = s.buildReliabilityRecommendationPrompt(context)
	case "cost":
		prompt = s.buildCostRecommendationPrompt(context)
	default:
		prompt = s.buildGeneralRecommendationPrompt(context)
	}

	aiResponse, err := s.callDashScopeAPI(prompt, "提供系统优化建议")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get AI recommendations: %v", err)
	}

	recommendations := s.parseRecommendations(aiResponse, recType)

	return &grpc.GetAIRecommendationsResponse{
		Recommendations: recommendations,
	}, nil
}

func (s *AISchedulerService) callDashScopeAPI(context, userPrompt string) (string, error) {
	// 阿里云百炼API配置
	apiURL := "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	apiKey := s.config.AI.DashScopeAPIKey // 需要在配置中添加

	// 构建请求
	request := DashScopeRequest{
		Model: "qwen-plus", // 或其他模型
		Input: DashScopeInput{
			Messages: []DashScopeMessage{
				{
					Role:    "system",
					Content: "你是一个专业的任务调度系统分析师，擅长分析定时任务的性能、可靠性和优化策略。请基于提供的数据进行分析并给出具体的优化建议。",
				},
				{
					Role:    "user",
					Content: fmt.Sprintf("上下文数据：\n%s\n\n问题：%s", context, userPrompt),
				},
			},
		},
		Parameters: DashScopeParameters{
			Temperature: 0.7,
			MaxTokens:   2000,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response DashScopeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 提取内容
	if len(response.Output.Choices) > 0 {
		return response.Output.Choices[0].Message.Content, nil
	}

	return response.Output.Text, nil
}

func (s *AISchedulerService) buildJobAnalysisContext(job *models.Job, executions []models.JobExecution, additionalContext string) string {
	var context strings.Builder

	context.WriteString(fmt.Sprintf("任务信息：\n"))
	context.WriteString(fmt.Sprintf("- 名称：%s\n", job.Name))
	context.WriteString(fmt.Sprintf("- 描述：%s\n", job.Description))
	context.WriteString(fmt.Sprintf("- Cron表达式：%s\n", job.Cron))
	context.WriteString(fmt.Sprintf("- 命令：%s\n", job.Command))
	context.WriteString(fmt.Sprintf("- 超时时间：%d秒\n", job.Timeout))
	context.WriteString(fmt.Sprintf("- 重试次数：%d\n", job.RetryAttempts))
	context.WriteString(fmt.Sprintf("- 优先级：%d\n", job.Priority))

	context.WriteString(fmt.Sprintf("\n执行历史（最近%d次）：\n", len(executions)))
	successCount := 0
	failCount := 0
	var totalDuration time.Duration

	for i, exec := range executions {
		status := string(exec.Status)
		duration := "未知"
		if exec.StartedAt != nil && exec.FinishedAt != nil {
			d := exec.FinishedAt.Sub(*exec.StartedAt)
			duration = d.String()
			totalDuration += d
		}

		context.WriteString(fmt.Sprintf("  %d. 状态：%s, 耗时：%s, 退出码：%d\n",
			i+1, status, duration, exec.ExitCode))

		if exec.Status == models.ExecutionStatusSuccess {
			successCount++
		} else if exec.Status == models.ExecutionStatusFailed {
			failCount++
		}
	}

	if len(executions) > 0 {
		successRate := float64(successCount) / float64(len(executions)) * 100
		avgDuration := totalDuration / time.Duration(len(executions))
		context.WriteString(fmt.Sprintf("\n统计信息：\n"))
		context.WriteString(fmt.Sprintf("- 成功率：%.2f%%\n", successRate))
		context.WriteString(fmt.Sprintf("- 平均执行时间：%s\n", avgDuration))
		context.WriteString(fmt.Sprintf("- 失败次数：%d\n", failCount))
	}

	if additionalContext != "" {
		context.WriteString(fmt.Sprintf("\n补充信息：%s\n", additionalContext))
	}

	return context.String()
}

func (s *AISchedulerService) buildSystemContext() string {
	var context strings.Builder

	// 获取系统统计
	var totalJobs, activeJobs int64
	s.db.Model(&models.Job{}).Count(&totalJobs)
	s.db.Model(&models.Job{}).Where("enabled = ?", true).Count(&activeJobs)

	var totalWorkers, onlineWorkers int64
	s.db.Model(&models.Worker{}).Count(&totalWorkers)
	s.db.Model(&models.Worker{}).Where("status = ?", models.WorkerStatusOnline).Count(&onlineWorkers)

	context.WriteString(fmt.Sprintf("系统状态：\n"))
	context.WriteString(fmt.Sprintf("- 总任务数：%d\n", totalJobs))
	context.WriteString(fmt.Sprintf("- 活跃任务数：%d\n", activeJobs))
	context.WriteString(fmt.Sprintf("- 总工作节点：%d\n", totalWorkers))
	context.WriteString(fmt.Sprintf("- 在线工作节点：%d\n", onlineWorkers))

	// 获取最近的执行统计
	var recentExecutions []models.JobExecution
	s.db.Where("created_at > ?", time.Now().Add(-24*time.Hour)).Find(&recentExecutions)

	successCount := 0
	for _, exec := range recentExecutions {
		if exec.Status == models.ExecutionStatusSuccess {
			successCount++
		}
	}

	if len(recentExecutions) > 0 {
		successRate := float64(successCount) / float64(len(recentExecutions)) * 100
		context.WriteString(fmt.Sprintf("- 24小时成功率：%.2f%%\n", successRate))
		context.WriteString(fmt.Sprintf("- 24小时执行次数：%d\n", len(recentExecutions)))
	}

	return context.String()
}

func (s *AISchedulerService) buildOptimizationPrompt(jobs []models.Job, goal, systemContext string) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("优化目标：%s\n\n", goal))
	prompt.WriteString(systemContext)
	prompt.WriteString("\n待优化任务：\n")

	for i, job := range jobs {
		prompt.WriteString(fmt.Sprintf("%d. %s (Cron: %s, 超时: %ds, 优先级: %d)\n",
			i+1, job.Name, job.Cron, job.Timeout, job.Priority))
	}

	prompt.WriteString("\n请分析这些任务的调度冲突和资源竞争情况，并提供优化建议。")

	return prompt.String()
}

func (s *AISchedulerService) buildPerformanceRecommendationPrompt(context map[string]string) string {
	return "请基于系统性能数据，提供性能优化建议。"
}

func (s *AISchedulerService) buildReliabilityRecommendationPrompt(context map[string]string) string {
	return "请基于系统可靠性数据，提供可靠性改进建议。"
}

func (s *AISchedulerService) buildCostRecommendationPrompt(context map[string]string) string {
	return "请基于系统资源使用情况，提供成本优化建议。"
}

func (s *AISchedulerService) buildGeneralRecommendationPrompt(context map[string]string) string {
	return "请基于系统整体状况，提供综合优化建议。"
}

func (s *AISchedulerService) calculatePriority(aiResponse string) int {
	// 简单的优先级计算逻辑
	if strings.Contains(strings.ToLower(aiResponse), "critical") ||
		strings.Contains(strings.ToLower(aiResponse), "urgent") {
		return 5
	}
	if strings.Contains(strings.ToLower(aiResponse), "important") {
		return 4
	}
	if strings.Contains(strings.ToLower(aiResponse), "medium") {
		return 3
	}
	return 2
}

func (s *AISchedulerService) extractRecommendations(aiResponse string) []string {
	// 简单的建议提取逻辑
	var recommendations []string

	// 按行分割响应
	lines := strings.Split(aiResponse, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "建议") ||
			strings.HasPrefix(line, "推荐") ||
			strings.HasPrefix(line, "-") ||
			strings.HasPrefix(line, "*") {
			recommendations = append(recommendations, line)
		}
	}

	if len(recommendations) == 0 {
		// 如果没有找到明确的建议，返回默认建议
		recommendations = []string{
			"监控任务执行时间，适当调整超时设置",
			"定期检查任务执行成功率",
			"考虑在低负载时段调度资源密集型任务",
		}
	}

	return recommendations
}

func (s *AISchedulerService) parseOptimizationResponse(aiResponse string, jobs []models.Job) []*grpc.ScheduleOptimization {
	var optimizations []*grpc.ScheduleOptimization

	// 为每个任务生成一个优化建议（简化处理）
	for _, job := range jobs {
		optimization := &grpc.ScheduleOptimization{
			JobId:           job.ID,
			RecommendedCron: job.Cron, // 暂时保持原样
			Priority:        int32(job.Priority),
			Reason:          "基于AI分析的优化建议",
		}
		optimizations = append(optimizations, optimization)
	}

	return optimizations
}

func (s *AISchedulerService) parseRecommendations(aiResponse, recType string) []*grpc.AIRecommendation {
	var recommendations []*grpc.AIRecommendation

	// 简单的建议解析
	lines := strings.Split(aiResponse, "\n")
	priority := 3

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "建议")) {
			recommendation := &grpc.AIRecommendation{
				Type:        recType,
				Title:       "AI优化建议",
				Description: line,
				Action:      "review_and_implement",
				Priority:    int32(priority),
			}
			recommendations = append(recommendations, recommendation)

			if priority > 1 {
				priority--
			}
		}
	}

	if len(recommendations) == 0 {
		// 默认建议
		recommendations = []*grpc.AIRecommendation{
			{
				Type:        recType,
				Title:       "系统监控建议",
				Description: "建议定期监控系统性能指标",
				Action:      "setup_monitoring",
				Priority:    3,
			},
		}
	}

	return recommendations
}
