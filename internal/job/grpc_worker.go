package job

import (
	"context"
	"fmt"
	"time"

	"distributedJob/internal/model/entity"
	"distributedJob/pkg/logger"
)

// GRPCWorker 处理GRPC类型的任务
type GRPCWorker struct {
	workers     int
	workQueue   chan *JobContext
	resultQueue chan *WorkResult
}

// NewGRPCWorker 创建一个新的GRPC工作线程池
func NewGRPCWorker(workers int, jobQueue chan *JobContext) *GRPCWorker {
	return &GRPCWorker{
		workers:     workers,
		workQueue:   jobQueue,
		resultQueue: make(chan *WorkResult, 100),
	}
}

// Start 启动GRPC工作线程池
func (w *GRPCWorker) Start(ctx context.Context) {
	logger.Infof("Starting GRPC worker pool with %d workers", w.workers)

	// 启动指定数量的工作线程
	for i := 0; i < w.workers; i++ {
		go w.startWorker(ctx, i+1)
	}

	<-ctx.Done()
	logger.Info("GRPC worker pool shutting down")
}

// startWorker 启动一个工作线程
func (w *GRPCWorker) startWorker(ctx context.Context, id int) {
	logger.Infof("GRPC worker #%d started", id)

	for {
		select {
		case <-ctx.Done():
			logger.Infof("GRPC worker #%d shutting down", id)
			return
		case job := <-w.workQueue:
			if job.Task.TaskType == "GRPC" {
				result := w.processGRPCJob(job)
				w.resultQueue <- &WorkResult{
					Task:   job,
					Result: result,
				}
			}
		}
	}
}

// Submit 提交一个GRPC任务到工作队列
func (w *GRPCWorker) Submit(job *JobContext) chan<- *JobContext {
	return w.workQueue
}

// Results 返回结果队列
func (w *GRPCWorker) Results() <-chan *WorkResult {
	return w.resultQueue
}

// processGRPCJob 处理GRPC任务
func (w *GRPCWorker) processGRPCJob(job *JobContext) *JobResult {
	task := job.Task
	startTime := time.Now()
	logger.Infof("Processing GRPC job: %s (ID: %d)", task.Name, task.ID)

	result := &JobResult{
		Success:     false,
		RetryTimes:  0,
		UseFallback: false,
	}

	// 执行GRPC请求
	resp, grpcStatus, err := w.executeGRPCRequest(
		task.GrpcService,
		task.GrpcMethod,
		task.GrpcParams,
		task,
	)

	result.GrpcStatus = grpcStatus

	// 处理重试逻辑
	if err != nil && task.RetryCount > 0 {
		for i := 0; i < task.RetryCount; i++ {
			logger.Infof("Retrying GRPC job (%d/%d): %s (ID: %d)", i+1, task.RetryCount, task.Name, task.ID)
			result.RetryTimes++

			// 等待指定的重试间隔
			if task.RetryInterval > 0 {
				time.Sleep(time.Duration(task.RetryInterval) * time.Second)
			}

			resp, grpcStatus, err = w.executeGRPCRequest(
				task.GrpcService,
				task.GrpcMethod,
				task.GrpcParams,
				task,
			)

			result.GrpcStatus = grpcStatus
			if err == nil {
				break
			}
		}
	}
	// 如果主服务失败且配置了备用服务，尝试备用服务
	if err != nil && task.FallbackGrpcService != "" && task.FallbackGrpcMethod != "" {
		logger.Infof("Using fallback GRPC service for job: %s (ID: %d)", task.Name, task.ID)
		result.UseFallback = true
		var fallbackErr error
		resp, grpcStatus, fallbackErr = w.executeGRPCRequest(
			task.FallbackGrpcService,
			task.FallbackGrpcMethod,
			task.GrpcParams,
			task,
		)
		result.GrpcStatus = grpcStatus
		// 如果备用服务也失败，保留原始错误信息，否则清除错误
		if fallbackErr == nil {
			err = nil
		} else {
			logger.Warnf("Fallback GRPC service also failed for job: %s (ID: %d), Error: %v", task.Name, task.ID, fallbackErr)
		}
	}

	// 设置执行结果
	if err != nil {
		result.Success = false
		result.Response = fmt.Sprintf("Error: %v", err)
		result.Error = err
	} else {
		result.Success = true
		result.Response = resp
	}

	// 计算执行耗时
	result.CostTime = int(time.Since(startTime).Milliseconds())
	return result
}

// executeGRPCRequest 执行GRPC请求
// 注意：这是一个简化实现，实际生产环境中需要根据实际情况修改
func (w *GRPCWorker) executeGRPCRequest(service, method, paramsJSON string, task *entity.Task) (string, *int, error) {
	// 在实际环境中，这里需要使用反射或其他机制动态调用gRPC服务
	// 这里仅模拟gRPC调用过程并返回结果

	logger.Infof("Executing GRPC request: %s.%s with params: %s", service, method, paramsJSON)

	// 使用任务配置的超时或默认60秒
	timeout := 60 * time.Second
	if task != nil && task.Timeout > 0 {
		timeout = time.Duration(task.Timeout) * time.Second
	}

	// 创建一个可取消的context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 记录开始时间用于计算执行时长
	startTime := time.Now()
	logger.Debugf("Starting GRPC request for task %d with timeout %v", task.ID, timeout)

	// 模拟gRPC调用
	// 在实际实现中，这里需要连接到gRPC服务器并调用相应方法
	// 例如：
	/*
		// 创建支持超时和取消的连接选项
		dialOptions := []grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithTimeout(10 * time.Second), // 连接超时
		}

		// 使用ctx创建连接，确保可以通过ctx取消
		conn, err := grpc.DialContext(ctx, target, dialOptions...)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return "", nil, fmt.Errorf("connection timed out after %v", time.Since(startTime))
			}
			return "", nil, err
		}
		defer conn.Close()

		// 使用反射服务
		client := reflectionClient.NewClient(conn)

		// 使用ctx构建并执行请求，确保请求可以被取消
		// resp, err := client.InvokeMethod(ctx, service, method, paramsJSON)
	*/

	// 模拟请求处理时间和可能的超时/取消
	select {
	case <-time.After(100 * time.Millisecond): // 假设正常情况下请求处理需要100ms
		// 模拟成功状态码
		status := 0

		// 模拟响应
		resp := fmt.Sprintf("{\"result\":\"success\",\"service\":\"%s\",\"method\":\"%s\",\"executionTime\":\"%v\"}",
			service, method, time.Since(startTime))

		logger.Debugf("GRPC request completed in %v for task %d with status %d", time.Since(startTime), task.ID, status)
		return resp, &status, nil

	case <-ctx.Done():
		elapsedTime := time.Since(startTime)
		if ctx.Err() == context.DeadlineExceeded {
			logger.Warnf("GRPC request timed out after %v (limit: %v) for task %d", elapsedTime, timeout, task.ID)
			return "", nil, fmt.Errorf("request timed out after %v (configured timeout: %v)", elapsedTime, timeout)
		}
		logger.Warnf("GRPC request was canceled after %v for task %d", elapsedTime, task.ID)
		return "", nil, fmt.Errorf("request was canceled after %v", elapsedTime)
	}
}
