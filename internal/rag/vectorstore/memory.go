package vectorstore

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/google/uuid"
)

// MemoryVectorStore 实现基于内存的向量存储
type MemoryVectorStore struct {
	documents map[string]Document
	mu        sync.RWMutex
}

// NewMemoryVectorStore 创建新的内存向量存储
func NewMemoryVectorStore() *MemoryVectorStore {
	return &MemoryVectorStore{
		documents: make(map[string]Document),
	}
}

// Add 添加文档到向量存储
func (s *MemoryVectorStore) Add(ctx context.Context, documents []Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range documents {
		// 如果文档没有ID，生成一个
		if documents[i].ID == "" {
			documents[i].ID = uuid.New().String()
		}

		// 添加文档到存储
		s.documents[documents[i].ID] = documents[i]
	}

	return nil
}

// Search 基于查询向量搜索最相似的文档
func (s *MemoryVectorStore) Search(ctx context.Context, queryVector []float32, limit int, filters map[string]interface{}) ([]SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 10 // 默认限制为10个结果
	}

	// 存储所有搜索结果的切片
	results := make([]SearchResult, 0, len(s.documents))

	// 遍历所有文档
	for _, doc := range s.documents {
		// 应用过滤器
		if !s.matchesFilters(doc, filters) {
			continue
		}

		// 计算余弦相似度
		similarity := cosineSimilarity(queryVector, doc.Vector)
		distance := 1 - similarity // 余弦距离

		// 添加到结果
		results = append(results, SearchResult{
			Document: doc,
			Score:    similarity,
			Distance: distance,
		})
	}

	// 按相似度排序（从高到低）
	sortSearchResults(results)

	// 限制结果数量
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// Delete 删除指定ID的文档
func (s *MemoryVectorStore) Delete(ctx context.Context, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range ids {
		delete(s.documents, id)
	}

	return nil
}

// Get 获取指定ID的文档
func (s *MemoryVectorStore) Get(ctx context.Context, id string) (Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, exists := s.documents[id]
	if !exists {
		return Document{}, fmt.Errorf("document with ID %s not found", id)
	}

	return doc, nil
}

// List 列出所有文档（可分页）
func (s *MemoryVectorStore) List(ctx context.Context, offset, limit int) ([]Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100 // 默认限制
	}

	// 获取所有文档ID
	ids := make([]string, 0, len(s.documents))
	for id := range s.documents {
		ids = append(ids, id)
	}

	// 检查偏移量是否有效
	if offset >= len(ids) {
		return []Document{}, nil
	}

	// 计算结束位置
	end := offset + limit
	if end > len(ids) {
		end = len(ids)
	}

	// 获取分页结果
	result := make([]Document, 0, end-offset)
	for i := offset; i < end; i++ {
		result = append(result, s.documents[ids[i]])
	}

	return result, nil
}

// Count 获取文档总数
func (s *MemoryVectorStore) Count(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.documents), nil
}

// DeleteCollection 删除整个集合
func (s *MemoryVectorStore) DeleteCollection(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.documents = make(map[string]Document)
	return nil
}

// CreateCollection 创建一个新集合
func (s *MemoryVectorStore) CreateCollection(ctx context.Context, dimension int) error {
	// 对于内存存储，这个操作基本上是一个空操作
	// 因为我们不需要预先创建一个集合结构
	return nil
}

// CollectionExists 检查集合是否存在
func (s *MemoryVectorStore) CollectionExists(ctx context.Context) (bool, error) {
	// 对于内存存储，集合总是存在的
	return true, nil
}

// 私有辅助方法

// matchesFilters 检查文档是否匹配过滤条件
func (s *MemoryVectorStore) matchesFilters(doc Document, filters map[string]interface{}) bool {
	if filters == nil {
		return true
	}

	for key, filterValue := range filters {
		// 首先尝试从元数据中获取
		metadataValue, exists := doc.Metadata[key]
		if !exists {
			return false
		}

		// 比较值
		if !compareValues(metadataValue, filterValue) {
			return false
		}
	}

	return true
}

// 两个辅助函数，用于计算余弦相似度和排序结果

// cosineSimilarity 计算两个向量的余弦相似度
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(sqrt(float64(normA))) * float32(sqrt(float64(normB))))
}

// sqrt 简单的平方根计算
func sqrt(x float64) float64 {
	return float64(math.Sqrt(x))
}

// sortSearchResults 按相似度从高到低排序搜索结果
func sortSearchResults(results []SearchResult) {
	// 使用简单的冒泡排序（小项目中足够了）
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-i-1; j++ {
			if results[j].Score < results[j+1].Score {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// compareValues 比较两个值是否相等
func compareValues(a, b interface{}) bool {
	// 根据类型进行适当的比较
	switch aTyped := a.(type) {
	case string:
		if bTyped, ok := b.(string); ok {
			return aTyped == bTyped
		}
	case int:
		if bTyped, ok := b.(int); ok {
			return aTyped == bTyped
		}
	case float64:
		if bTyped, ok := b.(float64); ok {
			return aTyped == bTyped
		}
	case bool:
		if bTyped, ok := b.(bool); ok {
			return aTyped == bTyped
		}
	case []interface{}:
		if bTyped, ok := b.([]interface{}); ok {
			// 检查数组长度
			if len(aTyped) != len(bTyped) {
				return false
			}
			// 比较每个元素
			for i := range aTyped {
				if !compareValues(aTyped[i], bTyped[i]) {
					return false
				}
			}
			return true
		}
	case map[string]interface{}:
		if bTyped, ok := b.(map[string]interface{}); ok {
			// 检查映射长度
			if len(aTyped) != len(bTyped) {
				return false
			}
			// 比较每个键值对
			for k, v := range aTyped {
				bv, exists := bTyped[k]
				if !exists || !compareValues(v, bv) {
					return false
				}
			}
			return true
		}
	}
	return false
}
