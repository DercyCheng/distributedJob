package worker

import (
	"context"
	"fmt"
	"go-job/api/grpc"
	"go-job/pkg/config"
	"go-job/pkg/logger"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Worker 工作节点
type Worker struct {
	config      *config.Config
	id          string
	name        string
	ip          string
	port        int32
	capacity    int32
	currentLoad int32
	mu          sync.RWMutex
	client      grpc.SchedulerServiceClient
	conn        *grpcpkg.ClientConn
	tasks       map[string]*TaskExecution
	tasksMu     sync.RWMutex
	quit        chan struct{}
}

// TaskExecution 任务执行信息
type TaskExecution struct {
	Task      *grpc.Task
	Cmd       *exec.Cmd
	StartTime time.Time
	Cancel    context.CancelFunc
}

// NewWorker 创建工作节点
func NewWorker(cfg *config.Config) *Worker {
	hostname, _ := os.Hostname()
	ip := getLocalIP()

	return &Worker{
		config:      cfg,
		name:        fmt.Sprintf("worker-%s", hostname),
		ip:          ip,
		port:        9091, // 工作节点端口
		capacity:    10,   // 默认容量
		currentLoad: 0,
		tasks:       make(map[string]*TaskExecution),
		quit:        make(chan struct{}),
	}
}

// SetName 设置工作节点名称
func (w *Worker) SetName(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.name = name
}

// Start 启动工作节点
func (w *Worker) Start(ctx context.Context) error {
	logger.Info("启动工作节点")

	// 连接调度器
	if err := w.connectToScheduler(); err != nil {
		return fmt.Errorf("连接调度器失败: %w", err)
	}

	// 注册工作节点
	if err := w.register(); err != nil {
		return fmt.Errorf("注册工作节点失败: %w", err)
	}

	// 启动心跳
	go w.heartbeat(ctx)

	// 启动任务获取循环
	go w.taskLoop(ctx)

	// 启动任务清理
	go w.taskCleaner(ctx)

	<-ctx.Done()
	logger.Info("工作节点已停止")
	return nil
}

// Stop 停止工作节点
func (w *Worker) Stop() {
	logger.Info("正在停止工作节点")

	// 取消所有正在执行的任务
	w.tasksMu.Lock()
	for _, task := range w.tasks {
		if task.Cancel != nil {
			task.Cancel()
		}
	}
	w.tasksMu.Unlock()

	close(w.quit)

	if w.conn != nil {
		w.conn.Close()
	}
}

// connectToScheduler 连接调度器
func (w *Worker) connectToScheduler() error {
	conn, err := grpcpkg.Dial(
		w.config.GetGRPCAddr(),
		grpcpkg.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	w.conn = conn
	w.client = grpc.NewSchedulerServiceClient(conn)

	logger.Infof("已连接到调度器: %s", w.config.GetGRPCAddr())
	return nil
}

// register 注册工作节点
func (w *Worker) register() error {
	req := &grpc.RegisterWorkerRequest{
		Name:     w.name,
		Ip:       w.ip,
		Port:     w.port,
		Capacity: w.capacity,
		Metadata: map[string]string{
			"hostname": w.name,
			"version":  "1.0.0",
		},
	}

	resp, err := w.client.RegisterWorker(context.Background(), req)
	if err != nil {
		return err
	}

	w.id = resp.GetWorkerId()
	logger.Infof("工作节点注册成功: %s", w.id)
	return nil
}

// heartbeat 心跳
func (w *Worker) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(w.config.Scheduler.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.sendHeartbeat()
		}
	}
}

// sendHeartbeat 发送心跳
func (w *Worker) sendHeartbeat() {
	w.mu.RLock()
	currentLoad := w.currentLoad
	w.mu.RUnlock()

	var status grpc.WorkerStatus
	if currentLoad >= w.capacity {
		status = grpc.WorkerStatus_BUSY
	} else {
		status = grpc.WorkerStatus_ONLINE
	}

	req := &grpc.HeartbeatRequest{
		WorkerId:    w.id,
		CurrentLoad: currentLoad,
		Status:      status,
	}

	_, err := w.client.Heartbeat(context.Background(), req)
	if err != nil {
		logger.WithError(err).Error("发送心跳失败")
	}
}

// taskLoop 任务获取循环
func (w *Worker) taskLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.getTasks()
		}
	}
}

// getTasks 获取任务
func (w *Worker) getTasks() {
	w.mu.RLock()
	availableCapacity := w.capacity - w.currentLoad
	w.mu.RUnlock()

	if availableCapacity <= 0 {
		return
	}

	req := &grpc.GetTaskRequest{
		WorkerId: w.id,
		Capacity: availableCapacity,
	}

	resp, err := w.client.GetTask(context.Background(), req)
	if err != nil {
		logger.WithError(err).Error("获取任务失败")
		return
	}

	for _, task := range resp.GetTasks() {
		go w.executeTask(task)
	}
}

