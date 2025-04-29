package server

import (
	"context"
	"time"

	"github.com/distributedJob/internal/service"
	pb "github.com/distributedJob/internal/rpc/proto"
	"github.com/distributedJob/pkg/logger"
)

// DataServiceServer implements the DataService RPC service
type DataServiceServer struct {
	pb.UnimplementedDataServiceServer
	taskService service.TaskService
}

// NewDataServiceServer creates a new DataServiceServer
func NewDataServiceServer(taskService service.TaskService) *DataServiceServer {
	return &DataServiceServer{taskService: taskService}
}

// GetTaskHistory implements the GetTaskHistory RPC method
func (s *DataServiceServer) GetTaskHistory(ctx context.Context, req *pb.TaskHistoryRequest) (*pb.TaskHistoryResponse, error) {
	// Parse time strings to time.Time
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		logger.Errorf("Failed to parse start time: %v", err)
		return &pb.TaskHistoryResponse{
			Success: false,
		}, nil
	}
	
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		logger.Errorf("Failed to parse end time: %v", err)
		return &pb.TaskHistoryResponse{
			Success: false,
		}, nil
	}
	
	// Extract year and month for record lookup
	year, month, _ := startTime.Date()
	
	// Get task ID
	taskID := req.TaskId
	
	// Retrieve records from service
	records, total, err := s.taskService.GetRecordListByTimeRange(
		year, 
		int(month),
		&taskID, 
		nil, // departmentID 
		nil, // success
		int(req.Offset/req.Limit + 1), // page
		int(req.Limit), // size
		startTime,
		endTime,
	)
	
	if err != nil {
		logger.Errorf("Failed to get task history via RPC: %v", err)
		return &pb.TaskHistoryResponse{
			Success: false,
		}, nil
	}
	
	// Convert records to the response format
	var protoRecords []*pb.TaskRecord
	for _, record := range records {
		// Create task record response
		taskRecord := &pb.TaskRecord{
			Id:          record.ID,
			TaskId:      record.TaskID,
			TaskName:    record.TaskName,
			ExecuteTime: record.CreateTime.Format(time.RFC3339),
			Success:     record.Success == 1,
			Result:      record.Response,
		}
		
		// If task failed, include error information
		if record.Success == 0 {
			// Use Response as the error message when task failed
			// or we could use a different field based on your business logic
			taskRecord.Error = record.Response
		}
		
		protoRecords = append(protoRecords, taskRecord)
	}
	
	return &pb.TaskHistoryResponse{
		Records: protoRecords,
		Total:   total,
		Success: true,
	}, nil
}

// GetStatistics implements the GetStatistics RPC method
func (s *DataServiceServer) GetStatistics(ctx context.Context, req *pb.StatisticsRequest) (*pb.StatisticsResponse, error) {
	// Convert period string to time range
	now := time.Now()
	var year, month int
	var departmentID *int64 = nil
	
	if req.DepartmentId > 0 {
		deptID := req.DepartmentId
		departmentID = &deptID
	}
	
	// Set period based on request
	switch req.Period {
	case "daily":
		year = now.Year()
		month = int(now.Month())
	case "weekly":
		year = now.Year()
		month = int(now.Month())
	case "monthly":
		year = now.Year()
		month = int(now.Month())
	default:
		year = now.Year()
		month = int(now.Month())
	}
	
	// Get statistics from service
	stats, err := s.taskService.GetRecordStats(year, month, nil, departmentID)
	if err != nil {
		logger.Errorf("Failed to get statistics via RPC: %v", err)
		return nil, err
	}
	
	// Extract values from the stats map
	taskCount := int32(0)
	successRate := float32(0)
	avgExecutionTime := float32(0)
	executionStats := make(map[string]float32)
	
	if count, ok := stats["taskCount"].(int); ok {
		taskCount = int32(count)
	}
	
	if rate, ok := stats["successRate"].(float64); ok {
		successRate = float32(rate)
	}
	
	if avgTime, ok := stats["avgExecutionTime"].(float64); ok {
		avgExecutionTime = float32(avgTime)
	}
	
	// Convert any additional statistics
	if dailyStats, ok := stats["dailyStats"].(map[string]interface{}); ok {
		for date, value := range dailyStats {
			if floatValue, ok := value.(float64); ok {
				executionStats[date] = float32(floatValue)
			}
		}
	}
	
	return &pb.StatisticsResponse{
		TaskCount:        taskCount,
		SuccessRate:      successRate,
		AvgExecutionTime: avgExecutionTime,
		ExecutionStats:   executionStats,
	}, nil
}