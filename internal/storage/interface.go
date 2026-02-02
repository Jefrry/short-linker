package storage

import "time"

type Store interface {
	Get(key string) (string, bool)
	Set(key, value string, ttl time.Duration) error
	Exists(key string) bool
	Delete(key string) error
}
