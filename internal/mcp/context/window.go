package context

import (
	"distributedJob/internal/mcp/protocol"
)

// Window 实现滑动窗口上下文管理
type Window struct {
	maxSize  int
	messages []protocol.Message
}

// NewWindow 创建一个新的滑动窗口
func NewWindow(maxSize int) *Window {
	if maxSize <= 0 {
		maxSize = 20 // 默认最大容量
	}

	return &Window{
		maxSize:  maxSize,
		messages: make([]protocol.Message, 0, maxSize),
	}
}

// Add 添加消息到窗口，如果超出容量会移除最早的消息
func (w *Window) Add(message protocol.Message) {
	if len(w.messages) >= w.maxSize {
		// 移除最早的消息
		w.messages = w.messages[1:]
	}

	w.messages = append(w.messages, message)
}

// AddAll 一次性添加多条消息
func (w *Window) AddAll(messages []protocol.Message) {
	for _, msg := range messages {
		w.Add(msg)
	}
}

// Clear 清空窗口
func (w *Window) Clear() {
	w.messages = w.messages[:0]
}

// GetMessages 获取所有消息
func (w *Window) GetMessages() []protocol.Message {
	return w.messages
}

// GetLatestMessages 获取最近的n条消息
func (w *Window) GetLatestMessages(n int) []protocol.Message {
	if n <= 0 || n > len(w.messages) {
		return w.messages
	}

	return w.messages[len(w.messages)-n:]
}

// Size 返回当前窗口中的消息数量
func (w *Window) Size() int {
	return len(w.messages)
}

// Capacity 返回窗口的最大容量
func (w *Window) Capacity() int {
	return w.maxSize
}
