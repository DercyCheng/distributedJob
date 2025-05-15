package embedding

import (
	"context"
	"fmt"
	"sync"

	"distributedJob/internal/rag/document"
)

// DocumentEmbedder 负责为文档或文档块生成嵌入向量
type DocumentEmbedder struct {
	provider  Provider
	batchSize int
	mu        sync.Mutex
}

// DocumentEmbeddingResult 文档嵌入结果
type DocumentEmbeddingResult struct {
	Document document.Document // 原始文档
	Chunks   []ChunkEmbedding  // 文档块嵌入
}

// ChunkEmbedding 文档块嵌入
type ChunkEmbedding struct {
	Chunk      document.Chunk // 文档块
	Embedding  []float32      // 嵌入向量
	Dimensions int            // 向量维度
}

// NewDocumentEmbedder 创建新的文档嵌入器
func NewDocumentEmbedder(provider Provider, batchSize int) *DocumentEmbedder {
	if batchSize <= 0 {
		batchSize = 16 // 默认批处理大小
	}

	return &DocumentEmbedder{
		provider:  provider,
		batchSize: batchSize,
	}
}

// EmbedDocument 为整个文档生成嵌入向量
func (e *DocumentEmbedder) EmbedDocument(ctx context.Context, doc document.Document) ([]float32, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if doc.Content == "" {
		return nil, fmt.Errorf("empty document content")
	}

	// 直接为文档内容生成嵌入
	embeddings, err := e.provider.Embed(ctx, []string{doc.Content})
	if err != nil {
		return nil, fmt.Errorf("failed to embed document: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated for document")
	}

	return embeddings[0], nil
}

// EmbedDocumentWithChunks 对文档进行分块并为每个块生成嵌入
func (e *DocumentEmbedder) EmbedDocumentWithChunks(
	ctx context.Context,
	doc document.Document,
	chunker document.Chunker,
) (*DocumentEmbeddingResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 分块
	chunks, err := chunker.SplitDocument(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to split document: %w", err)
	}

	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks generated from document")
	}

	// 为所有块生成嵌入
	result := &DocumentEmbeddingResult{
		Document: doc,
		Chunks:   make([]ChunkEmbedding, len(chunks)),
	}

	// 收集所有块的文本内容
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	// 分批处理嵌入生成
	var allEmbeddings [][]float32
	for i := 0; i < len(texts); i += e.batchSize {
		end := i + e.batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		batchEmbeddings, err := e.provider.Embed(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to embed chunks batch %d-%d: %w", i, end, err)
		}

		allEmbeddings = append(allEmbeddings, batchEmbeddings...)
	}

	// 构建结果
	dimension := e.provider.GetDimension()
	for i, chunk := range chunks {
		result.Chunks[i] = ChunkEmbedding{
			Chunk:      chunk,
			Embedding:  allEmbeddings[i],
			Dimensions: dimension,
		}
	}

	return result, nil
}

// EmbedQuery 嵌入查询文本
func (e *DocumentEmbedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	// 使用提供者的查询嵌入方法
	embedding, err := e.provider.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	return embedding, nil
}
