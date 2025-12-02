package service

import (
	"errors"
	"fmt"
	"strings"

	"short-linker/internal/model"
	"short-linker/internal/repository"
	"short-linker/pkg"
)

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
		return "", fmt.Errorf("failed to generate unique ID: %w", err)
	}
	id, err = s.repo.Save(model.LinkItem{
		ID:          id,
		OriginalURL: originalURL,
	})
	if err != nil && errors.Is(err, model.ErrOriginalURLExists) {
		return s.buildShortLink(id), model.ErrOriginalURLExists
	}
	if err != nil {
		return "", fmt.Errorf("failed to store link: %w", err)
	}

	shortLink := s.buildShortLink(id)
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
			return nil, fmt.Errorf("failed to generate unique ID for batch item: %w", err)
		}

		batchItems = append(batchItems, model.LinkItem{
			ID:          id,
			OriginalURL: item.URL,
		})

		shortLink := s.buildShortLink(id)
		resItems = append(resItems, model.LinkBatchResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortLink,
		})
	}

	if len(batchItems) == 0 {
		return []model.LinkBatchResponse{}, nil
	}

	if err := s.repo.SaveBatch(batchItems); err != nil {
		return nil, fmt.Errorf("failed to store link batch: %w", err)
	}

	return resItems, nil
}

func (s *LinkDataService) GetOriginalURL(id string) (string, error) {
	originalURL, err := s.repo.Get(id)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve link: %w", err)
	}
	return originalURL, nil
}

func (s *LinkDataService) generateUniqueID() (string, error) {
	retries := 0
	const maxRetries = 5
	for {
		if retries >= maxRetries {
			return "", fmt.Errorf("failed to generate unique short link after %d attempts", maxRetries)
		}
		retries++

		id, err := pkg.RandomStringDefault()
		if err != nil {
			return "", fmt.Errorf("failed to generate random string: %w", err)
		}

		if !s.repo.Exists(id) {
			return id, nil
		}
	}
}

func (s *LinkDataService) buildShortLink(id string) string {
	return strings.TrimRight(s.baseHost, "/") + "/" + id
}
