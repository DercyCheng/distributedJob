package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"distributedJob/internal/mcp/protocol"
)

// LocalClient 实现本地LLM服务器的MCP客户端
type LocalClient struct {
	apiServerURL string
	model        string
	httpClient   *http.Client
	models       []LocalModelConfig
}

// NewLocalClient 创建新的本地模型客户端
func NewLocalClient(config Config) *LocalClient {
	return &LocalClient{
		apiServerURL: config.LocalConfig.APIServerURL,
		model:        config.Model,
		httpClient:   &http.Client{},
		models:       config.LocalConfig.Models,
	}
}

// Chat 实现Client接口，向本地模型服务器发送聊天请求
func (c *LocalClient) Chat(ctx context.Context, request *protocol.ChatRequest) (*protocol.ChatResponse, error) {
	// 构建API端点
	endpoint := fmt.Sprintf("%s/v1/chat/completions", c.apiServerURL)

	// 复制请求数据
	localReq := map[string]interface{}{
		"messages": request.Messages,
	}

	// 设置模型
	modelName := request.Model
	if modelName == "" {
		modelName = c.model
	}
	localReq["model"] = modelName

	// 设置其他可选参数
	if request.MaxTokens > 0 {
		localReq["max_tokens"] = request.MaxTokens
	}
	if request.Temperature > 0 {
		localReq["temperature"] = request.Temperature
	}
	if request.Stream {
		localReq["stream"] = request.Stream
	}
	if len(request.Tools) > 0 {
		localReq["tools"] = request.Tools
	}

	// 将请求编码为JSON
	reqBody, err := json.Marshal(localReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to local model server: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("local model server request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var localResp struct {
		Id      string `json:"id"`
		Object  string `json:"object"`
		Model   string `json:"model"`
		Choices []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&localResp); err != nil {
		return nil, fmt.Errorf("failed to decode response from local model server: %w", err)
	}

	// 检查是否有有效的选择
	if len(localResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response from local model server")
	}

	// 转换为MCP响应格式
	mcpResponse := &protocol.ChatResponse{
		Content:      localResp.Choices[0].Message.Content,
		FinishReason: localResp.Choices[0].FinishReason,
		Usage: protocol.Usage{
			PromptTokens:     localResp.Usage.PromptTokens,
			CompletionTokens: localResp.Usage.CompletionTokens,
			TotalTokens:      localResp.Usage.TotalTokens,
		},
	}

	return mcpResponse, nil
}

// StreamChat 实现Client接口，向本地模型服务器发送流式聊天请求
func (c *LocalClient) StreamChat(ctx context.Context, request *protocol.ChatRequest) (<-chan protocol.ChatResponse, <-chan error, error) {
	responseChan := make(chan protocol.ChatResponse)
	errChan := make(chan error, 1)

	// 构建API端点
	endpoint := fmt.Sprintf("%s/v1/chat/completions", c.apiServerURL)

	// 复制请求数据
	localReq := map[string]interface{}{
		"messages": request.Messages,
		"stream":   true,
	}

	// 设置模型
	modelName := request.Model
	if modelName == "" {
		modelName = c.model
	}
	localReq["model"] = modelName

	// 设置其他可选参数
	if request.MaxTokens > 0 {
		localReq["max_tokens"] = request.MaxTokens
	}
	if request.Temperature > 0 {
		localReq["temperature"] = request.Temperature
	}
	if len(request.Tools) > 0 {
		localReq["tools"] = request.Tools
	}

	// 将请求编码为JSON
	reqBody, err := json.Marshal(localReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	// 在后台处理流式响应
	go func() {
		defer close(responseChan)
		defer close(errChan)

		// 发送请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to send request to local model server: %w", err)
			return
		}
		defer resp.Body.Close()

		// 检查HTTP状态码
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("local model server request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
			return
		}

		// 使用流式解析器处理SSE格式的响应
		reader := bufio.NewReader(resp.Body)
		for {
			// 检查上下文是否取消
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				// 继续处理
			}

			// 读取一行
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					errChan <- fmt.Errorf("error reading stream from local model server: %w", err)
				}
				return
			}

			// 跳过空行
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// 检查是否是数据行
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			// 提取JSON数据
			data := strings.TrimPrefix(line, "data: ")

			// 检查流结束
			if data == "[DONE]" {
				return
			}

			// 解析数据
			var streamResp struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				errChan <- fmt.Errorf("failed to unmarshal stream data from local model server: %w", err)
				continue
			}

			// 检查是否有有效数据
			if len(streamResp.Choices) == 0 {
				continue
			}

			// 发送响应片段
			responseChan <- protocol.ChatResponse{
				Content:      streamResp.Choices[0].Delta.Content,
				FinishReason: streamResp.Choices[0].FinishReason,
			}
		}
	}()

	return responseChan, errChan, nil
}

// GetProvider 返回提供者名称
func (c *LocalClient) GetProvider() string {
	return "local"
}

// GetModel 返回模型名称
func (c *LocalClient) GetModel() string {
	return c.model
}

// ListAvailableModels 获取本地可用模型列表
func (c *LocalClient) ListAvailableModels(ctx context.Context) ([]string, error) {
	// 如果已配置了模型列表，直接返回
	if len(c.models) > 0 {
		modelNames := make([]string, 0, len(c.models))
		for _, model := range c.models {
			modelNames = append(modelNames, model.Name)
		}
		return modelNames, nil
	}

	// 否则，查询API获取可用模型
	endpoint := fmt.Sprintf("%s/v1/models", c.apiServerURL)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, 0, len(modelsResp.Data))
	for _, model := range modelsResp.Data {
		models = append(models, model.ID)
	}

	return models, nil
}
