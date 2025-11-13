package repository

import (
	"errors"

	"short-linker/internal/storage"
)

type LinkRepository interface {
	Save(id string, originalURL string) error
	Get(id string) (string, error)
	Exists(id string) bool
}

type LinkDataRepository struct {
	storage *storage.Memory
}

func NewLinkRepository(storage *storage.Memory) *LinkDataRepository {
	return &LinkDataRepository{
		storage: storage,
	}
}

func (r *LinkDataRepository) Save(id string, originalURL string) error {
	return r.storage.Set(id, originalURL)
}

func (r *LinkDataRepository) Get(id string) (string, error) {
	originalURL, exists := r.storage.Get(id)
	if !exists {
		return "", errors.New("link not found")
	}
	return originalURL, nil
}

func (r *LinkDataRepository) Exists(id string) bool {
	return r.storage.Exists(id)
}