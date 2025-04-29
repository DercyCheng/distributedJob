package server

import (
	"context"
	"time"

	"github.com/distributedJob/internal/job"
	"github.com/distributedJob/internal/model/entity"
	pb "github.com/distributedJob/internal/rpc/proto"
	"github.com/distributedJob/pkg/logger"
)

// TaskSchedulerServer implements the TaskScheduler RPC service
type TaskSchedulerServer struct {
	pb.UnimplementedTaskSchedulerServer
	scheduler *job.Scheduler
}

// NewTaskSchedulerServer creates a new TaskSchedulerServer
func NewTaskSchedulerServer(scheduler *job.Scheduler) *TaskSchedulerServer {
	return &TaskSchedulerServer{scheduler: scheduler}
}

// ScheduleTask implements the ScheduleTask RPC method
func (s *TaskSchedulerServer) ScheduleTask(ctx context.Context, req *pb.ScheduleTaskRequest) (*pb.ScheduleTaskResponse, error) {
	// Create a task entity from the request
	task := &entity.Task{
		Name:   req.Name,
		Cron:   req.CronExpression,
		Status: 1, // Enabled
	}

	if req.Handler == "http" {
		task.TaskType = "HTTP"
		// Additional HTTP-specific parameters would be set here
	} else if req.Handler == "grpc" {
		task.TaskType = "GRPC"
		// Additional gRPC-specific parameters would be set here
	}

	// Set retry count
	task.RetryCount = int(req.MaxRetry)

	// Add task to scheduler
	taskID, err := s.scheduler.AddTaskAndStore(task)
	if err != nil {
		logger.Errorf("Failed to schedule task via RPC: %v", err)
		return &pb.ScheduleTaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ScheduleTaskResponse{
		TaskId:  taskID,
		Success: true,
		Message: "Task scheduled successfully",
	}, nil
}

// PauseTask implements the PauseTask RPC method
func (s *TaskSchedulerServer) PauseTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	err := s.scheduler.PauseTask(req.TaskId)
	if err != nil {
		return &pb.TaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.TaskResponse{
		Success: true,
		Message: "Task paused successfully",
	}, nil
}

// ResumeTask implements the ResumeTask RPC method
func (s *TaskSchedulerServer) ResumeTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	err := s.scheduler.ResumeTask(req.TaskId)
	if err != nil {
		return &pb.TaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.TaskResponse{
		Success: true,
		Message: "Task resumed successfully",
	}, nil
}

// GetTaskStatus implements the GetTaskStatus RPC method
func (s *TaskSchedulerServer) GetTaskStatus(ctx context.Context, req *pb.TaskRequest) (*pb.TaskStatusResponse, error) {
	task, err := s.scheduler.GetTaskStatus(req.TaskId)
	if err != nil {
		logger.Errorf("Failed to get task status via RPC: %v", err)
		return nil, err
	}

	// Convert time.Time to string
	var lastExecuteTime, nextExecuteTime string
	if task.LastExecuteTime != nil {
		lastExecuteTime = task.LastExecuteTime.Format(time.RFC3339)
	}
	if task.NextExecuteTime != nil {
		nextExecuteTime = task.NextExecuteTime.Format(time.RFC3339)
	}

	return &pb.TaskStatusResponse{
		TaskId:          req.TaskId,
		Status:          int32(task.Status),
		LastExecuteTime: lastExecuteTime,
		NextExecuteTime: nextExecuteTime,
		RetryCount:      int32(task.RetryCount),
	}, nil
}