// Package vectorstore provides vector database implementations for RAG
package vectorstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"distributedJob/internal/rag/embedding"
)

// QdrantStore 是Qdrant向量数据库的实现
type QdrantStore struct {
	endpoint    string
	collection  string
	apiKey      string
	client      *http.Client
	embeddingFn func([]string) ([][]float32, error)
	dimension   int
}

// QdrantConfig Qdrant配置
type QdrantConfig struct {
	Endpoint    string             `json:"endpoint"`
	Collection  string             `json:"collection"`
	APIKey      string             `json:"api_key"`
	Dimension   int                `json:"dimension"`
	EmbeddingFn embedding.Provider `json:"embedding_fn"`
}

// NewQdrantStore 创建一个新的Qdrant存储
func NewQdrantStore(config QdrantConfig) (*QdrantStore, error) {
	if config.Endpoint == "" {
		return nil, errors.New("Qdrant endpoint is required")
	}

	if config.Collection == "" {
		return nil, errors.New("collection name is required")
	}

	if config.Dimension <= 0 {
		if config.EmbeddingFn != nil {
			config.Dimension = config.EmbeddingFn.GetDimension()
		} else {
			return nil, errors.New("either dimension or embedding function must be provided")
		}
	}

	embeddingFunc := func(texts []string) ([][]float32, error) {
		if config.EmbeddingFn == nil {
			return nil, errors.New("embedding function not configured")
		}
		return config.EmbeddingFn.Embed(context.Background(), texts)
	}

	store := &QdrantStore{
		endpoint:    config.Endpoint,
		collection:  config.Collection,
		apiKey:      config.APIKey,
		client:      &http.Client{Timeout: 30 * time.Second},
		dimension:   config.Dimension,
		embeddingFn: embeddingFunc,
	}

	return store, nil
}

// CreateCollection 创建一个新集合
func (q *QdrantStore) CreateCollection(ctx context.Context, dimension int) error {
	// 如果提供了维度，使用它，否则使用配置的维度
	if dimension <= 0 {
		dimension = q.dimension
	}

	url := fmt.Sprintf("%s/collections/%s", q.endpoint, q.collection)

	body := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     dimension,
			"distance": "Cosine", // 使用余弦相似度
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal create collection request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if q.apiKey != "" {
		req.Header.Set("Api-Key", q.apiKey)
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection, status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

// CollectionExists 检查集合是否存在
func (q *QdrantStore) CollectionExists(ctx context.Context) (bool, error) {
	url := fmt.Sprintf("%s/collections/%s", q.endpoint, q.collection)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	if q.apiKey != "" {
		req.Header.Set("Api-Key", q.apiKey)
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// DeleteCollection 删除集合
func (q *QdrantStore) DeleteCollection(ctx context.Context) error {
	url := fmt.Sprintf("%s/collections/%s", q.endpoint, q.collection)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if q.apiKey != "" {
		req.Header.Set("Api-Key", q.apiKey)
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete collection, status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Add 添加文档到向量存储
func (q *QdrantStore) Add(ctx context.Context, documents []Document) error {
	if len(documents) == 0 {
		return nil
	}

	// 检查文档是否已经有向量
	textsToEmbed := []string{}
	textsIndices := []int{}

	for i, doc := range documents {
		if doc.Vector == nil || len(doc.Vector) == 0 {
			textsToEmbed = append(textsToEmbed, doc.Text)
			textsIndices = append(textsIndices, i)
		}
	}

	// 如果需要嵌入，使用嵌入函数
	if len(textsToEmbed) > 0 {
		vectors, err := q.embeddingFn(textsToEmbed)
		if err != nil {
			return fmt.Errorf("failed to generate embeddings: %w", err)
		}

		// 将嵌入向量添加到文档中
		for i, idx := range textsIndices {
			if i < len(vectors) {
				documents[idx].Vector = vectors[i]
			}
		}
	}

	// 准备批量添加请求
	points := make([]map[string]interface{}, 0, len(documents))

	for _, doc := range documents {
		// 制作一个唯一的点ID
		if doc.ID == "" {
			doc.ID = strconv.FormatInt(time.Now().UnixNano(), 10)
		}

		point := map[string]interface{}{
			"id":     doc.ID,
			"vector": doc.Vector,
			"payload": map[string]interface{}{
				"text":     doc.Text,
				"metadata": doc.Metadata,
			},
		}

		points = append(points, point)
	}

	// 构造批量上传请求
	url := fmt.Sprintf("%s/collections/%s/points", q.endpoint, q.collection)

	body := map[string]interface{}{
		"points": points,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal points: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if q.apiKey != "" {
		req.Header.Set("Api-Key", q.apiKey)
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to add points, status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Search 搜索向量存储
func (q *QdrantStore) Search(ctx context.Context, query []float32, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10 // 默认限制
	}

	url := fmt.Sprintf("%s/collections/%s/points/search", q.endpoint, q.collection)

	body := map[string]interface{}{
		"vector":       query,
		"limit":        limit,
		"with_payload": true, // 包含payload数据
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if q.apiKey != "" {
		req.Header.Set("Api-Key", q.apiKey)
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed, status: %d, response: %s", resp.StatusCode, string(body))
	}

	var searchResponse struct {
		Result []struct {
			ID      string  `json:"id"`
			Score   float32 `json:"score"`
			Payload struct {
				Text     string                 `json:"text"`
				Metadata map[string]interface{} `json:"metadata"`
			} `json:"payload"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	results := make([]SearchResult, 0, len(searchResponse.Result))

	for _, hit := range searchResponse.Result {
		results = append(results, SearchResult{
			ID:    hit.ID,
			Score: hit.Score,
			Document: Document{
				ID:       hit.ID,
				Text:     hit.Payload.Text,
				Metadata: hit.Payload.Metadata,
			},
		})
	}

	return results, nil
}

// Delete 删除指定ID的文档
func (q *QdrantStore) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// 为每个ID创建一个points_selector
	pointsSelector := map[string]interface{}{
		"ids": ids,
	}

	url := fmt.Sprintf("%s/collections/%s/points/delete", q.endpoint, q.collection)

	body := map[string]interface{}{
		"points": pointsSelector,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if q.apiKey != "" {
		req.Header.Set("Api-Key", q.apiKey)
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("delete failed, status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}
