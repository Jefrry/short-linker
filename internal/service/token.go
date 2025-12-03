package service

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
)

type TokenDataService struct {
	secrectKey []byte
}

func NewTokenService(secretKey string) *TokenDataService {
	return &TokenDataService{
		secrectKey: []byte(secretKey),
	}
}

func (s *TokenDataService) GenerateToken(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(s.secrectKey)
}