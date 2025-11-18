package storage

import (
	"hash/fnv"
	"sync"
)

type shard struct {
	data map[string]string
	mu   sync.RWMutex
}

type Memory struct {
	shards    []*shard
}

const numShards = 8

// Temp memory before db implementation
func NewMemory() *Memory {
	m := &Memory{}
	
	for range numShards {
		m.shards = append(m.shards, &shard{
			data: make(map[string]string),
		})
	}
	
	return m
}

func (m *Memory) getShard(key string) *shard {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	return m.shards[hash.Sum32()%numShards]
}

func (m *Memory) Get(key string) (string, bool) {
	shard := m.getShard(key)

	shard.mu.RLock()
	defer shard.mu.RUnlock()

	value, exists := shard.data[key]

	return value, exists
}

func (m *Memory) Set(key, value string) error {
	shard := m.getShard(key)
	
	shard.mu.Lock()
	defer shard.mu.Unlock()

	shard.data[key] = value

	return nil
}

func (m *Memory) Exists(key string) bool {
	shard := m.getShard(key)

	shard.mu.RLock()
	defer shard.mu.RUnlock()

	_, exists := shard.data[key]

	return exists
}

var _ Store = (*Memory)(nil)