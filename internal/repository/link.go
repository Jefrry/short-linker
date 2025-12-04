package repository

import (
	"context"
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

func (r *LinkDataRepository) Save(ctx context.Context, item model.LinkItem) (string, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	const query = `
        INSERT INTO links (id, original_url)
		VALUES ($1, $2)
		ON CONFLICT (original_url) DO UPDATE
			SET original_url = EXCLUDED.original_url
		RETURNING links.id
    `

	var savedID string
	err = tx.QueryRowContext(ctx, query, item.ID, item.OriginalURL).Scan(&savedID)
	if err != nil {
		return "", err
	}

	if savedID != item.ID {
		return savedID, model.ErrOriginalURLExists
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	r.storage.Set(item.ID, item.OriginalURL, 0)

	return item.ID, nil
}

func (r *LinkDataRepository) SaveBatch(ctx context.Context, items []model.LinkItem) error {
	if len(items) == 0 {
		return fmt.Errorf("no items to save")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

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

	query += " ON CONFLICT (original_url) DO NOTHING"

	_, err = tx.ExecContext(ctx, query, values...)
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

func (r *LinkDataRepository) Get(ctx context.Context, id string) (string, error) {
	var originalURL string

	if url, ok := r.storage.Get(id); ok && url != "" {
		originalURL = url
		return originalURL, nil
	}

	err := r.db.QueryRowContext(ctx, "SELECT original_url FROM links WHERE id = $1", id).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("link not found")
		}
		return "", err
	}

	return originalURL, nil
}

func (r *LinkDataRepository) Exists(ctx context.Context, id string) bool {
	var exists bool

	if r.storage.Exists(id) {
		return true
	}

	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM links WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}
