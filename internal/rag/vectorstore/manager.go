package vectorstore

import (
	"fmt"
	"sync"
)

// Manager 向量存储管理器，负责管理不同的向量存储后端
type Manager struct {
	stores map[string]VectorStore
	mu     sync.RWMutex
}

// NewManager 创建新的向量存储管理器
func NewManager() *Manager {
	return &Manager{
		stores: make(map[string]VectorStore),
	}
}

// RegisterStore 注册一个向量存储后端
func (m *Manager) RegisterStore(name string, store VectorStore) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stores[name] = store
}

// GetStore 获取指定名称的向量存储
func (m *Manager) GetStore(name string) (VectorStore, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	store, exists := m.stores[name]
	if !exists {
		return nil, fmt.Errorf("vector store '%s' not found", name)
	}

	return store, nil
}

// ListStores 获取所有可用的向量存储名称
func (m *Manager) ListStores() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.stores))
	for name := range m.stores {
		names = append(names, name)
	}

	return names
}

// RemoveStore 移除指定的向量存储
func (m *Manager) RemoveStore(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.stores, name)
}
