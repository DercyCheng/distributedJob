// Package entity defines database entities and data models
package entity

import (
	"time"
)

// Document 表示可被索引的文档
type Document struct {
	ID        uint                   `json:"id" gorm:"primaryKey"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Source    string                 `json:"source"`
	Metadata  map[string]interface{} `json:"metadata" gorm:"type:json"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ChatSession 表示聊天会话
type ChatSession struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Title     string    `json:"title"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChatMessage 表示聊天消息
type ChatMessage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"` // 'user', 'assistant', 'system'
	Content   string    `json:"content"`
	Tokens    int       `json:"tokens"`
	CreatedAt time.Time `json:"created_at"`
}

// AgentExecution 表示代理执行记录
type AgentExecution struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	AgentID   string    `json:"agent_id"`
	UserID    uint      `json:"user_id"`
	Input     string    `json:"input"`
	Output    string    `json:"output"`
	Status    string    `json:"status"` // 'success', 'failed', 'running'
	Steps     string    `json:"steps" gorm:"type:json"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AIPreference 表示用户AI首选项
type AIPreference struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RAGSource 表示RAG检索的结果来源
type RAGSource struct {
	DocumentID    string                 `json:"document_id"`
	DocumentTitle string                 `json:"document_title"`
	Content       string                 `json:"content"`
	Relevance     float64                `json:"relevance"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}
