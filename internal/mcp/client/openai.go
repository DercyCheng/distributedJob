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

// OpenAIClient 实现OpenAI API的MCP客户端
type OpenAIClient struct {
	apiKey      string
	apiEndpoint string
	model       string
	httpClient  *http.Client
}

// OpenAIRequest 表示发送到OpenAI API的请求格式
type OpenAIRequest struct {
	Model       string                   `json:"model"`
	Messages    []protocol.Message       `json:"messages"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Temperature float32                  `json:"temperature,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
	Tools       []map[string]interface{} `json:"tools,omitempty"`
}

// OpenAIResponse 表示从OpenAI API接收的响应格式
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int              `json:"index"`
		Message      protocol.Message `json:"message"`
		FinishReason string           `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIClient 创建新的OpenAI客户端
func NewOpenAIClient(config Config) *OpenAIClient {
	endpoint := config.APIEndpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}

	return &OpenAIClient{
		apiKey:      config.APIKey,
		apiEndpoint: endpoint,
		model:       config.Model,
		httpClient:  &http.Client{},
	}
}

// Chat 实现Client接口，向OpenAI发送聊天请求
func (c *OpenAIClient) Chat(ctx context.Context, request *protocol.ChatRequest) (*protocol.ChatResponse, error) {
	// 准备OpenAI格式的请求
	openAIReq := OpenAIRequest{
		Model:       request.Model,
		Messages:    request.Messages,
		MaxTokens:   request.MaxTokens,
		Temperature: request.Temperature,
		Tools:       request.Tools,
	}

	// 如果没有指定模型，使用客户端配置的模型
	if openAIReq.Model == "" {
		openAIReq.Model = c.model
	}

	// 将请求编码为JSON
	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 解析响应
	var openAIResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 检查是否有有效的选择
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// 转换为MCP响应格式
	mcpResponse := &protocol.ChatResponse{
		Content:      openAIResp.Choices[0].Message.Content,
		FinishReason: openAIResp.Choices[0].FinishReason,
		Usage: protocol.Usage{
			PromptTokens:     openAIResp.Usage.PromptTokens,
			CompletionTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:      openAIResp.Usage.TotalTokens,
		},
	}

	return mcpResponse, nil
}

// StreamChat 实现Client接口，向OpenAI发送流式聊天请求
func (c *OpenAIClient) StreamChat(ctx context.Context, request *protocol.ChatRequest) (<-chan protocol.ChatResponse, <-chan error, error) {
	responseChan := make(chan protocol.ChatResponse)
	errChan := make(chan error, 1)

	// 准备流式请求
	openAIReq := OpenAIRequest{
		Model:       request.Model,
		Messages:    request.Messages,
		MaxTokens:   request.MaxTokens,
		Temperature: request.Temperature,
		Stream:      true,
		Tools:       request.Tools,
	}

	// 如果没有指定模型，使用客户端配置的模型
	if openAIReq.Model == "" {
		openAIReq.Model = c.model
	}

	// 将请求编码为JSON
	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	// 在后台处理流式响应
	go func() {
		defer close(responseChan)
		defer close(errChan)

		// 发送请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		// 检查HTTP状态码
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
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
					errChan <- fmt.Errorf("error reading stream: %w", err)
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
				errChan <- fmt.Errorf("failed to unmarshal stream data: %w", err)
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
func (c *OpenAIClient) GetProvider() string {
	return "openai"
}

// GetModel 返回模型名称
func (c *OpenAIClient) GetModel() string {
	return c.model
}
