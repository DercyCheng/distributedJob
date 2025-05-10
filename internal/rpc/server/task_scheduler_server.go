package server

import (
	"context"
	"encoding/json"
	"time"

	"distributedJob/internal/job"
	"distributedJobodel/entity"
	pb "distributedJobpc/proto"
	"distributedJob"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TaskSchedulerServer implements the TaskScheduler gRPC service
type TaskSchedulerServer struct {
	pb.UnimplementedTaskSchedulerServer
	scheduler *job.Scheduler
}

// NewTaskSchedulerServer creates a new TaskScheduler service server
func NewTaskSchedulerServer(scheduler *job.Scheduler) *TaskSchedulerServer {
	return &TaskSchedulerServer{
		scheduler: scheduler,
	}
}

// ScheduleTask handles task scheduling requests
func (s *TaskSchedulerServer) ScheduleTask(ctx context.Context, req *pb.ScheduleTaskRequest) (*pb.ScheduleTaskResponse, error) {
	// Validate request
	if req.Name == "" || req.CronExpression == "" || req.Handler == "" {
		return nil, status.Error(codes.InvalidArgument, "name, cron_expression, and handler are required")
	}

	// Create task entity
	task := &entity.Task{
		Name:       req.Name,
		Cron:       req.CronExpression,
		TaskType:   req.Handler, // Handler represents task type (HTTP or GRPC)
		Status:     1,           // Enabled by default
		RetryCount: int(req.MaxRetry),
		CreateTime: time.Now(),
	}

	// Set task parameters based on handler type
	if req.Handler == "HTTP" {
		// Parse HTTP parameters from params
		var httpParams map[string]interface{}
		if err := json.Unmarshal(req.Params, &httpParams); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid HTTP parameters format")
		}

		// Set HTTP specific fields
		if url, ok := httpParams["url"].(string); ok {
			task.URL = url
		} else {
			return nil, status.Error(codes.InvalidArgument, "URL is required for HTTP tasks")
		}

		if method, ok := httpParams["method"].(string); ok {
			task.HTTPMethod = method
		} else {
			task.HTTPMethod = "GET" // Default to GET if not specified
		}

		// Optional parameters
		if body, ok := httpParams["body"].(string); ok {
			task.Body = body
		}

		if headers, ok := httpParams["headers"].(string); ok {
			task.Headers = headers
		}

		if fallbackURL, ok := httpParams["fallback_url"].(string); ok {
			task.FallbackURL = fallbackURL
		}
	} else if req.Handler == "GRPC" {
		// Parse gRPC parameters from params
		var grpcParams map[string]interface{}
		if err := json.Unmarshal(req.Params, &grpcParams); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid gRPC parameters format")
		}

		// Set gRPC specific fields
		if service, ok := grpcParams["service"].(string); ok {
			task.GrpcService = service
		} else {
			return nil, status.Error(codes.InvalidArgument, "service is required for gRPC tasks")
		}

		if method, ok := grpcParams["method"].(string); ok {
			task.GrpcMethod = method
		} else {
			return nil, status.Error(codes.InvalidArgument, "method is required for gRPC tasks")
		}

		// Optional parameters
		if params, ok := grpcParams["params"].(string); ok {
			task.GrpcParams = params
		}

		if fallbackService, ok := grpcParams["fallback_service"].(string); ok {
			task.FallbackGrpcService = fallbackService
		}

		if fallbackMethod, ok := grpcParams["fallback_method"].(string); ok {
			task.FallbackGrpcMethod = fallbackMethod
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "handler must be either HTTP or GRPC")
	}

	// Store and schedule the task
	taskID, err := s.scheduler.AddTaskAndStore(task)
	if err != nil {
		logger.Error("Failed to schedule task", "error", err, "taskName", req.Name)
		return &pb.ScheduleTaskResponse{
			Success: false,
			Message: "Failed to schedule task: " + err.Error(),
		}, nil
	}

	// Return successful response
	return &pb.ScheduleTaskResponse{
		TaskId:  taskID,
		Success: true,
		Message: "Task scheduled successfully",
	}, nil
}

// PauseTask pauses a running task
func (s *TaskSchedulerServer) PauseTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	// Validate request
	if req.TaskId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid task_id is required")
	}

	// Pause the task
	err := s.scheduler.PauseTask(req.TaskId)
	if err != nil {
		logger.Error("Failed to pause task", "error", err, "taskID", req.TaskId)
		return &pb.TaskResponse{
			Success: false,
			Message: "Failed to pause task: " + err.Error(),
		}, nil
	}

	// Return successful response
	return &pb.TaskResponse{
		Success: true,
		Message: "Task paused successfully",
	}, nil
}

// ResumeTask resumes a paused task
func (s *TaskSchedulerServer) ResumeTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	// Validate request
	if req.TaskId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid task_id is required")
	}

	// Resume the task
	err := s.scheduler.ResumeTask(req.TaskId)
	if err != nil {
		logger.Error("Failed to resume task", "error", err, "taskID", req.TaskId)
		return &pb.TaskResponse{
			Success: false,
			Message: "Failed to resume task: " + err.Error(),
		}, nil
	}

	// Return successful response
	return &pb.TaskResponse{
		Success: true,
		Message: "Task resumed successfully",
	}, nil
}

// GetTaskStatus gets the status of a task
func (s *TaskSchedulerServer) GetTaskStatus(ctx context.Context, req *pb.TaskRequest) (*pb.TaskStatusResponse, error) {
	// Validate request
	if req.TaskId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid task_id is required")
	}

	// Get task status
	task, err := s.scheduler.GetTaskStatus(req.TaskId)
	if err != nil {
		logger.Error("Failed to get task status", "error", err, "taskID", req.TaskId)
		return nil, status.Error(codes.NotFound, "task not found or error retrieving status")
	}

	// Format last execute time
	var lastExecuteTime string
	if task.LastExecuteTime != nil {
		lastExecuteTime = task.LastExecuteTime.Format(time.RFC3339)
	}

	// Format next execute time
	var nextExecuteTime string
	if task.NextExecuteTime != nil {
		nextExecuteTime = task.NextExecuteTime.Format(time.RFC3339)
	}

	// Return task status
	return &pb.TaskStatusResponse{
		TaskId:          task.ID,
		Status:          int32(task.Status),
		LastExecuteTime: lastExecuteTime,
		NextExecuteTime: nextExecuteTime,
		RetryCount:      int32(task.RetryCount),
	}, nil
}
