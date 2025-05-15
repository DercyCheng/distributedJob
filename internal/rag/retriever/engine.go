package retriever

import (
	"context"
	"fmt"
	"sync"

	"distributedJob/internal/rag/embedding"
	"distributedJob/internal/rag/vectorstore"
)

// Result 表示检索结果
type Result struct {
	Document vectorstore.Document `json:"document"`
	Score    float32              `json:"score"`
	Distance float32              `json:"distance"`
}

// Engine 检索引擎接口
type Engine interface {
	// Retrieve 基于查询检索相关文档
	Retrieve(ctx context.Context, query string, limit int, filters map[string]interface{}) ([]Result, error)

	// IndexDocument 索引文档，将其转换为向量并存储
	IndexDocument(ctx context.Context, doc vectorstore.Document) error

	// IndexDocuments 批量索引文档
	IndexDocuments(ctx context.Context, docs []vectorstore.Document) error

	// DeleteDocument 删除文档
	DeleteDocument(ctx context.Context, id string) error
}

// Config 检索引擎配置
type Config struct {
	VectorStore       string                 `json:"vector_store"`        // 向量存储类型 "memory", "postgres", "qdrant" 等
	VectorStoreConfig map[string]interface{} `json:"vector_store_config"` // 向量存储配置
	EmbeddingConfig   embedding.Config       `json:"embedding_config"`    // 嵌入配置
}

// BasicEngine 基础检索引擎实现
type BasicEngine struct {
	store    vectorstore.VectorStore
	embedder embedding.Provider
	mu       sync.Mutex
}

// NewBasicEngine 创建一个新的基础检索引擎
func NewBasicEngine(config Config) (*BasicEngine, error) {
	// 创建嵌入提供者
	embedder, err := embedding.Factory(config.EmbeddingConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding provider: %w", err)
	}

	// 创建向量存储
	var store vectorstore.VectorStore

	switch config.VectorStore {
	case "memory":
		store = vectorstore.NewMemoryVectorStore()
	case "postgres":
		pgConfig := vectorstore.PostgresConfig{
			Dimension: embedder.GetDimension(),
		}

		// 从config中提取PostgreSQL配置
		if connStr, ok := config.VectorStoreConfig["connection_string"].(string); ok {
			pgConfig.ConnectionString = connStr
		}
		if tableName, ok := config.VectorStoreConfig["table_name"].(string); ok {
			pgConfig.TableName = tableName
		}

		pgStore, err := vectorstore.NewPostgresVectorStore(pgConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL vector store: %w", err)
		}

		store = pgStore
	default:
		return nil, fmt.Errorf("unsupported vector store type: %s", config.VectorStore)
	}

	return &BasicEngine{
		store:    store,
		embedder: embedder,
	}, nil
}

// NewEngine 创建检索引擎
func NewEngine(store vectorstore.VectorStore, config Config) (Engine, error) {
	// 创建基础引擎
	engine := &BasicEngine{
		store: store,
	}

	// 如果需要，可以在这里基于配置添加其他逻辑

	return engine, nil
}

// Retrieve 实现Engine接口，基于查询检索相关文档
func (e *BasicEngine) Retrieve(ctx context.Context, query string, limit int, filters map[string]interface{}) ([]Result, error) {
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	// 将查询文本转换为向量
	queryVector, err := e.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// 搜索向量数据库
	searchResults, err := e.store.Search(ctx, queryVector, limit, filters)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 转换为检索结果格式
	results := make([]Result, len(searchResults))
	for i, sr := range searchResults {
		results[i] = Result{
			Document: sr.Document,
			Score:    sr.Score,
			Distance: sr.Distance,
		}
	}

	return results, nil
}

// IndexDocument 实现Engine接口，索引单个文档
func (e *BasicEngine) IndexDocument(ctx context.Context, doc vectorstore.Document) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 检查文档内容
	content := doc.Content
	if content == "" && doc.Text != "" {
		content = doc.Text
	}

	if content == "" {
		return fmt.Errorf("document has no content or text")
	}

	// 将文档内容转换为向量
	embeddings, err := e.embedder.Embed(ctx, []string{content})
	if err != nil {
		return fmt.Errorf("failed to embed document: %w", err)
	}

	if len(embeddings) == 0 {
		return fmt.Errorf("no embeddings generated")
	}

	// 为文档添加向量
	doc.Vector = embeddings[0]

	// 存储文档及其向量
	err = e.store.Add(ctx, []vectorstore.Document{doc})
	if err != nil {
		return fmt.Errorf("failed to add document to vector store: %w", err)
	}

	return nil
}

// IndexDocuments 实现Engine接口，批量索引文档
func (e *BasicEngine) IndexDocuments(ctx context.Context, docs []vectorstore.Document) error {
	if len(docs) == 0 {
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// 提取所有文档内容
	contents := make([]string, len(docs))
	for i, doc := range docs {
		content := doc.Content
		if content == "" && doc.Text != "" {
			content = doc.Text
		}

		if content == "" {
			return fmt.Errorf("document at index %d has no content or text", i)
		}
		contents[i] = content
	}

	// 批量嵌入
	embeddings, err := e.embedder.Embed(ctx, contents)
	if err != nil {
		return fmt.Errorf("failed to embed documents: %w", err)
	}

	if len(embeddings) != len(docs) {
		return fmt.Errorf("embedding count mismatch: expected %d, got %d", len(docs), len(embeddings))
	}

	// 将向量添加到文档
	for i := range docs {
		docs[i].Vector = embeddings[i]
	}

	// 批量存储文档
	err = e.store.Add(ctx, docs)
	if err != nil {
		return fmt.Errorf("failed to add documents to vector store: %w", err)
	}

	return nil
}

// DeleteDocument 实现Engine接口，删除文档
func (e *BasicEngine) DeleteDocument(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("document ID is empty")
	}

	return e.store.Delete(ctx, []string{id})
}
