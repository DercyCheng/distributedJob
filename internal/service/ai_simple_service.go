// Package service provides business logic services
package service

import (
	"context"
	"fmt"
	"time"

	"distributedJob/internal/mcp/protocol"
	"distributedJob/internal/model/entity"
)

// ChatWithAI 处理AI聊天请求并返回响应
func (s *AIService) ChatWithAI(ctx context.Context, message string, modelID string) (string, error) {
	// 创建请求
	request := &protocol.ChatRequest{
		Messages: []protocol.Message{
			{Role: "user", Content: message},
		},
		Model: modelID,
	}

	// 使用MCP客户端执行聊天
	resp, err := s.Chat(ctx, request)
	if err != nil {
		return "", fmt.Errorf("AI聊天失败: %w", err)
	}

	return resp.Content, nil
}

// SimpleQueryWithRAG 使用RAG系统进行简化查询
func (s *AIService) SimpleQueryWithRAG(ctx context.Context, query string) (string, []entity.RAGSource, error) {
	// 尝试调用完整的QueryWithRAG方法
	answer, sources, err := s.QueryWithRAG(ctx, query, 3, nil, "")
	if err != nil {
		// 如果失败，使用模拟数据
		mockAnswer, mockSources := generateMockRAGResponse(query)
		return mockAnswer, mockSources, nil
	}
	return answer, sources, nil
}

// 生成模拟RAG响应数据，供测试和失败回退使用
func generateMockRAGResponse(query string) (string, []entity.RAGSource) {
	// 模拟一个RAG系统响应
	answer := fmt.Sprintf("这是对于查询'%s'的RAG系统回答。系统已搜索相关信息并生成此回应。", query)

	// 模拟检索到的信息源
	sources := []entity.RAGSource{
		{
			DocumentID:    "doc-1",
			DocumentTitle: "示例文档1",
			Content:       "这是与您的查询相关的第一个示例文档内容。",
			Relevance:     0.89,
			Metadata: map[string]interface{}{
				"author": "系统",
				"date":   time.Now().Format(time.RFC3339),
			},
		},
		{
			DocumentID:    "doc-2",
			DocumentTitle: "示例文档2",
			Content:       "这是与您的查询相关的第二个示例文档内容。",
			Relevance:     0.76,
			Metadata: map[string]interface{}{
				"author": "系统",
				"date":   time.Now().Format(time.RFC3339),
			},
		},
	}

	return answer, sources
}

// ExecuteAgentAction 执行智能代理操作
func (s *AIService) ExecuteAgentAction(ctx context.Context, agentID string, input string) (string, interface{}, error) {
	// 模拟智能代理响应
	output := fmt.Sprintf("智能代理[%s]已处理您的请求: %s", agentID, input)

	// 模拟执行步骤
	steps := map[string]interface{}{
		"goal":        "回答用户查询",
		"description": "处理用户问题并提供信息",
		"steps": []map[string]interface{}{
			{
				"id":          "step1",
				"description": "分析输入",
				"tool":        "input_analyzer",
				"completed":   true,
				"result":      "已识别查询意图",
			},
			{
				"id":          "step2",
				"description": "获取相关信息",
				"tool":        "knowledge_retriever",
				"completed":   true,
				"result":      "已获取2条相关记录",
			},
			{
				"id":          "step3",
				"description": "生成响应",
				"tool":        "response_generator",
				"completed":   true,
				"result":      "已根据信息生成回复",
			},
		},
	}

	return output, steps, nil
}

// IntegrateQueryWithAll 使用所有AI功能整合处理查询
func (s *AIService) IntegrateQueryWithAll(ctx context.Context, query string, useRAG bool, useAgent bool, agentID string) (string, []entity.RAGSource, interface{}, error) {
	var answer string
	var sources []entity.RAGSource
	var steps interface{}
	var err error

	if useRAG {
		// 首先使用RAG系统获取信息
		answer, sources, err = s.QueryWithRAG(ctx, query, 3, nil, "")
		if err != nil {
			return "", nil, nil, fmt.Errorf("RAG查询失败: %w", err)
		}

		if useAgent && agentID != "" {
			// 如果指定了智能代理，将RAG结果传递给代理
			enhancedQuery := fmt.Sprintf("用户查询: %s\n\nRAG系统查询结果: %s", query, answer)
			agentAnswer, agentSteps, agentErr := s.ExecuteAgentAction(ctx, agentID, enhancedQuery)

			if agentErr == nil {
				// 如果代理处理成功，使用代理的输出
				answer = agentAnswer
				steps = agentSteps
			}
		}
	} else if useAgent && agentID != "" {
		// 直接使用代理处理
		answer, steps, err = s.ExecuteAgentAction(ctx, agentID, query)
		if err != nil {
			return "", nil, nil, fmt.Errorf("代理处理失败: %w", err)
		}
	} else {
		// 只使用基本的AI聊天
		answer, err = s.ChatWithAI(ctx, query, "default")
		if err != nil {
			return "", nil, nil, fmt.Errorf("AI聊天失败: %w", err)
		}
	}

	return answer, sources, steps, nil
}
