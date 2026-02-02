package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

	// ISSUE: there is a business issue here,
	// original_url should be unique per user,
	// but current implementation makes it unique globally.
	// So if two different users shorten the same URL,
	// they will get the same short link and user_id will be first user.
	const query = `
        INSERT INTO links (id, original_url, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (original_url) DO UPDATE
			SET original_url = EXCLUDED.original_url
		RETURNING links.id
    `

	var userID any
	if item.UserID == 0 {
		userID = nil
	} else {
		userID = item.UserID
	}

	var savedID string
	err = tx.QueryRowContext(ctx, query, item.ID, item.OriginalURL, userID).Scan(&savedID)
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

	var queryBuilder strings.Builder
	queryBuilder.WriteString("INSERT INTO links (id, original_url, user_id) VALUES ")
	values := make([]any, 0, len(items)*3)

	for i, item := range items {
		if i > 0 {
			queryBuilder.WriteString(",")
		}

		n := i*3 + 1
		fmt.Fprintf(&queryBuilder, "($%d, $%d, $%d)", n, n+1, n+2)

		var userID any
		if item.UserID == 0 {
			userID = nil
		} else {
			userID = item.UserID
		}

		values = append(values, item.ID, item.OriginalURL, userID)
	}

	queryBuilder.WriteString(" ON CONFLICT (original_url) DO NOTHING")

	_, err = tx.ExecContext(ctx, queryBuilder.String(), values...)
	if err != nil {
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

func (r *LinkDataRepository) Get(ctx context.Context, id string) (model.LinkItem, error) {
	linkItem := model.LinkItem{}

	if url, ok := r.storage.Get(id); ok && url != "" {
		linkItem.OriginalURL = url
		return linkItem, nil
	}

	err := r.db.QueryRowContext(ctx, "SELECT id, original_url, user_id, deleted FROM links WHERE id = $1", id).Scan(&linkItem.ID, &linkItem.OriginalURL, &linkItem.UserID, &linkItem.Deleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LinkItem{}, fmt.Errorf("link not found")
		}
		return model.LinkItem{}, err
	}

	return linkItem, nil
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

func (r *LinkDataRepository) GetByUserID(ctx context.Context, userID int64) ([]model.LinkItem, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	const query = `
		SELECT id, original_url, user_id
		FROM links
		WHERE user_id = $1
	`

	rows, err := tx.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.LinkItem
	for rows.Next() {
		var item model.LinkItem
		err := rows.Scan(
			&item.ID,
			&item.OriginalURL,
			&item.UserID,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *LinkDataRepository) IsOwner(ctx context.Context, id string, userID int64) (bool, error) {
	var isOwner bool

	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM links WHERE id = $1 AND user_id = $2)
	`, id, userID).Scan(&isOwner)

	if err != nil {
		fmt.Printf("IsOwner error for id %q, userID %d: %v\n", id, userID, err)
		return false, err
	}

	return isOwner, nil
}

func (r *LinkDataRepository) MarkAsDeleted(ctx context.Context, ids []string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE links SET deleted = TRUE WHERE id = ANY($1) AND deleted = FALSE", ids)
	if err != nil {
		return err
	}

	for _, id := range ids {
		r.storage.Delete(id)
	}

	return nil
}
