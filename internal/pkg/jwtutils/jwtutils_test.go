package jwtutils_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtutils"
	"github.com/stretchr/testify/assert"
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

func TestParseToken(t *testing.T) {
	secret := "mysecretkey"
	tm := jwtutils.New(secret, time.Minute*10)

	userID := "user123"
	login := "testlogin"

	// Генерируем корректный токен
	tokenStr, err := tm.GenerateToken(userID, login)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	// Парсим корректный токен
	claims, err := tm.ParseToken(tokenStr)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims["sub"])
	assert.Equal(t, login, claims["login"])

	// Парсим токен с неправильной подписью (подделанный токен)
	badToken := tokenStr + "invalid"
	_, err = tm.ParseToken(badToken)
	assert.Error(t, err)

	// Парсим пустую строку
	_, err = tm.ParseToken("")
	assert.Error(t, err)

	// Парсим токен с истекшим сроком (создадим вручную с отрицательным временем exp)
	expiredToken := createExpiredToken(t, secret, userID, login)
	_, err = tm.ParseToken(expiredToken)
	assert.Error(t, err)
}

// createExpiredToken создает JWT токен с истекшим сроком (для теста).
func createExpiredToken(t *testing.T, secret, userID, login string) string {
	claims := map[string]interface{}{
		"sub":   userID,
		"login": login,
		"exp":   time.Now().Add(-time.Hour).Unix(), // истекший час назад
		"iat":   time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwtutils.New(secret, time.Minute*10)
	jwtToken := jwtTokenWithClaims(t, token, claims)
	return jwtToken
}

func jwtTokenWithClaims(t *testing.T, tm *jwtutils.TokenManager, claims map[string]interface{}) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	tokenStr, err := token.SignedString(tm.Secret())
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenStr
}
