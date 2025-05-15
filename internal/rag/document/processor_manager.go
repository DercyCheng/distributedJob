package document

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// ProcessorManager 文档处理管理器，负责文档的加载、处理和分块
type ProcessorManager struct {
	processor Processor
	chunker   Chunker
	// 可能的扩展：支持文件类型检测器、文档转换器等
}

// ProcessorOptions 处理器选项
type ProcessorOptions struct {
	ChunkSize    int                    // 分块大小
	ChunkOverlap int                    // 块重叠大小
	ChunkerType  string                 // 分块器类型：recursive, text, semantic
	Metadata     map[string]interface{} // 要添加到所有文档的元数据
}

// ProcessorResult 处理结果
type ProcessorResult struct {
	Document Document // 原始文档
	Chunks   []Chunk  // 生成的块
	Errors   []error  // 处理过程中的错误
}

// NewProcessorManager 创建新的处理器管理器
func NewProcessorManager(options ProcessorOptions) *ProcessorManager {
	// 创建文档处理器
	processor := NewDefaultProcessor()

	// 根据配置选择适当的分块器
	var chunker Chunker
	switch strings.ToLower(options.ChunkerType) {
	case "recursive":
		chunker = NewRecursiveCharacterTextSplitter(options.ChunkSize, options.ChunkOverlap)
	case "semantic":
		chunker = NewSemanticChunker(options.ChunkSize, options.ChunkOverlap, nil)
	default:
		// 默认使用文本分块器
		chunker = NewTextChunker(options.ChunkSize, options.ChunkOverlap)
	}

	return &ProcessorManager{
		processor: processor,
		chunker:   chunker,
	}
}

// ProcessText 处理文本内容
func (pm *ProcessorManager) ProcessText(ctx context.Context, content string, metadata map[string]interface{}) (*ProcessorResult, error) {
	// 创建文档
	doc := Document{
		ID:       uuid.New().String(),
		Content:  content,
		Metadata: metadata,
	}

	// 处理文档
	processedDoc, err := pm.processor.Process(doc)
	if err != nil {
		return &ProcessorResult{
			Document: doc,
			Errors:   []error{err},
		}, nil
	}

	// 分块
	chunks, err := pm.chunker.SplitDocument(processedDoc)
	if err != nil {
		return &ProcessorResult{
			Document: processedDoc,
			Errors:   []error{err},
		}, nil
	}

	return &ProcessorResult{
		Document: processedDoc,
		Chunks:   chunks,
	}, nil
}

// ProcessFile 处理单个文件
func (pm *ProcessorManager) ProcessFile(ctx context.Context, filePath string, metadata map[string]interface{}) (*ProcessorResult, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 准备元数据
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// 添加文件相关元数据
	metadata["source"] = filePath
	metadata["filename"] = filepath.Base(filePath)
	metadata["extension"] = strings.TrimPrefix(filepath.Ext(filePath), ".")

	// 处理文件
	doc, err := pm.processor.ProcessReader(file, metadata)
	if err != nil {
		return &ProcessorResult{
			Errors: []error{err},
		}, nil
	}

	// 分块
	chunks, err := pm.chunker.SplitDocument(doc)
	if err != nil {
		return &ProcessorResult{
			Document: doc,
			Errors:   []error{err},
		}, nil
	}

	return &ProcessorResult{
		Document: doc,
		Chunks:   chunks,
	}, nil
}

// ProcessDirectory 处理目录中的所有文件
func (pm *ProcessorManager) ProcessDirectory(ctx context.Context, dirPath string, baseMetadata map[string]interface{}) ([]*ProcessorResult, error) {
	var results []*ProcessorResult
	var resultMutex sync.Mutex
	var wg sync.WaitGroup

	// 遍历目录
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if d.IsDir() {
			return nil
		}

		// 只处理文本文件（可以扩展支持更多类型）
		ext := strings.ToLower(filepath.Ext(path))
		if !isSupportedFileType(ext) {
			return nil
		}

		// 创建文件特定的元数据
		metadata := make(map[string]interface{})
		for k, v := range baseMetadata {
			metadata[k] = v
		}

		// 计算相对路径
		relPath, err := filepath.Rel(dirPath, path)
		if err == nil {
			metadata["relative_path"] = relPath
		}

		// 异步处理文件
		wg.Add(1)
		go func(filePath string, meta map[string]interface{}) {
			defer wg.Done()

			result, err := pm.ProcessFile(ctx, filePath, meta)
			if err != nil {
				// 创建错误结果
				result = &ProcessorResult{
					Errors: []error{err},
				}
			}

			// 添加到结果列表
			resultMutex.Lock()
			results = append(results, result)
			resultMutex.Unlock()
		}(path, metadata)

		return nil
	})

	// 等待所有处理完成
	wg.Wait()

	if err != nil {
		return results, fmt.Errorf("error walking directory: %w", err)
	}

	return results, nil
}

// ProcessReader 从reader处理内容
func (pm *ProcessorManager) ProcessReader(ctx context.Context, reader io.Reader, metadata map[string]interface{}) (*ProcessorResult, error) {
	// 使用处理器处理内容
	doc, err := pm.processor.ProcessReader(reader, metadata)
	if err != nil {
		return &ProcessorResult{
			Errors: []error{err},
		}, nil
	}

	// 分块
	chunks, err := pm.chunker.SplitDocument(doc)
	if err != nil {
		return &ProcessorResult{
			Document: doc,
			Errors:   []error{err},
		}, nil
	}

	return &ProcessorResult{
		Document: doc,
		Chunks:   chunks,
	}, nil
}

// isSupportedFileType 检查文件类型是否支持
func isSupportedFileType(ext string) bool {
	supportedTypes := map[string]bool{
		".txt":  true,
		".md":   true,
		".csv":  true,
		".json": true,
		".yaml": true,
		".yml":  true,
		".html": true,
		".xml":  true,
		// 可以添加更多支持的文件类型
	}

	return supportedTypes[ext]
}
