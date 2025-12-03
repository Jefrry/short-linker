package service

import (
	"short-linker/internal/model"
)

type LinkService interface {
	CreateShortLink(originalURL string) (string, error)
	CreateShortLinkBatch(items []model.LinkBatchPayload) ([]model.LinkBatchResponse, error)
	GetOriginalURL(id string) (string, error)
}

type UserService interface {
	Signup(data model.SignupPayload) (model.User, error)
	Signin(email, password string) (string, error)
	GetProfile(userID int64) (model.User, error)
}

type TokenService interface {
	GenerateToken(userID int64) (string, error)
}