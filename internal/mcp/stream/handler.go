package stream

import (
	"context"
	"io"
	"sync"

	"distributedJob/internal/mcp/protocol"
)

// Handler 处理流式AI响应
type Handler struct {
	mu             sync.Mutex
	responseBuffer []protocol.ChatResponse
	errors         []error
	content        string
	done           bool
}

// NewHandler 创建一个新的流处理器
func NewHandler() *Handler {
	return &Handler{
		responseBuffer: make([]protocol.ChatResponse, 0),
		errors:         make([]error, 0),
		done:           false,
	}
}

// HandleStream 处理来自流式API的响应
func (h *Handler) HandleStream(ctx context.Context, responseChan <-chan protocol.ChatResponse, errChan <-chan error) {
	h.mu.Lock()
	h.done = false
	h.responseBuffer = h.responseBuffer[:0]
	h.errors = h.errors[:0]
	h.content = ""
	h.mu.Unlock()

	go func() {
		var fullContent string

		for {
			select {
			case <-ctx.Done():
				h.mu.Lock()
				h.errors = append(h.errors, ctx.Err())
				h.done = true
				h.mu.Unlock()
				return

			case err, ok := <-errChan:
				if !ok {
					// 错误通道已关闭
					continue
				}
				h.mu.Lock()
				h.errors = append(h.errors, err)
				h.mu.Unlock()

			case resp, ok := <-responseChan:
				if !ok {
					// 响应通道已关闭，流已完成
					h.mu.Lock()
					h.done = true
					h.content = fullContent
					h.mu.Unlock()
					return
				}

				h.mu.Lock()
				h.responseBuffer = append(h.responseBuffer, resp)
				fullContent += resp.Content
				h.mu.Unlock()

				// 检查finish reason
				if resp.FinishReason != "" {
					h.mu.Lock()
					h.done = true
					h.content = fullContent
					h.mu.Unlock()
					return
				}
			}
		}
	}()
}

// IsDone 检查流是否已完成
func (h *Handler) IsDone() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.done
}

// GetContent 获取完整的响应内容
func (h *Handler) GetContent() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.content
}

// GetBufferedResponses 获取缓冲的响应片段
func (h *Handler) GetBufferedResponses() []protocol.ChatResponse {
	h.mu.Lock()
	defer h.mu.Unlock()
	// 返回副本以避免竞态条件
	result := make([]protocol.ChatResponse, len(h.responseBuffer))
	copy(result, h.responseBuffer)
	return result
}

// GetErrors 获取遇到的所有错误
func (h *Handler) GetErrors() []error {
	h.mu.Lock()
	defer h.mu.Unlock()
	// 返回副本以避免竞态条件
	result := make([]error, len(h.errors))
	copy(result, h.errors)
	return result
}

// WriteContentTo 将完整内容写入指定的writer
func (h *Handler) WriteContentTo(w io.Writer) (int64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	n, err := io.WriteString(w, h.content)
	return int64(n), err
}

// Reset 重置处理器状态
func (h *Handler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.responseBuffer = h.responseBuffer[:0]
	h.errors = h.errors[:0]
	h.content = ""
	h.done = false
}
