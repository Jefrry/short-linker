package storage

type Store interface {
	Get(key string) (string, bool)
	Set(key, value string) error
	Exists(key string) bool
}