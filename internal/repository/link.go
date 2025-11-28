package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"short-linker/internal/storage"
)

type LinkRepository interface {
	Save(id string, originalURL string) error
	Get(id string) (string, error)
	Exists(id string) bool
}

type LinkDataRepository struct {
	storage *storage.Memory
	db *sql.DB
}

func NewLinkRepository(storage *storage.Memory, db *sql.DB) *LinkDataRepository {
	return &LinkDataRepository{
		storage: storage,
		db: db,
	}
}

func (r *LinkDataRepository) Save(id string, originalURL string) error {
	r.storage.Set(id, originalURL, 0)
	_, err := r.db.Exec("INSERT INTO links (id, original_url) VALUES ($1, $2)", id, originalURL)

	return err
}

func (r *LinkDataRepository) Get(id string) (string, error) {
	var originalURL string

	if url, ok := r.storage.Get(id); ok && url != "" {
		originalURL = url
		return originalURL, nil
	}

	err := r.db.QueryRow("SELECT original_url FROM links WHERE id = $1", id).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("link not found")
		}
		return "", err
	}

	return originalURL, nil
}

func (r *LinkDataRepository) Exists(id string) bool {
	var exists bool

	if r.storage.Exists(id) {
		return true
	}

	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM links WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}