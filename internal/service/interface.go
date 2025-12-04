package service

import (
	"context"
	"short-linker/internal/model"
)

type LinkService interface {
	CreateShortLink(ctx context.Context, originalURL string) (string, error)
	CreateShortLinkBatch(ctx context.Context, items []model.LinkBatchPayload) ([]model.LinkBatchResponse, error)
	GetOriginalURL(ctx context.Context, id string) (string, error)
}

type UserService interface {
	Signup(ctx context.Context, data model.SignupPayload) (model.User, error)
	Signin(ctx context.Context, email, password string) (string, error)
	GetProfile(ctx context.Context, userID int64) (model.User, error)
}

type TokenService interface {
	GenerateToken(userID int64) (string, error)
}
