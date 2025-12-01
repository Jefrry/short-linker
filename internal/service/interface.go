package service

import (
	"short-linker/internal/model"
)

type LinkService interface {
	CreateShortLink(originalURL string) (string, error)
	CreateShortLinkBatch(items []model.LinkBatchPayload) ([]model.LinkBatchResponse, error)
	GetOriginalURL(id string) (string, error)
}