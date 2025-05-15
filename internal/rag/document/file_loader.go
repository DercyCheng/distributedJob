package document

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileLoader 文件加载器接口
type FileLoader interface {
	// Load 加载文件内容
	Load(filePath string) (Document, error)

	// CanLoad 检查是否可以加载此文件
	CanLoad(filePath string) bool
}

// FileLoaderRegistry 文件加载器注册表
type FileLoaderRegistry struct {
	loaders []FileLoader
}

// NewFileLoaderRegistry 创建新的文件加载器注册表
func NewFileLoaderRegistry() *FileLoaderRegistry {
	registry := &FileLoaderRegistry{}

	// 注册默认加载器
	registry.RegisterLoader(NewTextLoader())
	registry.RegisterLoader(NewMarkdownLoader())
	registry.RegisterLoader(NewCSVLoader())
	registry.RegisterLoader(NewJSONLoader())
	registry.RegisterLoader(NewYAMLLoader())

	return registry
}

// RegisterLoader 注册新的文件加载器
func (r *FileLoaderRegistry) RegisterLoader(loader FileLoader) {
	r.loaders = append(r.loaders, loader)
}

// LoadFile 加载文件
func (r *FileLoaderRegistry) LoadFile(filePath string) (Document, error) {
	// 查找可以处理此文件的加载器
	for _, loader := range r.loaders {
		if loader.CanLoad(filePath) {
			return loader.Load(filePath)
		}
	}

	// 没有找到合适的加载器，尝试作为文本文件加载
	return NewTextLoader().Load(filePath)
}

// TextLoader 文本文件加载器
type TextLoader struct{}

// NewTextLoader 创建新的文本文件加载器
func NewTextLoader() *TextLoader {
	return &TextLoader{}
}

// Load 实现FileLoader接口，加载文本文件
func (l *TextLoader) Load(filePath string) (Document, error) {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Document{}, fmt.Errorf("failed to read file: %w", err)
	}

	// 创建文档
	return Document{
		Content: string(content),
		Metadata: map[string]interface{}{
			"source":    filePath,
			"filename":  filepath.Base(filePath),
			"extension": strings.TrimPrefix(filepath.Ext(filePath), "."),
			"mime_type": "text/plain",
			"file_type": "text",
		},
	}, nil
}

// CanLoad 实现FileLoader接口，检查是否是文本文件
func (l *TextLoader) CanLoad(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".txt"
}

// MarkdownLoader Markdown文件加载器
type MarkdownLoader struct{}

// NewMarkdownLoader 创建新的Markdown文件加载器
func NewMarkdownLoader() *MarkdownLoader {
	return &MarkdownLoader{}
}

// Load 实现FileLoader接口，加载Markdown文件
func (l *MarkdownLoader) Load(filePath string) (Document, error) {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Document{}, fmt.Errorf("failed to read file: %w", err)
	}

	// 创建文档
	return Document{
		Content: string(content),
		Metadata: map[string]interface{}{
			"source":    filePath,
			"filename":  filepath.Base(filePath),
			"extension": strings.TrimPrefix(filepath.Ext(filePath), "."),
			"mime_type": "text/markdown",
			"file_type": "markdown",
		},
	}, nil
}

// CanLoad 实现FileLoader接口，检查是否是Markdown文件
func (l *MarkdownLoader) CanLoad(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".md" || ext == ".markdown"
}

// CSVLoader CSV文件加载器
type CSVLoader struct{}

// NewCSVLoader 创建新的CSV文件加载器
func NewCSVLoader() *CSVLoader {
	return &CSVLoader{}
}

