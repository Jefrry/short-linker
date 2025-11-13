package service

import (
	"errors"
	"strings"

	"short-linker/internal/repository"
	"short-linker/pkg"
)

type LinkService interface {
	CreateShortLink(originalURL string) (string, error)
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
	var id string
	for {
		var err error
		id, err = pkg.RandomStringDefault()
		if err != nil {
			return "", errors.New("failed to generate short link")
		}

		if !s.repo.Exists(id) {
			err = s.repo.Save(id, originalURL)
			if err != nil {
				return "", errors.New("failed to store link")
			}
			break
		}
	}

	shortLink := strings.TrimRight(s.baseHost, "/") + "/" + id

	return shortLink, nil
}

func (s *LinkDataService) GetOriginalURL(id string) (string, error) {
	originalURL, err := s.repo.Get(id)
	if err != nil {
		return "", errors.New("link not found")
	}
	return originalURL, nil
}