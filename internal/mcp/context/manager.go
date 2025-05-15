package context

import (
	"sync"

	"distributedJob/internal/mcp/protocol"
)

// Manager 负责管理AI对话的上下文
type Manager struct {
	maxContextSize int
	mu             sync.Mutex
}

// NewManager 创建一个新的上下文管理器
func NewManager(maxContextSize int) *Manager {
	if maxContextSize <= 0 {
		maxContextSize = 4096 // 默认最大上下文大小
	}

	return &Manager{
		maxContextSize: maxContextSize,
	}
}

// TrimContext 修剪消息列表以确保其不超过上下文窗口大小
// 保留系统消息和最新的用户-助手交互
func (m *Manager) TrimContext(messages []protocol.Message) []protocol.Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(messages) == 0 {
		return messages
	}

	// 简单的估计：每条消息平均100个token
	// 实际应用中应使用真实的token计数器
	estimatedTokens := len(messages) * 100

	if estimatedTokens <= m.maxContextSize {
		return messages
	}

	// 保留所有系统消息
	var systemMessages []protocol.Message
	var nonSystemMessages []protocol.Message

	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg)
		} else {
			nonSystemMessages = append(nonSystemMessages, msg)
		}
	}

	// 计算需要保留多少非系统消息
	systemTokens := len(systemMessages) * 100
	remainingTokens := m.maxContextSize - systemTokens

	if remainingTokens <= 0 {
		// 如果系统消息已占据全部空间，只保留最后一条系统消息
		if len(systemMessages) > 0 {
			return []protocol.Message{systemMessages[len(systemMessages)-1]}
		}
		return []protocol.Message{}
	}

	// 估算可以保留的非系统消息数量
	messagesToKeep := remainingTokens / 100
	if messagesToKeep > len(nonSystemMessages) {
		messagesToKeep = len(nonSystemMessages)
	}

	// 保留最近的消息
	result := append(systemMessages, nonSystemMessages[len(nonSystemMessages)-messagesToKeep:]...)

	return result
}

// MergeContexts 合并多个上下文，保留最重要的消息
func (m *Manager) MergeContexts(contexts ...[]protocol.Message) []protocol.Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	var allMessages []protocol.Message

	// 合并所有上下文
	for _, ctx := range contexts {
		allMessages = append(allMessages, ctx...)
	}

	// 按时间排序（假设按添加顺序已经排好）

	// 修剪到允许的大小
	return m.TrimContext(allMessages)
}
