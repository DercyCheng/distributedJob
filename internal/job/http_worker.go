package job

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"distributedJob/internal/model/entity"
	"distributedJob"
)

// HTTPWorker 处理HTTP类型的任务
type HTTPWorker struct {
	client      *http.Client
	workers     int
	workQueue   chan *JobContext
	resultQueue chan *WorkResult
	workerDone  chan struct{}
}

// WorkResult 工作执行结果
type WorkResult struct {
	Task   *JobContext
	Result *JobResult
}

// NewHTTPWorker 创建一个新的HTTP工作线程池
func NewHTTPWorker(workers int, jobQueue chan *JobContext) *HTTPWorker {
	return &HTTPWorker{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		workers:     workers,
		workQueue:   jobQueue,
		resultQueue: make(chan *WorkResult, 100),
		workerDone:  make(chan struct{}),
	}
}

// Start 启动HTTP工作线程池
func (w *HTTPWorker) Start(ctx context.Context) {
	logger.Infof("Starting HTTP worker pool with %d workers", w.workers)

	// 启动指定数量的工作线程
	for i := 0; i < w.workers; i++ {
		go w.startWorker(ctx, i+1)
	}

	<-ctx.Done()
	logger.Info("HTTP worker pool shutting down")
}

// startWorker 启动一个工作线程
func (w *HTTPWorker) startWorker(ctx context.Context, id int) {
	logger.Infof("HTTP worker #%d started", id)

	for {
		select {
		case <-ctx.Done():
			logger.Infof("HTTP worker #%d shutting down", id)
			return
		case job := <-w.workQueue:
			if job.Task.TaskType == "HTTP" {
				result := w.processHTTPJob(job)
				w.resultQueue <- &WorkResult{
					Task:   job,
					Result: result,
				}
			}
		}
	}
}

// Submit 提交一个HTTP任务到工作队列
func (w *HTTPWorker) Submit(job *JobContext) {
	w.workQueue <- job
}

// Results 返回结果队列
func (w *HTTPWorker) Results() <-chan *WorkResult {
	return w.resultQueue
}

// processHTTPJob 处理HTTP任务
func (w *HTTPWorker) processHTTPJob(job *JobContext) *JobResult {
	task := job.Task
	startTime := time.Now()
	logger.Infof("Processing HTTP job: %s (ID: %d)", task.Name, task.ID)

	result := &JobResult{
		Success:      false,
		RetryTimes:   0,
		UseFallback:  false,
		ActualURL:    task.URL,
		ActualMethod: task.HTTPMethod,
	}

	// 执行HTTP请求，包括重试逻辑
	response, err := w.executeHTTPRequest(task, result)
	if err != nil && task.RetryCount > 0 {
		// 重试逻辑
		for i := 0; i < task.RetryCount; i++ {
			logger.Infof("Retrying HTTP job (%d/%d): %s (ID: %d)", i+1, task.RetryCount, task.Name, task.ID)
			result.RetryTimes++

			// 等待指定的重试间隔
			if task.RetryInterval > 0 {
				time.Sleep(time.Duration(task.RetryInterval) * time.Second)
			}

			response, err = w.executeHTTPRequest(task, result)
			if err == nil {
				break
			}
		}
	}

	// 如果主URL失败且配置了备用URL，尝试备用URL
	if err != nil && task.FallbackURL != "" {
		logger.Infof("Using fallback URL for HTTP job: %s (ID: %d)", task.Name, task.ID)
		result.UseFallback = true
		result.ActualURL = task.FallbackURL
		response, err = w.executeHTTPRequest(task, result)
	}

	// 设置执行结果
	if err != nil {
		result.Success = false
		result.Response = fmt.Sprintf("Error: %v", err)
		result.Error = err
	} else {
		// 根据HTTP状态码判断是否成功
		statusCode := response.StatusCode
		result.StatusCode = &statusCode
		result.Success = statusCode >= 200 && statusCode < 300

		// 读取响应内容
		body, _ := ioutil.ReadAll(response.Body)
		defer response.Body.Close()

		result.Response = string(body)
	}

	// 计算执行耗时
	result.CostTime = int(time.Since(startTime).Milliseconds())
	return result
}

// executeHTTPRequest 执行HTTP请求
func (w *HTTPWorker) executeHTTPRequest(task *entity.Task, result *JobResult) (*http.Response, error) {
	// 创建请求
	var reqBody *bytes.Buffer
	if task.Body != "" {
		reqBody = bytes.NewBufferString(task.Body)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequest(task.HTTPMethod, result.ActualURL, reqBody)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "DistributedJob-Scheduler/1.0")

	// 如果有自定义的请求头，添加到请求中
	if task.Headers != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(task.Headers), &headers); err == nil {
			for key, value := range headers {
				req.Header.Set(key, value)
			}
		} else {
			logger.Warnf("Failed to parse headers for task %d: %v", task.ID, err)
		}
	}

	// 执行请求
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
