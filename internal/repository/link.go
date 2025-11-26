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
	storage *storage.Memory // TODO: add ttl to memory storage and use it
	db *sql.DB
}

func NewLinkRepository(storage *storage.Memory, db *sql.DB) *LinkDataRepository {
	return &LinkDataRepository{
		storage: storage,
		db: db,
	}
}

func (r *LinkDataRepository) Save(id string, originalURL string) error {
	_, err := r.db.Exec("INSERT INTO links (id, original_url) VALUES ($1, $2)", id, originalURL)

	return err
}

func (r *LinkDataRepository) Get(id string) (string, error) {
	var originalURL string

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

	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM links WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}