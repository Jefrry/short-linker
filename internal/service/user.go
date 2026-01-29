package service

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"short-linker/internal/model"
	"short-linker/internal/repository"
)

type UserDataService struct {
	tokenService TokenService

	userRepo repository.UserRepository
	linkRepo repository.LinkRepository
}

func NewUserService(tokenService TokenService, userRepo repository.UserRepository, linkRepo repository.LinkRepository) *UserDataService {
	return &UserDataService{
		tokenService: tokenService,

		userRepo: userRepo,
		linkRepo: linkRepo,
	}
}

func (s *UserDataService) Signup(ctx context.Context, data model.SignupPayload) (model.User, error) {
	user := model.User{
		Name:  data.Name,
		Email: data.Email,
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}
	user.Password = string(hash)

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return model.User{}, err
	}

	return createdUser, nil
}

func (s *UserDataService) Signin(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", err
	}

	token, err := s.tokenService.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserDataService) GetProfile(ctx context.Context, userID int64) (model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *UserDataService) GetLinks(ctx context.Context, userID int64) ([]model.LinkItem, error) {
	return s.linkRepo.GetByUserID(ctx, userID)
}

func (s *UserDataService) DeleteLinks(ctx context.Context, links []string, userID int64) error {
	results := make(chan string, len(links))

	var wg = sync.WaitGroup{}

	for _, link := range links {
		wg.Add(1)

		go func(l string) {
			defer wg.Done()

			isOwner, err := s.linkRepo.IsOwner(ctx, l, userID)
			if err != nil || !isOwner {
				return
			}

			results <- l
		}(link)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var toDelete []string
	for link := range results {
		toDelete = append(toDelete, link)
	}

	if len(toDelete) > 0 {
		if err := s.linkRepo.MarkAsDeleted(ctx, toDelete); err != nil {
			return fmt.Errorf("failed to mark links as deleted: %w", err)
		}
	}

	return nil
}
