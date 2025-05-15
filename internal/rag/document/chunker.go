package document

import (
	"strings"
	"unicode"
)

// RecursiveCharacterTextSplitter 实现基于递归字符的文本分割
// 这是一种更高级的分块策略，先尝试按段落分，再按句子，再按词，最后按字符
type RecursiveCharacterTextSplitter struct {
	ChunkSize                  int
	ChunkOverlap               int
	LengthFunction             func(string) int
	Separators                 []string
	KeepSeparator              bool
	IsSeparatorMeaningfulWhite bool
}

// NewRecursiveCharacterTextSplitter 创建新的递归字符文本分割器
func NewRecursiveCharacterTextSplitter(chunkSize, chunkOverlap int) *RecursiveCharacterTextSplitter {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	if chunkOverlap < 0 || chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 10
	}

	// 默认分隔符列表，按优先级排序
	separators := []string{
		"\n\n", // 段落
		"\n",   // 换行
		". ",   // 句号
		"! ",   // 叹号
		"? ",   // 问号
		";",    // 分号
		":",    // 冒号
		",",    // 逗号
		" ",    // 空格
		"",     // 字符
	}

	return &RecursiveCharacterTextSplitter{
		ChunkSize:      chunkSize,
		ChunkOverlap:   chunkOverlap,
		LengthFunction: utf8Len,
		Separators:     separators,
		KeepSeparator:  true,
	}
}

// SplitDocument 实现Chunker接口，使用递归方式分块
func (s *RecursiveCharacterTextSplitter) SplitDocument(doc Document) ([]Chunk, error) {
	// 分割文本
	texts := s.splitText(doc.Content)

	// 转换为块
	chunks := make([]Chunk, len(texts))
	for i, text := range texts {
		// 复制元数据
		metadata := make(map[string]interface{})
		for k, v := range doc.Metadata {
			metadata[k] = v
		}

		// 添加块相关元数据
		metadata["chunk_index"] = i

		chunks[i] = Chunk{
			Content:    text,
			Metadata:   metadata,
			ParentID:   doc.ID,
			ChunkIndex: i,
		}
	}

	return chunks, nil
}

// splitText 递归地分割文本
func (s *RecursiveCharacterTextSplitter) splitText(text string) []string {
	// 如果文本长度小于等于块大小，直接返回
	if s.LengthFunction(text) <= s.ChunkSize {
		return []string{text}
	}

	// 尝试使用不同的分隔符分割
	for _, separator := range s.Separators {
		if separator == "" {
			// 最后的字符级别分割
			return s.splitByCharacters(text)
		}

		// 如果文本不包含当前分隔符，尝试下一个
		if !strings.Contains(text, separator) {
			continue
		}

		// 按当前分隔符分割
		splits := strings.Split(text, separator)
		var results []string
		var currentChunk strings.Builder

		// 处理每个分割后的部分
		for i, split := range splits {
			// 构造包含分隔符的文本片段
			var piece string
			if s.KeepSeparator && i > 0 {
				piece = separator + split
			} else {
				piece = split
			}

			// 如果当前块为空或添加新片段后不超过块大小，则添加
			pieceLength := s.LengthFunction(piece)
			currentLength := s.LengthFunction(currentChunk.String())

			if currentLength == 0 || currentLength+pieceLength <= s.ChunkSize {
				currentChunk.WriteString(piece)
			} else {
				// 当前块已满，保存并开始新块
				results = append(results, currentChunk.String())

				// 处理重叠
				currentChunk.Reset()
				if s.ChunkOverlap > 0 {
					// 计算重叠文本
					overlapText := getOverlapText(results[len(results)-1], s.ChunkOverlap, s.LengthFunction)
					currentChunk.WriteString(overlapText)
				}

				// 添加当前片段
				currentChunk.WriteString(piece)
			}
		}

		// 添加最后一个块
		if currentChunk.Len() > 0 {
			results = append(results, currentChunk.String())
		}

		// 递归处理每个结果，确保所有块都不超过指定大小
		var finalResults []string
		for _, result := range results {
			if s.LengthFunction(result) > s.ChunkSize {
				// 递归分割过大的块
				finalResults = append(finalResults, s.splitText(result)...)
			} else {
				finalResults = append(finalResults, result)
			}
		}

		return finalResults
	}

	// 如果所有分隔符都不存在，返回原文本
	return []string{text}
}

