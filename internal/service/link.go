package service

import (
	"errors"
	"strings"

	"short-linker/internal/model"
	"short-linker/internal/repository"
	"short-linker/pkg"
)

type LinkService interface {
	CreateShortLink(originalURL string) (string, error)
	CreateShortLinkBatch(items []model.LinkBatchPayload) ([]model.LinkBatchResponse, error)
	GetOriginalURL(id string) (string, error)
}

type LinkDataService struct {
	repo     repository.LinkRepository
	baseHost string
}

func NewLinkService(repo repository.LinkRepository, baseHost string) *LinkDataService {
	return &LinkDataService{
		repo:     repo,
		baseHost: baseHost,
	}
}

func (s *LinkDataService) CreateShortLink(originalURL string) (string, error) {
	id, err := s.generateUniqueID()
	if err != nil {
		return "", err
	}

	if err := s.repo.Save([]model.LinkItem{{ID: id, OriginalURL: originalURL}}); err != nil {
		return "", errors.New("failed to store link")
	}

	// TODO: move to a separate function
	shortLink := strings.TrimRight(s.baseHost, "/") + "/" + id
	return shortLink, nil
}

func (s *LinkDataService) CreateShortLinkBatch(items []model.LinkBatchPayload) ([]model.LinkBatchResponse, error) {
	resItems := make([]model.LinkBatchResponse, 0, len(items))
    batchItems := make([]model.LinkItem, 0, len(items))

	for _, item := range items {
		if item.URL == "" {
			continue
		}

		id, err := s.generateUniqueID()
		if err != nil {
			return nil, err
		}

		batchItems = append(batchItems, model.LinkItem{
			ID:          id,
			OriginalURL: item.URL,
		})
		
		shortLink := strings.TrimRight(s.baseHost, "/") + "/" + id
		resItems = append(resItems, model.LinkBatchResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortLink,
		})
	}

	if len(batchItems) == 0 {
        return []model.LinkBatchResponse{}, nil
    }

    if err := s.repo.Save(batchItems); err != nil {
        return nil, errors.New("failed to store link batch")
    }

	return resItems, nil
}

func (s *LinkDataService) GetOriginalURL(id string) (string, error) {
	originalURL, err := s.repo.Get(id)
	if err != nil {
		return "", errors.New("link not found")
	}
	return originalURL, nil
}

func (s *LinkDataService) generateUniqueID() (string, error) {
	retries := 0
	const maxRetries = 5
	for {
		if retries >= maxRetries {
			return "", errors.New("failed to generate unique short link after multiple attempts")
		}
		retries++

		id, err := pkg.RandomStringDefault()
		if err != nil {
			return "", errors.New("failed to generate short link")
		}

		if !s.repo.Exists(id) {
			return id, nil
		}
	}
}