package storage

import (
	"hash/fnv"
	"sync"
	"time"
)

type shard struct {
	data map[string]string
	ttl  map[string]time.Time
	mu   sync.RWMutex
}

func newShard() *shard {
	return &shard{
		data: make(map[string]string),
		ttl:  make(map[string]time.Time),
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

	if exp, ok := s.ttl[key]; ok && time.Now().After(exp) {
			delete(s.data, key)
			delete(s.ttl, key)
			return "", false
	}

	v, ok := s.data[key]
	return v, ok
}

func (s *shard) Set(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
	s.ttl[key] = time.Now().Add(ttl)
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

func (m *Memory) Set(key, value string, ttl time.Duration) error {
	shardIndex := m.getShardIndex(key)

	if ttl <= 0 {
		ttl = time.Hour * 24 * 365 * 100
	}

	m.shards[shardIndex].Set(key, value, ttl)

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
