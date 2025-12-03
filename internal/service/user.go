package service

import (
	"golang.org/x/crypto/bcrypt"

	"short-linker/internal/model"
	"short-linker/internal/repository"
)

type UserDataService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserDataService {
	return &UserDataService{
		repo: repo,
	}
}

func (s *UserDataService) Signup(data model.SignupPayload) (model.User, error) {
	user := model.User{
		Name:     data.Name,
		Email:    data.Email,
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}
	user.Password = string(hash)

	createdUser, err := s.repo.Create(user)
	if err != nil {
		return model.User{}, err
	}

	return createdUser, nil
}