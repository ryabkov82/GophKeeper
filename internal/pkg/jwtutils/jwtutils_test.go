package jwtutils_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtutils"
)

func TestGenerateToken(t *testing.T) {
	secret := "supersecretkey"
	ttl := time.Minute * 10
	tm := jwtutils.New(secret, ttl)

	userID := "12345"
	login := "testuser"

	tokenStr, err := tm.GenerateToken(userID, login)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем, что токен не пустой
	if tokenStr == "" {
		t.Fatal("token string is empty")
	}

	// Парсим и проверяем содержимое токена
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if !token.Valid {
		t.Fatal("token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to get claims")
	}

	// Проверяем userID и login
	if claims["sub"] != userID {
		t.Errorf("expected sub %s, got %v", userID, claims["sub"])
	}
	if claims["login"] != login {
		t.Errorf("expected login %s, got %v", login, claims["login"])
	}

	// Проверяем время жизни токена (exp и iat)
	exp, okExp := claims["exp"].(float64)
	iat, okIat := claims["iat"].(float64)
	if !okExp || !okIat {
		t.Fatal("exp or iat claim is missing or has wrong type")
	}

	expTime := time.Unix(int64(exp), 0)
	iatTime := time.Unix(int64(iat), 0)
	if expTime.Sub(iatTime) != ttl {
		t.Errorf("expected ttl %v, got %v", ttl, expTime.Sub(iatTime))
	}
}
