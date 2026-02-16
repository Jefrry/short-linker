package service

import (
	"context"
	"fmt"
	"strings"
	"time"

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

func (s *LinkDataService) CreateShortLink(ctx context.Context, originalURL string, userID int64) (model.LinkItem, error) {
	id, err := s.generateUniqueID(ctx)
	if err != nil {
		return model.LinkItem{}, fmt.Errorf("failed to generate unique ID: %w", err)
	}
	link := model.LinkItem{
		ID:          id,
		OriginalURL: originalURL,
		ShortURL:    s.buildShortLink(id),
		UserID:      userID,
	}
	_, err = s.repo.Save(ctx, link)
	if err != nil {
		return model.LinkItem{}, fmt.Errorf("failed to store link: %w", err)
	}

	return link, nil
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

func (s *LinkDataService) GetOriginalURL(ctx context.Context, id string) (model.LinkItem, error) {
	linkItem, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.LinkItem{}, fmt.Errorf("failed to retrieve link: %w", err)
	}
	linkItem.ShortURL = s.buildShortLink(linkItem.ID)
	return linkItem, nil
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

func (s *LinkDataService) RecordVisit(ctx context.Context, visit model.Visit) {
	if err := s.repo.RecordVisit(ctx, visit); err != nil {
		fmt.Printf("failed to record visit for link %s: %v\n", visit.LinkID, err)
	}
}

func (s *LinkDataService) GetLinkMetrics(ctx context.Context, linkID string, userID int64, from, to time.Time) ([]model.Visit, error) {
	isOwner, err := s.repo.IsOwner(ctx, linkID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check ownership: %w", err)
	}
	if !isOwner {
		return nil, fmt.Errorf("access denied: user %d is not the owner of link %s", userID, linkID)
	}

	visits, err := s.repo.GetVisits(ctx, linkID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get visits: %w", err)
	}

	return visits, nil
}