// Load 实现FileLoader接口，加载CSV文件
func (l *CSVLoader) Load(filePath string) (Document, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return Document{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 解析CSV
	reader := csv.NewReader(file)

	// 读取所有记录
	records, err := reader.ReadAll()
	if err != nil {
		return Document{}, fmt.Errorf("failed to read CSV: %w", err)
	}

	// 转换为文本
	var sb strings.Builder
	for _, record := range records {
		sb.WriteString(strings.Join(record, ", "))
		sb.WriteString("\n")
	}

	// 创建文档
	return Document{
		Content: sb.String(),
		Metadata: map[string]interface{}{
			"source":    filePath,
			"filename":  filepath.Base(filePath),
			"extension": strings.TrimPrefix(filepath.Ext(filePath), "."),
			"mime_type": "text/csv",
			"file_type": "csv",
			"row_count": len(records),
		},
	}, nil
}

// CanLoad 实现FileLoader接口，检查是否是CSV文件
func (l *CSVLoader) CanLoad(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".csv"
}

// JSONLoader JSON文件加载器
type JSONLoader struct{}

// NewJSONLoader 创建新的JSON文件加载器
func NewJSONLoader() *JSONLoader {
	return &JSONLoader{}
}

// Load 实现FileLoader接口，加载JSON文件
func (l *JSONLoader) Load(filePath string) (Document, error) {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Document{}, fmt.Errorf("failed to read file: %w", err)
	}

	// 验证JSON格式
	var jsonData interface{}
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return Document{}, fmt.Errorf("invalid JSON: %w", err)
	}

	// 格式化JSON以便于阅读
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, content, "", "  "); err != nil {
		return Document{}, fmt.Errorf("failed to format JSON: %w", err)
	}

	// 创建文档
	return Document{
		Content: prettyJSON.String(),
		Metadata: map[string]interface{}{
			"source":    filePath,
			"filename":  filepath.Base(filePath),
			"extension": strings.TrimPrefix(filepath.Ext(filePath), "."),
			"mime_type": "application/json",
			"file_type": "json",
		},
	}, nil
}

// CanLoad 实现FileLoader接口，检查是否是JSON文件
func (l *JSONLoader) CanLoad(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".json"
}

// YAMLLoader YAML文件加载器
type YAMLLoader struct{}

// NewYAMLLoader 创建新的YAML文件加载器
func NewYAMLLoader() *YAMLLoader {
	return &YAMLLoader{}
}

// Load 实现FileLoader接口，加载YAML文件
func (l *YAMLLoader) Load(filePath string) (Document, error) {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Document{}, fmt.Errorf("failed to read file: %w", err)
	}

	// 验证YAML格式
	var yamlData interface{}
	if err := yaml.Unmarshal(content, &yamlData); err != nil {
		return Document{}, fmt.Errorf("invalid YAML: %w", err)
	}

	// 创建文档
	return Document{
		Content: string(content),
		Metadata: map[string]interface{}{
			"source":    filePath,
			"filename":  filepath.Base(filePath),
			"extension": strings.TrimPrefix(filepath.Ext(filePath), "."),
			"mime_type": "application/yaml",
			"file_type": "yaml",
		},
	}, nil
}

// CanLoad 实现FileLoader接口，检查是否是YAML文件
func (l *YAMLLoader) CanLoad(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".yaml" || ext == ".yml"
}

// XMLLoader XML文件加载器
type XMLLoader struct{}

// NewXMLLoader 创建新的XML文件加载器
func NewXMLLoader() *XMLLoader {
	return &XMLLoader{}
}

// Load 实现FileLoader接口，加载XML文件
func (l *XMLLoader) Load(filePath string) (Document, error) {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return Document{}, fmt.Errorf("failed to read file: %w", err)
	}

	// 验证XML格式
	var xmlData interface{}
	if err := xml.Unmarshal(content, &xmlData); err != nil {
		return Document{}, fmt.Errorf("invalid XML: %w", err)
	}

	// 创建文档
	return Document{
		Content: string(content),
		Metadata: map[string]interface{}{
			"source":    filePath,
			"filename":  filepath.Base(filePath),
			"extension": strings.TrimPrefix(filepath.Ext(filePath), "."),
			"mime_type": "application/xml",
			"file_type": "xml",
		},
	}, nil
}

// CanLoad 实现FileLoader接口，检查是否是XML文件
func (l *XMLLoader) CanLoad(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".xml"
}

// LoadFileWithType 根据文件类型加载文件
func LoadFileWithType(filePath string) (Document, error) {
	// 使用注册表加载文件
	registry := NewFileLoaderRegistry()
	return registry.LoadFile(filePath)
}

// ReadFileContent 读取文件内容为字符串
func ReadFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}
