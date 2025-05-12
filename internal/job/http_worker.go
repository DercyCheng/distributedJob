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
	"distributedJob/pkg/logger"
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
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
			},
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
func (w *HTTPWorker) Submit(job *JobContext) chan<- *JobContext {
	return w.workQueue
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
		var fallbackErr error
		response, fallbackErr = w.executeHTTPRequest(task, result)
		// 如果备用URL成功，使用备用URL的结果
		if fallbackErr == nil {
			err = nil
		} else {
			logger.Warnf("Fallback URL also failed for HTTP job: %s (ID: %d), Error: %v", task.Name, task.ID, fallbackErr)
		}
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
		body, readErr := ioutil.ReadAll(response.Body)
		// 关闭响应体，防止内存泄漏
		defer response.Body.Close()

		if readErr != nil {
			logger.Warnf("Failed to read response body for task %d: %v", task.ID, readErr)
			result.Response = fmt.Sprintf("Error reading response: %v", readErr)
		} else {
			// 检查响应内容大小，避免存储过大的响应
			if len(body) > 1024*1024 { // 超过1MB
				logger.Warnf("Response body for task %d is too large (%d bytes), truncating", task.ID, len(body))
				result.Response = string(body[:1024*1024]) + "... [TRUNCATED]"
			} else {
				result.Response = string(body)
			}

			// 记录响应内容类型
			contentType := response.Header.Get("Content-Type")
			if contentType != "" {
				logger.Debugf("Response content type for task %d: %s", task.ID, contentType)
			}
		}
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

	// 使用任务配置的超时或默认60秒
	timeout := 60 * time.Second
	if task.Timeout > 0 {
		timeout = time.Duration(task.Timeout) * time.Second
	}

	// 覆盖客户端默认超时，确保使用任务特定的超时
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
		},
	}

	// 创建一个可取消的context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 使用context创建请求，确保超时能正确传播
	req, err := http.NewRequestWithContext(ctx, task.HTTPMethod, result.ActualURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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

	// 记录开始执行请求的时间，用于调试超时问题
	startTime := time.Now()
	logger.Debugf("Starting HTTP request for task %d with timeout %v", task.ID, timeout)

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		elapsedTime := time.Since(startTime)

		// 详细区分不同类型的错误，特别是超时和取消
		if ctx.Err() == context.DeadlineExceeded {
			logger.Warnf("HTTP request timed out after %v (limit: %v) for task %d", elapsedTime, timeout, task.ID)
			return nil, fmt.Errorf("request timed out after %v (configured timeout: %v)", elapsedTime, timeout)
		}
		if ctx.Err() == context.Canceled {
			logger.Warnf("HTTP request was canceled after %v for task %d", elapsedTime, task.ID)
			return nil, fmt.Errorf("request was canceled after %v", elapsedTime)
		}

		logger.Warnf("HTTP request failed after %v for task %d: %v", elapsedTime, task.ID, err)
		return nil, err
	}

	logger.Debugf("HTTP request completed in %v for task %d with status %d", time.Since(startTime), task.ID, resp.StatusCode)
	return resp, nil
}
