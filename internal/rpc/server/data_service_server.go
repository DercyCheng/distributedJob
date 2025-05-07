package server

import (
	"context"
	"time"

	pb "github.com/distributedJob/internal/rpc/proto"
	"github.com/distributedJob/internal/service"
	"github.com/distributedJob/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DataServiceServer implements the DataService gRPC service
type DataServiceServer struct {
	pb.UnimplementedDataServiceServer
	taskService service.TaskService
}

// NewDataServiceServer creates a new DataService server
func NewDataServiceServer(taskService service.TaskService) *DataServiceServer {
	return &DataServiceServer{
		taskService: taskService,
	}
}

// GetTaskHistory retrieves the execution history for a task
func (s *DataServiceServer) GetTaskHistory(ctx context.Context, req *pb.TaskHistoryRequest) (*pb.TaskHistoryResponse, error) {
	// Validate request
	if req.TaskId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid task_id is required")
	}

	// Parse time parameters
	var startTime, endTime time.Time
	var err error

	if req.StartTime != "" {
		startTime, err = time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid start_time format, use RFC3339")
		}
	} else {
		// Default to 7 days ago if not specified
		startTime = time.Now().AddDate(0, 0, -7)
	}

	if req.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid end_time format, use RFC3339")
		}
	} else {
		// Default to now if not specified
		endTime = time.Now()
	}

	// Set default limit and offset if not provided
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// Get task history
	records, total, err := s.taskService.GetTaskRecords(req.TaskId, startTime, endTime, limit, offset)
	if err != nil {
		logger.Error("Failed to get task history", "error", err, "taskID", req.TaskId)
		return &pb.TaskHistoryResponse{
			Success: false,
		}, nil
	}

	// Convert to protobuf response
	pbRecords := make([]*pb.TaskRecord, len(records))
	for i, record := range records {
		executeTime := record.CreateTime.Format(time.RFC3339)

		pbRecords[i] = &pb.TaskRecord{
			Id:          record.ID,
			TaskId:      record.TaskID,
			TaskName:    record.TaskName,
			ExecuteTime: executeTime,
			Success:     record.Success == 1,
			Result:      record.Response,
			Error:       record.Response, // Using Response field since there's no dedicated Error field
		}
	}

	return &pb.TaskHistoryResponse{
		Records: pbRecords,
		Total:   total,
		Success: true,
	}, nil
}

// GetStatistics retrieves statistics for tasks
func (s *DataServiceServer) GetStatistics(ctx context.Context, req *pb.StatisticsRequest) (*pb.StatisticsResponse, error) {
	// Determine time range based on period
	var startTime time.Time
	endTime := time.Now()

	switch req.Period {
	case "daily":
		startTime = time.Now().AddDate(0, 0, -1)
	case "weekly":
		startTime = time.Now().AddDate(0, 0, -7)
	case "monthly":
		startTime = time.Now().AddDate(0, -1, 0)
	default:
		// Default to last 7 days
		startTime = time.Now().AddDate(0, 0, -7)
	}

	// Get statistics
	stats, err := s.taskService.GetTaskStatistics(req.DepartmentId, startTime, endTime)
	if err != nil {
		logger.Error("Failed to get task statistics", "error", err, "departmentID", req.DepartmentId)
		return &pb.StatisticsResponse{}, nil
	}

	// Prepare execution stats map
	executionStats := make(map[string]float32)
	for k, v := range stats.ExecutionStats {
		executionStats[k] = float32(v)
	}

	return &pb.StatisticsResponse{
		TaskCount:        int32(stats.TaskCount),
		SuccessRate:      float32(stats.SuccessRate),
		AvgExecutionTime: float32(stats.AvgExecutionTime),
		ExecutionStats:   executionStats,
	}, nil
}
