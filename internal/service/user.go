package service

import (
	"context"
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
