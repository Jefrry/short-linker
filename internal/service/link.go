package service

import (
	"context"
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

func (s *LinkDataService) CreateShortLink(ctx context.Context, originalURL string, userID int64) (string, error) {
	id, err := s.generateUniqueID(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to generate unique ID: %w", err)
	}
	id, err = s.repo.Save(ctx, model.LinkItem{
		ID:          id,
		OriginalURL: originalURL,
		UserID:      userID,
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

func (s *LinkDataService) CreateShortLinkBatch(ctx context.Context, items []model.LinkBatchPayload, userID int64) ([]model.LinkBatchResponse, error) {
	resItems := make([]model.LinkBatchResponse, 0, len(items))
	batchItems := make([]model.LinkItem, 0, len(items))

	for _, item := range items {
		if item.URL == "" {
			continue
		}

		id, err := s.generateUniqueID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate unique ID for batch item: %w", err)
		}

		batchItems = append(batchItems, model.LinkItem{
			ID:          id,
			OriginalURL: item.URL,
			UserID:      userID,
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

	if err := s.repo.SaveBatch(ctx, batchItems); err != nil {
		return nil, fmt.Errorf("failed to store link batch: %w", err)
	}

	return resItems, nil
}

func (s *LinkDataService) GetOriginalURL(ctx context.Context, id string) (string, error) {
	originalURL, err := s.repo.Get(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve link: %w", err)
	}
	return originalURL, nil
}

func (s *LinkDataService) generateUniqueID(ctx context.Context) (string, error) {
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

		if !s.repo.Exists(ctx, id) {
			return id, nil
		}
	}
}

func (s *LinkDataService) buildShortLink(id string) string {
	return strings.TrimRight(s.baseHost, "/") + "/" + id
}
