package repository

import (
	"context"

	"short-linker/internal/model"
)

type LinkRepository interface {
	Save(ctx context.Context, item model.LinkItem) (string, error)
	SaveBatch(ctx context.Context, item []model.LinkItem) error
	Get(ctx context.Context, id string) (string, error)
	Exists(ctx context.Context, id string) bool
	GetByUserID(ctx context.Context, userID int64) ([]model.LinkItem, error)
	IsOwner(ctx context.Context, id string, userID int64) (bool, error)
	MarkAsDeleted(ctx context.Context, ids []string) error
}

type UserRepository interface {
	Create(ctx context.Context, user model.User) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
	GetByID(ctx context.Context, userID int64) (model.User, error)
}
