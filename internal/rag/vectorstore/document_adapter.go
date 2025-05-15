// Package vectorstore 提供向量存储相关功能
package vectorstore

import (
	"context"
	"fmt"

	"distributedJob/internal/rag/document"
)

// AddDocument 添加单个文档到向量存储的便捷方法
func (vs *PostgresVectorStore) AddDocument(ctx context.Context, doc document.Chunk) error {
	// 将document.Chunk转换为vectorstore.Document
	vsDoc := Document{
		ID:       doc.ID,
		Content:  doc.Content,
		Text:     doc.Content, // 有些系统使用Text字段
		Metadata: doc.Metadata,
	}

	// 使用基础Add方法添加文档
	return vs.Add(ctx, []Document{vsDoc})
}

// AddDocument 添加单个文档到向量存储的便捷方法（通用实现）
func AddDocument(ctx context.Context, store VectorStore, doc document.Chunk) error {
	if store == nil {
		return fmt.Errorf("vector store is nil")
	}

	// 将document.Chunk转换为vectorstore.Document
	vsDoc := Document{
		ID:       doc.ID,
		Content:  doc.Content,
		Text:     doc.Content, // 有些系统使用Text字段
		Metadata: doc.Metadata,
	}

	// 使用基础Add方法添加文档
	return store.Add(ctx, []Document{vsDoc})
}
