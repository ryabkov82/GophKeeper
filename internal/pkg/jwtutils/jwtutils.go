package jwtutils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

// New создает новый менеджер токенов
func New(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// GenerateToken создает JWT для указанного пользователя
func (tm *TokenManager) GenerateToken(userID, login string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"login": login,
		"exp":   time.Now().Add(tm.ttl).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secret)
}
