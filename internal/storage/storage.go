package storage

import (
	"hash/fnv"
	"sync"
)

type shard struct {
	data map[string]string
	mu   sync.RWMutex
}

func newShard() *shard {
	return &shard{
		data: make(map[string]string),
	}
}

type Memory struct {
	shards []*shard
}

const numShards = 8

func (m *Memory) getShardIndex(key string) int {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key))

	return int(hash.Sum32() % uint32(len(m.shards)))
}

func (s *shard) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	return v, ok
}

func (s *shard) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
}

func (s *shard) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[key]
	return ok
}

func (m *Memory) Get(key string) (string, bool) {
	shardIndex := m.getShardIndex(key)

	return m.shards[shardIndex].Get(key)
}

func (m *Memory) Set(key, value string) error {
	shardIndex := m.getShardIndex(key)
	m.shards[shardIndex].Set(key, value)

	return nil
}

func (m *Memory) Exists(key string) bool {
	shardIndex := m.getShardIndex(key)
	return m.shards[shardIndex].Exists(key)
}

// Temp memory before db implementation
func NewMemory() *Memory {
	m := &Memory{
		shards: make([]*shard, numShards),
	}

	for i := range numShards {
		m.shards[i] = newShard()
	}

	return m
}

var _ Store = (*Memory)(nil)