// splitByCharacters 按字符分割文本
func (s *RecursiveCharacterTextSplitter) splitByCharacters(text string) []string {
	var result []string
	var currentChunk strings.Builder

	// 按字符处理
	for _, r := range text {
		// 如果添加当前字符会超过块大小，保存当前块并开始新块
		if s.LengthFunction(currentChunk.String())+1 > s.ChunkSize {
			result = append(result, currentChunk.String())

			// 处理重叠
			currentChunk.Reset()
			if s.ChunkOverlap > 0 && len(result) > 0 {
				overlapText := getOverlapText(result[len(result)-1], s.ChunkOverlap, s.LengthFunction)
				currentChunk.WriteString(overlapText)
			}
		}

		// 添加当前字符
		currentChunk.WriteRune(r)
	}

	// 添加最后一个块
	if currentChunk.Len() > 0 {
		result = append(result, currentChunk.String())
	}

	return result
}

// getOverlapText 获取重叠文本
func getOverlapText(text string, overlapSize int, lengthFunc func(string) int) string {
	// 如果文本长度小于重叠大小，直接返回整个文本
	if lengthFunc(text) <= overlapSize {
		return text
	}

	// 获取最后overlapSize个字符
	runes := []rune(text)
	return string(runes[len(runes)-overlapSize:])
}

// utf8Len 计算UTF-8文本长度
func utf8Len(s string) int {
	return len([]rune(s))
}

// SemanticChunker 基于语义的分块器
type SemanticChunker struct {
	ChunkSize          int
	ChunkOverlap       int
	SimilarityFunction func(string, string) float32 // 计算两个文本片段的相似度
}

// NewSemanticChunker 创建新的语义分块器
func NewSemanticChunker(chunkSize, chunkOverlap int, similarityFunc func(string, string) float32) *SemanticChunker {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	if chunkOverlap < 0 || chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 10
	}

	// 如果没有提供相似度函数，使用简单的基于字符的重叠比例
	if similarityFunc == nil {
		similarityFunc = func(a, b string) float32 {
			if len(a) == 0 || len(b) == 0 {
				return 0
			}

			// 计算两个文本的字符重叠率
			setA := make(map[rune]bool)
			for _, r := range a {
				if !unicode.IsSpace(r) {
					setA[r] = true
				}
			}

			var overlap int
			for _, r := range b {
				if !unicode.IsSpace(r) && setA[r] {
					overlap++
				}
			}

			// 计算Jaccard相似度
			setB := make(map[rune]bool)
			for _, r := range b {
				if !unicode.IsSpace(r) {
					setB[r] = true
				}
			}

			union := len(setA)
			for r := range setB {
				if !setA[r] {
					union++
				}
			}

			if union == 0 {
				return 0
			}

			return float32(overlap) / float32(union)
		}
	}

	return &SemanticChunker{
		ChunkSize:          chunkSize,
		ChunkOverlap:       chunkOverlap,
		SimilarityFunction: similarityFunc,
	}
}

// SplitDocument 实现Chunker接口，使用语义方式分块
func (s *SemanticChunker) SplitDocument(doc Document) ([]Chunk, error) {
	// 首先使用递归分割器进行初始分块
	baseSplitter := NewRecursiveCharacterTextSplitter(s.ChunkSize, s.ChunkOverlap)
	initialChunks, err := baseSplitter.SplitDocument(doc)
	if err != nil {
		return nil, err
	}

	// 如果只有一个块，直接返回
	if len(initialChunks) <= 1 {
		return initialChunks, nil
	}

	// 执行语义合并
	return s.mergeChunksSemantically(initialChunks), nil
}

// mergeChunksSemantically 基于语义相似度合并块
func (s *SemanticChunker) mergeChunksSemantically(chunks []Chunk) []Chunk {
	if len(chunks) <= 1 {
		return chunks
	}

	var result []Chunk
	currentChunk := chunks[0]

	for i := 1; i < len(chunks); i++ {
		nextChunk := chunks[i]

		// 计算当前块与下一个块的语义相似度
		similarity := s.SimilarityFunction(currentChunk.Content, nextChunk.Content)

		// 如果相似度高于阈值并且合并后不会超出大小限制，则合并
		mergedLength := utf8Len(currentChunk.Content + nextChunk.Content)
		if similarity > 0.7 && mergedLength <= s.ChunkSize*2 { // 0.7是相似度阈值，可调整
			// 合并块
			currentChunk.Content = currentChunk.Content + nextChunk.Content
			// 更新元数据
			currentChunk.Metadata["merged"] = true
			currentChunk.Metadata["merged_chunks"] = []int{currentChunk.ChunkIndex, nextChunk.ChunkIndex}
		} else {
			// 添加当前块到结果，开始新块
			result = append(result, currentChunk)
			currentChunk = nextChunk
		}
	}

	// 添加最后一个块
	result = append(result, currentChunk)

	return result
}
