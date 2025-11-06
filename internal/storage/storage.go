package storage

import "sync"

type Memory struct {
	data map[string]string
	mu   sync.RWMutex
}

// Temp memory before db implementation
func NewMemory() *Memory {
	return &Memory{
		data: make(map[string]string),
	}
}

func (m *Memory) Get(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, exists := m.data[key]

	return value, exists
}

func (m *Memory) Set(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value

	return nil
}

func (m *Memory) Exists(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.data[key]

	return exists
}

var _ Store = (*Memory)(nil)