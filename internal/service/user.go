package service

import (
	"golang.org/x/crypto/bcrypt"

	"short-linker/internal/model"
	"short-linker/internal/repository"
)

type UserDataService struct {
	repo         repository.UserRepository
	tokenService TokenService
}

func NewUserService(repo repository.UserRepository, tokenService TokenService) *UserDataService {
	return &UserDataService{
		repo:         repo,
		tokenService: tokenService,
	}
}

func (s *UserDataService) Signup(data model.SignupPayload) (model.User, error) {
	user := model.User{
		Name:  data.Name,
		Email: data.Email,
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

func (s *UserDataService) Signin(email, password string) (string, error) {
	user, err := s.repo.GetByEmail(email)
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