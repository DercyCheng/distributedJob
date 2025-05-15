package document

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Document 表示一个文档，包含内容和元数据
type Document struct {
	ID          string                 `json:"id,omitempty"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	ContentType string                 `json:"content_type"`
	ChunkCount  int                    `json:"chunk_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Chunk 表示文档的一个块
type Chunk struct {
	ID         string                 `json:"id,omitempty"`
	Content    string                 `json:"content"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	ParentID   string                 `json:"parent_id,omitempty"` // 父文档ID
	ChunkIndex int                    `json:"chunk_index"`         // 块在文档中的索引
}

// Processor 文档处理接口
type Processor interface {
	// Process 处理文档，返回处理后的文档
	Process(doc Document) (Document, error)

	// ProcessReader 从Reader读取内容并处理
	ProcessReader(reader io.Reader, metadata map[string]interface{}) (Document, error)
}

// Chunker 文档分块接口
type Chunker interface {
	// SplitDocument 将文档分割成块
	SplitDocument(doc Document) ([]Chunk, error)
}

// DefaultProcessor 默认文档处理器实现
type DefaultProcessor struct {
	// 可能的配置参数...
}

// NewDefaultProcessor 创建默认文档处理器
func NewDefaultProcessor() *DefaultProcessor {
	return &DefaultProcessor{}
}

// NewProcessor 创建一个新的文档处理器
func NewProcessor() Processor {
	return NewDefaultProcessor()
}

// Process 实现Processor接口，处理文档
func (p *DefaultProcessor) Process(doc Document) (Document, error) {
	// 基本的处理，如去除多余空白，限制长度等
	if doc.Content == "" {
		return Document{}, errors.New("document content is empty")
	}

	// 去除多余空格
	content := strings.TrimSpace(doc.Content)

	// 确保元数据存在
	metadata := doc.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// 返回处理后的文档
	return Document{
		ID:       doc.ID,
		Content:  content,
		Metadata: metadata,
	}, nil
}

// ProcessReader 实现Processor接口，从Reader读取并处理内容
func (p *DefaultProcessor) ProcessReader(reader io.Reader, metadata map[string]interface{}) (Document, error) {
	// 读取所有内容
	content, err := io.ReadAll(reader)
	if err != nil {
		return Document{}, fmt.Errorf("failed to read content: %w", err)
	}

	// 创建文档
	doc := Document{
		Content:  string(content),
		Metadata: metadata,
	}

	// 处理文档
	return p.Process(doc)
}

// TextChunker 基于文本的分块器实现
type TextChunker struct {
	ChunkSize    int
	ChunkOverlap int
}

// NewTextChunker 创建新的文本分块器
func NewTextChunker(chunkSize, chunkOverlap int) *TextChunker {
	// 默认值和验证
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	if chunkOverlap < 0 || chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 10 // 默认为10%重叠
	}

	return &TextChunker{
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
	}
}

// SplitDocument 实现Chunker接口，将文档分成文本块
func (c *TextChunker) SplitDocument(doc Document) ([]Chunk, error) {
	content := doc.Content
	if content == "" {
		return nil, errors.New("document content is empty")
	}

	// 按段落切分
	paragraphs := splitIntoParagraphs(content)

	var chunks []Chunk
	var currentChunk strings.Builder
	chunkIndex := 0

	// 遍历段落进行分块
	for _, paragraph := range paragraphs {
		// 如果当前段落加上当前块会超过大小，先保存当前块
		if currentChunk.Len() > 0 && currentChunk.Len()+len(paragraph) > c.ChunkSize {
			chunks = append(chunks, createChunk(doc, currentChunk.String(), chunkIndex))
			chunkIndex++

			// 重置当前块，但保留部分重叠
			currentChunkContent := currentChunk.String()
			currentChunk.Reset()

			// 添加重叠内容
			if c.ChunkOverlap > 0 && len(currentChunkContent) > c.ChunkOverlap {
				overlapContent := currentChunkContent[len(currentChunkContent)-c.ChunkOverlap:]
				currentChunk.WriteString(overlapContent)
			}
		}

		// 添加当前段落到当前块
		currentChunk.WriteString(paragraph)
	}

	// 保存最后一个块
	if currentChunk.Len() > 0 {
		chunks = append(chunks, createChunk(doc, currentChunk.String(), chunkIndex))
	}

	return chunks, nil
}

// splitIntoParagraphs 将文本按段落分割
func splitIntoParagraphs(text string) []string {
	// 简单实现：按双换行符分割
	rawParagraphs := strings.Split(text, "\n\n")

	var paragraphs []string
	for _, p := range rawParagraphs {
		p = strings.TrimSpace(p)
		if p != "" {
			paragraphs = append(paragraphs, p+"\n\n")
		}
	}

	return paragraphs
}

// createChunk 创建一个文档块
func createChunk(doc Document, content string, index int) Chunk {
	// 复制元数据
	metadata := make(map[string]interface{})
	for k, v := range doc.Metadata {
		metadata[k] = v
	}

	// 添加块相关元数据
	metadata["chunk_index"] = index

	return Chunk{
		Content:    content,
		Metadata:   metadata,
		ParentID:   doc.ID,
		ChunkIndex: index,
	}
}