// executeTask 执行任务
func (w *Worker) executeTask(task *grpc.Task) {
	logger.Infof("开始执行任务: %s", task.GetId())

	// 增加当前负载
	w.mu.Lock()
	w.currentLoad++
	w.mu.Unlock()

	defer func() {
		// 减少当前负载
		w.mu.Lock()
		w.currentLoad--
		w.mu.Unlock()

		// 清理任务记录
		w.tasksMu.Lock()
		delete(w.tasks, task.GetId())
		w.tasksMu.Unlock()
	}()

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(task.GetTimeout())*time.Second)
	defer cancel()

	// 记录任务执行信息
	execution := &TaskExecution{
		Task:      task,
		StartTime: startTime,
		Cancel:    cancel,
	}

	w.tasksMu.Lock()
	w.tasks[task.GetId()] = execution
	w.tasksMu.Unlock()

	// 解析命令
	parts := strings.Fields(task.GetCommand())
	if len(parts) == 0 {
		w.reportResult(task, grpc.ExecutionStatus_FAILED, "", "命令为空", 1, startTime, time.Now())
		return
	}

	// 创建命令
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	execution.Cmd = cmd

	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range task.GetParams() {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// 添加工作节点信息到环境变量
	cmd.Env = append(cmd.Env, fmt.Sprintf("WORKER_ID=%s", w.id))
	cmd.Env = append(cmd.Env, fmt.Sprintf("WORKER_NAME=%s", w.name))
	cmd.Env = append(cmd.Env, fmt.Sprintf("TASK_ID=%s", task.GetId()))

	// 执行命令
	output, err := cmd.CombinedOutput()
	finishTime := time.Now()

	var status grpc.ExecutionStatus
	var errorMsg string
	var exitCode int32

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			status = grpc.ExecutionStatus_TIMEOUT
			errorMsg = "任务执行超时"
		} else {
			status = grpc.ExecutionStatus_FAILED
			errorMsg = err.Error()
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = int32(exitError.ExitCode())
		} else {
			exitCode = 1
		}
	} else {
		status = grpc.ExecutionStatus_SUCCESS
		exitCode = 0
	}

	w.reportResult(task, status, string(output), errorMsg, exitCode, startTime, finishTime)
}

// reportResult 报告任务结果
func (w *Worker) reportResult(task *grpc.Task, status grpc.ExecutionStatus, output, errorMsg string, exitCode int32, startTime, finishTime time.Time) {
	logger.Infof("报告任务结果: %s, 状态: %v", task.GetId(), status)

	req := &grpc.ReportTaskResultRequest{
		TaskId:     task.GetId(),
		WorkerId:   w.id,
		Status:     status,
		Output:     output,
		Error:      errorMsg,
		ExitCode:   exitCode,
		StartedAt:  timestamppb.New(startTime),
		FinishedAt: timestamppb.New(finishTime),
	}

	_, err := w.client.ReportTaskResult(context.Background(), req)
	if err != nil {
		logger.WithError(err).Errorf("报告任务结果失败: %s", task.GetId())
	}
}

// taskCleaner 任务清理器
func (w *Worker) taskCleaner(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.cleanupTimeoutTasks()
		}
	}
}

// cleanupTimeoutTasks 清理超时任务
func (w *Worker) cleanupTimeoutTasks() {
	w.tasksMu.Lock()
	defer w.tasksMu.Unlock()

	now := time.Now()
	for id, execution := range w.tasks {
		timeout := time.Duration(execution.Task.GetTimeout()) * time.Second
		if now.Sub(execution.StartTime) > timeout+time.Minute {
			logger.Warnf("清理超时任务: %s", id)
			if execution.Cancel != nil {
				execution.Cancel()
			}
			delete(w.tasks, id)
		}
	}
}

// GetMetrics 获取工作节点性能指标
func (w *Worker) GetMetrics() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	w.tasksMu.RLock()
	defer w.tasksMu.RUnlock()

	return map[string]interface{}{
		"worker_id":    w.id,
		"worker_name":  w.name,
		"capacity":     w.capacity,
		"current_load": w.currentLoad,
		"utilization":  float64(w.currentLoad) / float64(w.capacity) * 100,
		"active_tasks": len(w.tasks),
		"uptime":       time.Since(time.Now()).Seconds(), // 简化实现
	}
}

// SetCapacity 动态设置工作节点容量
func (w *Worker) SetCapacity(capacity int32) {
	w.mu.Lock()
	defer w.mu.Unlock()

	oldCapacity := w.capacity
	w.capacity = capacity

	logger.Infof("工作节点容量从 %d 调整为 %d", oldCapacity, capacity)
}

// getLocalIP 获取本地IP地址
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
