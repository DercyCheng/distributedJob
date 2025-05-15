package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Provider 嵌入模型提供者接口
type Provider interface {
	// Embed 将文本嵌入为向量
	Embed(ctx context.Context, texts []string) ([][]float32, error)

	// EmbedQuery 嵌入查询文本，可能使用不同的模型或参数
	EmbedQuery(ctx context.Context, text string) ([]float32, error)

	// GetDimension 获取嵌入向量的维度
	GetDimension() int
}

// Config 嵌入提供者配置
type Config struct {
	Type        string `json:"type"`         // "openai", "local", "cohere", "bge", etc.
	APIKey      string `json:"api_key"`      // API密钥（如果需要）
	APIEndpoint string `json:"api_endpoint"` // API端点地址
	Model       string `json:"model"`        // 模型名称
	BatchSize   int    `json:"batch_size"`   // 批处理大小
	Dimension   int    `json:"dimension"`    // 向量维度（如果已知）
	LocalPath   string `json:"local_path"`   // 本地模型路径（如适用）
}

// OpenAIEmbedder 使用OpenAI的嵌入模型
type OpenAIEmbedder struct {
	apiKey      string
	apiEndpoint string
	model       string
	batchSize   int
	dimension   int
	client      *http.Client
}

// OpenAIEmbeddingRequest OpenAI嵌入API请求格式
type OpenAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// OpenAIEmbeddingResponse OpenAI嵌入API响应格式
type OpenAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAIEmbedder 创建一个新的OpenAI嵌入提供者
func NewOpenAIEmbedder(config Config) (*OpenAIEmbedder, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	apiEndpoint := config.APIEndpoint
	if apiEndpoint == "" {
		apiEndpoint = "https://api.openai.com/v1/embeddings"
	}

	model := config.Model
	if model == "" {
		model = "text-embedding-3-small" // 默认使用3代小模型
	}

	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = 8 // 默认批处理大小
	}

	dimension := config.Dimension
	if dimension <= 0 {
		// 根据模型设置默认维度
		switch model {
		case "text-embedding-ada-002":
			dimension = 1536
		case "text-embedding-3-small":
			dimension = 1536
		case "text-embedding-3-large":
			dimension = 3072
		default:
			dimension = 1536 // 默认维度
		}
	}

	return &OpenAIEmbedder{
		apiKey:      config.APIKey,
		apiEndpoint: apiEndpoint,
		model:       model,
		batchSize:   batchSize,
		dimension:   dimension,
		client:      &http.Client{},
	}, nil
}

// Embed 实现Provider接口，将文本嵌入为向量
func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// 如果文本数量超过批处理大小，分批处理
	if len(texts) > e.batchSize {
		var allEmbeddings [][]float32
		for i := 0; i < len(texts); i += e.batchSize {
			end := i + e.batchSize
			if end > len(texts) {
				end = len(texts)
			}

			batch := texts[i:end]
			batchEmbeddings, err := e.embedBatch(ctx, batch)
			if err != nil {
				return nil, err
			}

			allEmbeddings = append(allEmbeddings, batchEmbeddings...)
		}
		return allEmbeddings, nil
	}

	// 对于单批处理
	return e.embedBatch(ctx, texts)
}

// embedBatch 对一批文本进行嵌入
func (e *OpenAIEmbedder) embedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	// 准备请求
	req := OpenAIEmbeddingRequest{
		Model: e.model,
		Input: texts,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", e.apiEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+e.apiKey)

	// 发送请求
	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		var errorResponse struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return nil, fmt.Errorf("API error: %s (%s)", errorResponse.Error.Message, errorResponse.Error.Type)
		}

		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// 解析响应
	var embeddingResp OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 提取嵌入向量
	embeddings := make([][]float32, len(embeddingResp.Data))
	for _, data := range embeddingResp.Data {
		embeddings[data.Index] = data.Embedding
	}

	return embeddings, nil
}

// EmbedQuery 实现Provider接口，嵌入单个查询文本
func (e *OpenAIEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := e.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned for query")
	}

	return embeddings[0], nil
}

// GetDimension 实现Provider接口，获取嵌入向量的维度
func (e *OpenAIEmbedder) GetDimension() int {
	return e.dimension
}

// LocalEmbedder 使用本地嵌入模型
type LocalEmbedder struct {
	apiEndpoint string
	model       string
	dimension   int
	client      *http.Client
	modelCache  map[string]bool // 缓存已知模型
	mu          sync.RWMutex    // 保护modelCache
}

// NewLocalEmbedder 创建一个新的本地嵌入提供者
func NewLocalEmbedder(config Config) (*LocalEmbedder, error) {
	apiEndpoint := config.APIEndpoint
	if apiEndpoint == "" {
		return nil, fmt.Errorf("local embedder API endpoint is required")
	}

	model := config.Model
	if model == "" {
		model = "bge-base-zh" // 默认使用BGE-base中文模型
	}

	dimension := config.Dimension
	if dimension <= 0 {
		// 根据模型设置默认维度
		switch model {
		case "bge-small-zh":
			dimension = 512
		case "bge-base-zh":
			dimension = 768
		case "bge-large-zh":
			dimension = 1024
		default:
			dimension = 768 // 默认维度
		}
	}

	return &LocalEmbedder{
		apiEndpoint: apiEndpoint,
		model:       model,
		dimension:   dimension,
		client:      &http.Client{},
		modelCache:  make(map[string]bool),
	}, nil
}

// Embed 实现Provider接口，将文本嵌入为向量
func (e *LocalEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// 准备请求
	req := map[string]interface{}{
		"model":       e.model,
		"input":       texts,
		"encode_type": "passage", // 文本编码类型
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 构造API端点
	endpoint := fmt.Sprintf("%s/v1/embeddings", e.apiEndpoint)

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return nil, fmt.Errorf("API error: %v", errorResponse)
		}
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// 解析响应
	var embeddingResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 提取嵌入向量
	embeddings := make([][]float32, len(texts))
	for _, data := range embeddingResp.Data {
		embeddings[data.Index] = data.Embedding
	}

	return embeddings, nil
}

// EmbedQuery 实现Provider接口，嵌入单个查询文本
func (e *LocalEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	// 准备请求
	req := map[string]interface{}{
		"model":       e.model,
		"input":       []string{text},
		"encode_type": "query", // 查询编码类型
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 构造API端点
	endpoint := fmt.Sprintf("%s/v1/embeddings", e.apiEndpoint)

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// 解析响应
	var embeddingResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned for query")
	}

	return embeddingResp.Data[0].Embedding, nil
}

// GetDimension 实现Provider接口，获取嵌入向量的维度
func (e *LocalEmbedder) GetDimension() int {
	return e.dimension
}

// Factory 根据配置创建合适的嵌入提供者
func Factory(config Config) (Provider, error) {
	switch config.Type {
	case "openai":
		return NewOpenAIEmbedder(config)
	case "local":
		return NewLocalEmbedder(config)
	default:
		return nil, fmt.Errorf("unsupported embedding provider type: %s", config.Type)
	}
}

// NewProvider 是Factory的别名，用于创建嵌入提供者
func NewProvider(config Config) (Provider, error) {
	return Factory(config)
}
