package core

import (
	"container/list"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"distributedJob/internal/agent/types"
)

// BufferMemory 实现基于简单缓冲区的记忆
type BufferMemory struct {
	items    *list.List               // 双向链表存储记忆项
	keyMap   map[string]*list.Element // 键到链表元素的映射
	capacity int                      // 最大容量
	mu       sync.RWMutex
}

// NewBufferMemory 创建一个新的缓冲区记忆
func NewBufferMemory(capacity int) *BufferMemory {
	if capacity <= 0 {
		capacity = 100 // 默认容量
	}

	return &BufferMemory{
		items:    list.New(),
		keyMap:   make(map[string]*list.Element),
		capacity: capacity,
	}
}

// AddMemory 实现 MemoryManager 接口，添加一条记忆
func (m *BufferMemory) AddMemory(item types.MemoryItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查键是否为空
	if item.Key == "" {
		return errors.New("memory key cannot be empty")
	}

	// 确保元数据存在
	if item.Metadata == nil {
		item.Metadata = make(map[string]interface{})
	}

	// 添加时间戳
	item.Metadata["timestamp"] = time.Now().Format(time.RFC3339)

	// 如果已存在同键记忆，先移除它
	if elem, exists := m.keyMap[item.Key]; exists {
		m.items.Remove(elem)
		delete(m.keyMap, item.Key)
	}

	// 添加新记忆到链表头部
	elem := m.items.PushFront(item)
	m.keyMap[item.Key] = elem

	// 如果超过容量，移除最老的记忆（链表尾部）
	if m.items.Len() > m.capacity {
		oldest := m.items.Back()
		if oldest != nil {
			oldestItem := oldest.Value.(types.MemoryItem)
			delete(m.keyMap, oldestItem.Key)
			m.items.Remove(oldest)
		}
	}

	return nil
}

// GetMemory 实现 MemoryManager 接口，获取指定键的记忆
func (m *BufferMemory) GetMemory(key string) (types.MemoryItem, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	elem, exists := m.keyMap[key]
	if !exists {
		return types.MemoryItem{}, false
	}

	return elem.Value.(types.MemoryItem), true
}

// SearchMemory 实现 MemoryManager 接口，搜索相关记忆
// 注意：简单实现，只支持关键词完全匹配
func (m *BufferMemory) SearchMemory(query string, limit int) ([]types.MemoryItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 {
		limit = 10 // 默认限制
	}

	var results []types.MemoryItem

	// 简单线性搜索，在实际应用中可能需要更高级的搜索算法
	for elem := m.items.Front(); elem != nil && len(results) < limit; elem = elem.Next() {
		item := elem.Value.(types.MemoryItem)

		// 尝试将值转换为字符串进行简单匹配
		var valueStr string
		if str, ok := item.Value.(string); ok {
			valueStr = str
		} else {
			// 尝试JSON序列化
			if bytes, err := json.Marshal(item.Value); err == nil {
				valueStr = string(bytes)
			}
		}

		// 简单的包含检查
		if valueStr != "" && contains(valueStr, query) {
			results = append(results, item)
		}
	}

	return results, nil
}

// GetRecentMemories 实现 MemoryManager 接口，获取最近的记忆
func (m *BufferMemory) GetRecentMemories(limit int) []types.MemoryItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > m.items.Len() {
		limit = m.items.Len()
	}

	results := make([]types.MemoryItem, 0, limit)

	// 从链表头部（最新）开始获取
	for elem := m.items.Front(); elem != nil && len(results) < limit; elem = elem.Next() {
		results = append(results, elem.Value.(types.MemoryItem))
	}

	return results
}

// ClearMemories 实现 MemoryManager 接口，清除所有记忆
func (m *BufferMemory) ClearMemories() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items.Init()                            // 清空链表
	m.keyMap = make(map[string]*list.Element) // 重新初始化映射
}

// LoadMemories 实现 MemoryManager 接口，从JSON字符串加载记忆
func (m *BufferMemory) LoadMemories(jsonData string) error {
	items, err := types.DeserializeMemories(jsonData)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 清除现有记忆
	m.items.Init()
	m.keyMap = make(map[string]*list.Element)

	// 加载新记忆
	for _, item := range items {
		elem := m.items.PushBack(item)
		m.keyMap[item.Key] = elem

		// 如果超过容量，移除最老的记忆
		if m.items.Len() > m.capacity {
			oldest := m.items.Front()
			if oldest != nil {
				oldestItem := oldest.Value.(types.MemoryItem)
				delete(m.keyMap, oldestItem.Key)
				m.items.Remove(oldest)
			}
		}
	}

	return nil
}

// SaveMemories 实现 MemoryManager 接口，将记忆保存为JSON字符串
func (m *BufferMemory) SaveMemories() (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	memories := make([]types.MemoryItem, 0, m.items.Len())

	for elem := m.items.Front(); elem != nil; elem = elem.Next() {
		memories = append(memories, elem.Value.(types.MemoryItem))
	}

	return types.SerializeMemories(memories)
}

// GetCapacity 实现 MemoryManager 接口，获取记忆容量
func (m *BufferMemory) GetCapacity() int {
	return m.capacity
}

// GetCount 实现 MemoryManager 接口，获取当前记忆数量
func (m *BufferMemory) GetCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.items.Len()
}

// contains 检查字符串s是否包含子字符串substr
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
