package generator

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"distributedJob/internal/mcp/client"
	"distributedJob/internal/mcp/protocol"
	"distributedJob/internal/rag/retriever"
)

// Response 表示RAG生成的响应
type Response struct {
	Answer  string             `json:"answer"`  // 生成的回答
	Sources []retriever.Result `json:"sources"` // 用于生成回答的来源文档
}

// Generator 负责组合检索结果和LLM生成最终回答
type Generator struct {
	retriever     retriever.Engine
	mcpClient     client.Client
	systemPrompt  string
	maxSourceDocs int
	mu            sync.Mutex
}

// Config 生成器配置
type Config struct {
	SystemPrompt  string        `json:"system_prompt"`   // 系统提示词
	MaxSourceDocs int           `json:"max_source_docs"` // 最大源文档数量
	MCPConfig     client.Config `json:"mcp_config"`      // MCP客户端配置
}

// NewGenerator 创建一个新的生成器
func NewGenerator(retriver retriever.Engine, config Config) (*Generator, error) {
	// 创建MCP客户端
	mcpClient := client.NewClient(config.MCPConfig)

	systemPrompt := config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "你是一个可靠的AI助手，会根据提供的信息回答用户问题，并保持客观准确。如果提供的信息不足以回答问题，请坦诚告知。"
	}

	maxSourceDocs := config.MaxSourceDocs
	if maxSourceDocs <= 0 {
		maxSourceDocs = 5 // 默认最多使用5个源文档
	}

	return &Generator{
		retriever:     retriver,
		mcpClient:     mcpClient,
		systemPrompt:  systemPrompt,
		maxSourceDocs: maxSourceDocs,
	}, nil
}

// Generate 生成对查询的回答
func (g *Generator) Generate(ctx context.Context, query string, filters map[string]interface{}) (*Response, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 检索相关文档
	results, err := g.retriever.Retrieve(ctx, query, g.maxSourceDocs, filters)
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	if len(results) == 0 {
		// 没有找到相关文档，调用MCP生成回答
		return g.generateWithoutContext(ctx, query)
	}

	// 使用检索结果和查询构建增强提示
	prompt := g.buildPromptWithDocuments(query, results)

	// 创建聊天请求
	chatReq := &protocol.ChatRequest{
		Messages: []protocol.Message{
			{Role: "system", Content: g.systemPrompt},
			{Role: "user", Content: prompt},
		},
	}

	// 调用MCP客户端生成回答
	chatResp, err := g.mcpClient.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("MCP chat failed: %w", err)
	}

	return &Response{
		Answer:  chatResp.Content,
		Sources: results,
	}, nil
}

// GenerateStream 流式生成回答
func (g *Generator) GenerateStream(ctx context.Context, query string, filters map[string]interface{}) (<-chan string, <-chan error, error) {
	answerChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(answerChan)
		defer close(errChan)

		// 检索相关文档
		results, err := g.retriever.Retrieve(ctx, query, g.maxSourceDocs, filters)
		if err != nil {
			errChan <- fmt.Errorf("retrieval failed: %w", err)
			return
		}

		var prompt string
		if len(results) == 0 {
			// 没有找到相关文档，使用原始查询
			prompt = query
		} else {
			// 使用检索结果和查询构建增强提示
			prompt = g.buildPromptWithDocuments(query, results)
		}

		// 创建聊天请求
		chatReq := &protocol.ChatRequest{
			Messages: []protocol.Message{
				{Role: "system", Content: g.systemPrompt},
				{Role: "user", Content: prompt},
			},
			Stream: true,
		}

		// 调用MCP客户端流式生成
		respChan, respErrChan, err := g.mcpClient.StreamChat(ctx, chatReq)
		if err != nil {
			errChan <- fmt.Errorf("MCP stream chat initialization failed: %w", err)
			return
		}

		// 转发流式响应
		for {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return

			case err, ok := <-respErrChan:
				if !ok {
					// 错误通道已关闭
					continue
				}
				errChan <- err

			case resp, ok := <-respChan:
				if !ok {
					// 响应通道已关闭，流已完成
					return
				}

				// 发送内容块
				if resp.Content != "" {
					answerChan <- resp.Content
				}
			}
		}
	}()

	return answerChan, errChan, nil
}

// buildPromptWithDocuments 使用检索的文档构建增强提示
func (g *Generator) buildPromptWithDocuments(query string, results []retriever.Result) string {
	var sb strings.Builder

	sb.WriteString("请回答以下问题。根据提供的上下文信息，如果无法从中得到答案，请坦诚地说不知道。\n\n")
	sb.WriteString("上下文信息:\n")

	// 添加检索结果作为上下文
	for i, result := range results {
		content := result.Document.Content
		if content == "" && result.Document.Text != "" {
			content = result.Document.Text
		}
		sb.WriteString(fmt.Sprintf("文档 %d:\n%s\n\n", i+1, content))
	}

	sb.WriteString("问题: " + query)

	return sb.String()
}

// generateWithoutContext 在没有相关上下文的情况下生成回答
func (g *Generator) generateWithoutContext(ctx context.Context, query string) (*Response, error) {
	chatReq := &protocol.ChatRequest{
		Messages: []protocol.Message{
			{Role: "system", Content: g.systemPrompt + "\n如果你不确定答案，请坦诚地说不知道。"},
			{Role: "user", Content: query},
		},
	}

	chatResp, err := g.mcpClient.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("MCP chat failed: %w", err)
	}

	return &Response{
		Answer:  chatResp.Content,
		Sources: nil,
	}, nil
}
