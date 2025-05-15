package vectorstore

import "context"

// VectorStore 定义向量存储接口
type VectorStore interface {
	// Add 添加文档到向量存储
	Add(ctx context.Context, documents []Document) error

	// Search 基于查询向量搜索最相似的文档
	Search(ctx context.Context, queryVector []float32, limit int, filters map[string]interface{}) ([]SearchResult, error)

	// Delete 删除指定ID的文档
	Delete(ctx context.Context, ids []string) error

	// Get 获取指定ID的文档
	Get(ctx context.Context, id string) (Document, error)

	// List 列出所有文档（可分页）
	List(ctx context.Context, offset, limit int) ([]Document, error)

	// Count 获取文档总数
	Count(ctx context.Context) (int, error)

	// DeleteCollection 删除整个集合
	DeleteCollection(ctx context.Context) error

	// CreateCollection 创建一个新集合
	CreateCollection(ctx context.Context, dimension int) error

	// CollectionExists 检查集合是否存在
	CollectionExists(ctx context.Context) (bool, error)
}

// Document 表示向量存储中的文档
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Text     string                 `json:"text,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
	Vector   []float32              `json:"vector,omitempty"`
}

// SearchResult 表示向量搜索结果
type SearchResult struct {
	Document Document `json:"document"`
	Score    float32  `json:"score"`
	Distance float32  `json:"distance"`
	ID       string   `json:"id,omitempty"`
}
