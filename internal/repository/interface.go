package repository

import "short-linker/internal/model"

type LinkRepository interface {
	Save(item model.LinkItem) (string, error)
	SaveBatch(item []model.LinkItem) error
	Get(id string) (string, error)
	Exists(id string) bool
}

type UserRepository interface {
	Create(user model.User) (model.User, error)
	GetByEmail(email string) (model.User, error)
}