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

// AnthropicClient 实现Anthropic API的MCP客户端
type AnthropicClient struct {
	apiKey      string
	apiEndpoint string
	model       string
	httpClient  *http.Client
}

// AnthropicRequest 表示发送到Anthropic API的请求格式
type AnthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []AnthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float32            `json:"temperature,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

// AnthropicMessage 表示Anthropic API中的消息格式
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse 表示从Anthropic API接收的响应格式
type AnthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// NewAnthropicClient 创建新的Anthropic客户端
func NewAnthropicClient(config Config) *AnthropicClient {
	endpoint := config.APIEndpoint
	if endpoint == "" {
		endpoint = "https://api.anthropic.com/v1/messages"
	}

	return &AnthropicClient{
		apiKey:      config.APIKey,
		apiEndpoint: endpoint,
		model:       config.Model,
		httpClient:  &http.Client{},
	}
}

// Chat 实现Client接口，向Anthropic发送聊天请求
func (c *AnthropicClient) Chat(ctx context.Context, request *protocol.ChatRequest) (*protocol.ChatResponse, error) {
	// 将MCP消息转换为Anthropic消息
	var anthropicMessages []AnthropicMessage
	for _, msg := range request.Messages {
		// Anthropic只支持"user"和"assistant"角色
		if msg.Role == "system" {
			// 将系统消息作为用户消息添加
			anthropicMessages = append(anthropicMessages, AnthropicMessage{
				Role:    "user",
				Content: "<<SYSTEM>>\n" + msg.Content + "\n<</SYSTEM>>",
			})
		} else if msg.Role == "user" || msg.Role == "assistant" {
			anthropicMessages = append(anthropicMessages, AnthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 准备Anthropic格式的请求
	anthropicReq := AnthropicRequest{
		Model:       request.Model,
		Messages:    anthropicMessages,
		MaxTokens:   request.MaxTokens,
		Temperature: request.Temperature,
	}

	// 如果没有指定模型，使用客户端配置的模型
	if anthropicReq.Model == "" {
		anthropicReq.Model = c.model
	}

	// 将请求编码为JSON
	reqBody, err := json.Marshal(anthropicReq)
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
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")

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
	var anthropicResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 提取响应内容
	var content strings.Builder
	for _, part := range anthropicResp.Content {
		if part.Type == "text" {
			content.WriteString(part.Text)
		}
	}

	// 转换为MCP响应格式
	mcpResponse := &protocol.ChatResponse{
		Content:      content.String(),
		FinishReason: anthropicResp.StopReason,
		Usage: protocol.Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}

	return mcpResponse, nil
}

// StreamChat 实现Client接口，向Anthropic发送流式聊天请求
func (c *AnthropicClient) StreamChat(ctx context.Context, request *protocol.ChatRequest) (<-chan protocol.ChatResponse, <-chan error, error) {
	responseChan := make(chan protocol.ChatResponse)
	errChan := make(chan error, 1)

	// 将MCP消息转换为Anthropic消息
	var anthropicMessages []AnthropicMessage
	for _, msg := range request.Messages {
		// Anthropic只支持"user"和"assistant"角色
		if msg.Role == "system" {
			// 将系统消息作为用户消息添加
			anthropicMessages = append(anthropicMessages, AnthropicMessage{
				Role:    "user",
				Content: "<<SYSTEM>>\n" + msg.Content + "\n<</SYSTEM>>",
			})
		} else if msg.Role == "user" || msg.Role == "assistant" {
			anthropicMessages = append(anthropicMessages, AnthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 准备流式请求
	anthropicReq := AnthropicRequest{
		Model:       request.Model,
		Messages:    anthropicMessages,
		MaxTokens:   request.MaxTokens,
		Temperature: request.Temperature,
		Stream:      true,
	}

	// 如果没有指定模型，使用客户端配置的模型
	if anthropicReq.Model == "" {
		anthropicReq.Model = c.model
	}

	// 将请求编码为JSON
	reqBody, err := json.Marshal(anthropicReq)
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
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")
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
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			// 检查上下文是否取消
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				// 继续处理
			}

			line := scanner.Text()

			// 跳过空行
			if line == "" {
				continue
			}

			// 检查是否是数据行
			if !strings.HasPrefix(line, "data:") {
				continue
			}

			// 提取JSON数据
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)

			// 检查流结束
			if data == "[DONE]" {
				return
			}

			// 解析数据
			var streamResp struct {
				Type    string `json:"type"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
				StopReason string `json:"stop_reason"`
			}

			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				errChan <- fmt.Errorf("failed to unmarshal stream data: %w", err)
				continue
			}

			// 仅处理"content_block_delta"类型的消息
			if streamResp.Type == "content_block_delta" && len(streamResp.Content) > 0 {
				for _, content := range streamResp.Content {
					if content.Type == "text" {
						responseChan <- protocol.ChatResponse{
							Content:      content.Text,
							FinishReason: streamResp.StopReason,
						}
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("error reading stream: %w", err)
		}
	}()

	return responseChan, errChan, nil
}

// GetProvider 返回提供者名称
func (c *AnthropicClient) GetProvider() string {
	return "anthropic"
}

// GetModel 返回模型名称
func (c *AnthropicClient) GetModel() string {
	return c.model
}
