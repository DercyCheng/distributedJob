package retriever

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"distributedJob/internal/rag/vectorstore"
)

// HybridEngine 实现同时支持向量搜索和关键词搜索的混合检索引擎
type HybridEngine struct {
	*BasicEngine
	keywordWeight  float32 // 关键词搜索得分的权重
	semanticWeight float32 // 语义搜索得分的权重
}

// NewHybridEngine 创建一个新的混合检索引擎
func NewHybridEngine(config Config, keywordWeight, semanticWeight float32) (*HybridEngine, error) {
	// 创建基础检索引擎
	baseEngine, err := NewBasicEngine(config)
	if err != nil {
		return nil, err
	}

	// 设置默认权重
	if keywordWeight <= 0 && semanticWeight <= 0 {
		keywordWeight = 0.3  // 关键词搜索默认权重
		semanticWeight = 0.7 // 语义搜索默认权重
	}

	return &HybridEngine{
		BasicEngine:    baseEngine,
		keywordWeight:  keywordWeight,
		semanticWeight: semanticWeight,
	}, nil
}

// Retrieve 重写Engine接口的方法，实现混合检索策略
func (e *HybridEngine) Retrieve(ctx context.Context, query string, limit int, filters map[string]interface{}) ([]Result, error) {
	// 执行语义（向量）搜索
	semanticResults, err := e.BasicEngine.Retrieve(ctx, query, limit*2, filters) // 获取更多结果以便后续合并
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	// 执行关键词搜索
	keywordResults, err := e.keywordSearch(ctx, query, limit*2, filters)
	if err != nil {
		// 关键词搜索失败，仅使用语义搜索结果
		return semanticResults, nil
	}

	// 合并结果
	mergedResults := e.mergeResults(semanticResults, keywordResults, limit)

	return mergedResults, nil
}

// keywordSearch 执行关键词搜索
func (e *HybridEngine) keywordSearch(ctx context.Context, query string, limit int, filters map[string]interface{}) ([]Result, error) {
	// 这里简单实现，实际中可能需要更复杂的全文搜索引擎（如Elasticsearch）
	// 目前我们将模拟一个简单的关键词搜索

	// 获取所有文档
	docs, err := getAllDocuments(ctx, e.store, 1000) // 限制数量以防止内存问题
	if err != nil {
		return nil, err
	}

	// 提取查询中的关键词（简化版本）
	keywords := extractKeywords(query)

	// 对每个文档计算关键词匹配得分
	var results []Result
	for _, doc := range docs {
		// 应用过滤器
		if !matchesFilters(doc, filters) {
			continue
		}

		content := doc.Content
		if content == "" && doc.Text != "" {
			content = doc.Text
		}

		score := calculateKeywordScore(content, keywords)
		if score > 0 {
			results = append(results, Result{
				Document: doc,
				Score:    score,
				Distance: 1 - score,
			})
		}
	}

	// 按得分排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 限制结果数量
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// mergeResults 合并语义搜索和关键词搜索的结果
func (e *HybridEngine) mergeResults(semanticResults, keywordResults []Result, limit int) []Result {
	// 创建结果映射，用文档ID作为键
	resultMap := make(map[string]Result)

	// 处理语义搜索结果
	for _, result := range semanticResults {
		// 标准化分数（已经在0-1范围内）
		normalizedScore := result.Score

		// 使用语义权重
		weightedScore := normalizedScore * e.semanticWeight

		resultMap[result.Document.ID] = Result{
			Document: result.Document,
			Score:    weightedScore,
			Distance: result.Distance,
		}
	}

	// 处理关键词搜索结果
	for _, result := range keywordResults {
		// 标准化分数（已经在0-1范围内）
		normalizedScore := result.Score

		// 使用关键词权重
		weightedScore := normalizedScore * e.keywordWeight

		// 如果文档已经在映射中，合并得分
		if existing, found := resultMap[result.Document.ID]; found {
			resultMap[result.Document.ID] = Result{
				Document: existing.Document,
				Score:    existing.Score + weightedScore, // 合并得分
				Distance: existing.Distance,              // 保留原距离
			}
		} else {
			// 否则添加新结果
			resultMap[result.Document.ID] = Result{
				Document: result.Document,
				Score:    weightedScore,
				Distance: result.Distance,
			}
		}
	}

	// 将映射转换回切片
	var mergedResults []Result
	for _, result := range resultMap {
		mergedResults = append(mergedResults, result)
	}

	// 按合并后的得分排序
	sort.Slice(mergedResults, func(i, j int) bool {
		return mergedResults[i].Score > mergedResults[j].Score
	})

	// 限制结果数量
	if len(mergedResults) > limit {
		mergedResults = mergedResults[:limit]
	}

	return mergedResults
}

// 辅助函数

// getAllDocuments 从向量存储中获取所有文档
func getAllDocuments(ctx context.Context, store vectorstore.VectorStore, limit int) ([]vectorstore.Document, error) {
	// 获取文档总数
	count, err := store.Count(ctx)
	if err != nil {
		return nil, err
	}

	// 限制上限
	if count > limit {
		count = limit
	}

	// 获取文档
	return store.List(ctx, 0, count)
}

// extractKeywords 从查询中提取关键词（简化版本）
func extractKeywords(query string) []string {
	// 在实际应用中，这里可以使用更复杂的NLP技术提取关键词
	// 比如TF-IDF、TextRank等算法，或使用专门的分词工具

	// 简单分词
	words := make(map[string]bool)
	var result []string

	// 停用词列表（简化版）
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "with": true, "by": true, "about": true, "like": true,
		"的": true, "了": true, "和": true, "与": true, "或": true,
		"在": true, "是": true, "有": true, "对": true, "等": true,
	}

	// 简单分词和过滤
	var word strings.Builder
	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			word.WriteRune(r)
		} else {
			if word.Len() > 0 {
				w := strings.ToLower(word.String())
				if !stopWords[w] && len(w) > 1 {
					words[w] = true
				}
				word.Reset()
			}
		}
	}

	// 处理最后一个词
	if word.Len() > 0 {
		w := strings.ToLower(word.String())
		if !stopWords[w] && len(w) > 1 {
			words[w] = true
		}
	}

	// 转换为列表
	for w := range words {
		result = append(result, w)
	}

	return result
}

// calculateKeywordScore 计算文档内容与关键词的匹配度
func calculateKeywordScore(content string, keywords []string) float32 {
	if len(keywords) == 0 {
		return 0
	}

	content = strings.ToLower(content)
	var matchCount int

	for _, keyword := range keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			matchCount++
		}
	}

	// 计算匹配率
	return float32(matchCount) / float32(len(keywords))
}

// matchesFilters 检查文档是否匹配过滤条件
func matchesFilters(doc vectorstore.Document, filters map[string]interface{}) bool {
	if filters == nil {
		return true
	}

	for key, filterValue := range filters {
		// 从元数据中获取
		metadataValue, exists := doc.Metadata[key]
		if !exists {
			return false
		}

		// 比较值
		equal := false
		switch v := filterValue.(type) {
		case string:
			if mv, ok := metadataValue.(string); ok {
				equal = v == mv
			}
		case int:
			if mv, ok := metadataValue.(int); ok {
				equal = v == mv
			} else if mv, ok := metadataValue.(float64); ok {
				equal = float64(v) == mv
			}
		case float64:
			if mv, ok := metadataValue.(float64); ok {
				equal = v == mv
			}
		case bool:
			if mv, ok := metadataValue.(bool); ok {
				equal = v == mv
			}
		}

		if !equal {
			return false
		}
	}

	return true
}
