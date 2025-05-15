package protocol

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session 表示一个MCP会话
type Session struct {
	ID        string                 `json:"id"`         // 会话唯一标识符
	Messages  []Message              `json:"messages"`   // 会话历史消息
	CreatedAt time.Time              `json:"created_at"` // 创建时间
	UpdatedAt time.Time              `json:"updated_at"` // 最后更新时间
	Metadata  map[string]interface{} `json:"metadata"`   // 会话元数据
}

// SessionManager 负责管理多个会话
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionManager 创建一个新的会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession 创建新会话并返回其ID
func (sm *SessionManager) CreateSession() *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := uuid.New().String()
	now := time.Now()

	session := &Session{
		ID:        id,
		Messages:  []Message{},
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]interface{}),
	}

	sm.sessions[id] = session
	return session
}

// GetSession 通过ID获取会话
func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[id]
	return session, exists
}

// UpdateSession 更新会话消息历史
func (sm *SessionManager) UpdateSession(id string, messages []Message) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[id]
	if !exists {
		return false
	}

	session.Messages = messages
	session.UpdatedAt = time.Now()
	return true
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(id string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, exists := sm.sessions[id]
	if !exists {
		return false
	}

	delete(sm.sessions, id)
	return true
}

// AddMessageToSession 向会话添加一条新消息
func (sm *SessionManager) AddMessageToSession(id string, message Message) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[id]
	if !exists {
		return false
	}

	session.Messages = append(session.Messages, message)
	session.UpdatedAt = time.Now()
	return true
}
