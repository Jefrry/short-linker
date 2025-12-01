package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"short-linker/internal/model"
	"short-linker/internal/storage"
)

type LinkDataRepository struct {
	storage *storage.Memory
	db      *sql.DB
}

func NewLinkRepository(storage *storage.Memory, db *sql.DB) *LinkDataRepository {
	return &LinkDataRepository{
		storage: storage,
		db:      db,
	}
}

func (r *LinkDataRepository) Save(items []model.LinkItem) error {
	if len(items) == 0 {
		return fmt.Errorf("no items to save")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := "INSERT INTO links (id, original_url) VALUES "
	values := []any{}

	for i, item := range items {
		if i > 0 {
			query += ","
		}

		n := i*2 + 1

		query += fmt.Sprintf("($%d, $%d)", n, n+1)

		values = append(values, item.ID, item.OriginalURL)
	}

	_, err = tx.Exec(query, values...)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	for _, item := range items {
		r.storage.Set(item.ID, item.OriginalURL, 0)
	}

	return nil
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
