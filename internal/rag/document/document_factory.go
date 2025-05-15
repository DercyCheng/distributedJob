// Package document provides document processing capabilities for RAG
package document

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

// CreateDocumentFromFile 从上传的文件创建文档对象
func CreateDocumentFromFile(file *multipart.FileHeader, content []byte, title string, metadata map[string]interface{}) Document {
	// 如果未提供标题，使用文件名
	if title == "" {
		title = file.Filename
	}

	// 准备元数据
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// 添加基本元数据
	metadata["title"] = title
	metadata["filename"] = file.Filename
	metadata["size"] = file.Size
	metadata["content_type"] = file.Header.Get("Content-Type")
	metadata["extension"] = strings.TrimPrefix(filepath.Ext(file.Filename), ".")
	metadata["uploaded_at"] = time.Now().Format(time.RFC3339)

	// 创建文档
	contentType := file.Header.Get("Content-Type")

	doc := Document{
		ID:          fmt.Sprintf("doc_%d", time.Now().UnixNano()),
		Title:       title,
		Content:     string(content),
		ContentType: contentType,
		ChunkCount:  0, // 初始为0，后续会更新
		Metadata:    metadata,
		CreatedAt:   time.Now(),
	}

	return doc
}

// CreateTextSplitter 创建文本分割器实例，避免重复声明
func CreateTextSplitter(chunkSize, chunkOverlap int) *RecursiveCharacterTextSplitter {
	// 使用chunker.go中定义的NewRecursiveCharacterTextSplitter创建实例
	splitter := NewRecursiveCharacterTextSplitter(chunkSize, chunkOverlap)
	// 可以在这里添加其他配置
	return splitter
}

// SplitDocumentContent 将文档内容分割成块的辅助函数
func SplitDocumentContent(doc Document, splitter *RecursiveCharacterTextSplitter) ([]Chunk, error) {
	// 使用现有的RecursiveCharacterTextSplitter来分割文档
	return splitter.SplitDocument(doc)
}

// CreateChunksFromText 从文本创建文档块的辅助函数
func CreateChunksFromText(text string, doc Document, chunkSize int, chunkOverlap int) []Chunk {
	var chunks []Chunk

	for i := 0; i < len(text); i += chunkSize - chunkOverlap {
		end := i + chunkSize
		if end > len(text) {
			end = len(text)
		}

		chunk := Chunk{
			ID:         fmt.Sprintf("%s_chunk_%d", doc.ID, len(chunks)),
			Content:    text[i:end],
			Metadata:   doc.Metadata,
			ParentID:   doc.ID,
			ChunkIndex: len(chunks),
		}
		chunks = append(chunks, chunk)

		if end == len(text) {
			break
		}
	}

	return chunks
}

// CreateChunksFromParts 从文本片段创建块的辅助函数
func CreateChunksFromParts(parts []string, separator string, doc Document, chunkSize int) []Chunk {
	var chunks []Chunk
	var currentChunk strings.Builder

	for _, part := range parts {
		// 如果添加这部分会超过块大小且当前块不为空，则完成当前块
		if currentChunk.Len() > 0 && currentChunk.Len()+len(separator)+len(part) > chunkSize {
			chunks = append(chunks, Chunk{
				ID:         fmt.Sprintf("%s_chunk_%d", doc.ID, len(chunks)),
				Content:    currentChunk.String(),
				Metadata:   doc.Metadata,
				ParentID:   doc.ID,
				ChunkIndex: len(chunks),
			})

			// 重置当前块，但考虑重叠
			chunkOverlap := chunkSize / 10 // 默认为块大小的10%
			if chunkOverlap > 0 {
				// 计算要包含在下一个块中的内容长度
				overlapStr := currentChunk.String()
				if len(overlapStr) > chunkOverlap {
					overlapStr = overlapStr[len(overlapStr)-chunkOverlap:]
				}
				currentChunk.Reset()
				currentChunk.WriteString(overlapStr)
			} else {
				currentChunk.Reset()
			}
		}

		// 添加分隔符（除了第一部分）
		if currentChunk.Len() > 0 {
			currentChunk.WriteString(separator)
		}

		// 添加当前部分
		currentChunk.WriteString(part)
	}

	// 添加最后一个块（如果有内容）
	if currentChunk.Len() > 0 {
		chunks = append(chunks, Chunk{
			ID:         fmt.Sprintf("%s_chunk_%d", doc.ID, len(chunks)),
			Content:    currentChunk.String(),
			Metadata:   doc.Metadata,
			ParentID:   doc.ID,
			ChunkIndex: len(chunks),
		})
	}

	return chunks
}
