package types

import (
	"encoding/json"
	"time"
)

// MemoryType 表示记忆类型
type MemoryType string

const (
	BufferMemory MemoryType = "buffer" // 简单缓冲区记忆
	VectorMemory MemoryType = "vector" // 向量记忆
)

// MemoryConfig 记忆配置
type MemoryConfig struct {
	Type     MemoryType `json:"type"`     // 记忆类型
	Capacity int        `json:"capacity"` // 记忆容量
}

// MemoryManager 表示Agent记忆管理器接口
type MemoryManager interface {
	// AddMemory 添加一条记忆
	AddMemory(memory MemoryItem) error

	// GetMemory 获取指定键的记忆
	GetMemory(key string) (MemoryItem, bool)

	// SearchMemory 搜索相关记忆
	SearchMemory(query string, limit int) ([]MemoryItem, error)

	// GetRecentMemories 获取最近的记忆
	GetRecentMemories(limit int) []MemoryItem

	// ClearMemories 清除所有记忆
	ClearMemories()

	// LoadMemories 从JSON字符串加载记忆
	LoadMemories(jsonData string) error

	// SaveMemories 将记忆保存为JSON字符串
	SaveMemories() (string, error)

	// GetCapacity 获取记忆容量
	GetCapacity() int

	// GetCount 获取当前记忆数量
	GetCount() int
}

// RawMemory 表示原始记忆项，用于序列化/反序列化
type RawMemory struct {
	Key       string                 `json:"key"`
	Value     interface{}            `json:"value"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// SerializeMemories 将记忆序列化为JSON
func SerializeMemories(memories []MemoryItem) (string, error) {
	rawMemories := make([]RawMemory, len(memories))

	for i, item := range memories {
		timestamp, _ := time.Parse(time.RFC3339, "")
		if ts, ok := item.Metadata["timestamp"].(time.Time); ok {
			timestamp = ts
		} else if ts, ok := item.Metadata["timestamp"].(string); ok {
			timestamp, _ = time.Parse(time.RFC3339, ts)
		}

		rawMemories[i] = RawMemory{
			Key:       item.Key,
			Value:     item.Value,
			Metadata:  item.Metadata,
			Timestamp: timestamp,
		}
	}

	data, err := json.Marshal(rawMemories)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// DeserializeMemories 从JSON反序列化记忆
func DeserializeMemories(jsonData string) ([]MemoryItem, error) {
	var rawMemories []RawMemory

	if err := json.Unmarshal([]byte(jsonData), &rawMemories); err != nil {
		return nil, err
	}

	memories := make([]MemoryItem, len(rawMemories))
	for i, raw := range rawMemories {
		// 确保元数据存在
		if raw.Metadata == nil {
			raw.Metadata = make(map[string]interface{})
		}

		// 添加时间戳到元数据
		raw.Metadata["timestamp"] = raw.Timestamp.Format(time.RFC3339)

		memories[i] = MemoryItem{
			Key:      raw.Key,
			Value:    raw.Value,
			Metadata: raw.Metadata,
		}
	}

	return memories, nil
}
