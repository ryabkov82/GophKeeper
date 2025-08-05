package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// HashPassword принимает пароль и возвращает хеш и соль
func HashPassword(password string) (hash, salt string, err error) {
	if password == "" {
		return "", "", errors.New("password is empty")
	}

	// Генерируем соль — 16 байт случайных данных
	saltBytes := make([]byte, 16)
	_, err = rand.Read(saltBytes)
	if err != nil {
		return "", "", err
	}
	salt = base64.StdEncoding.EncodeToString(saltBytes)

	// Хешируем пароль + соль
	hashBytes := sha256.Sum256([]byte(password + salt))
	hash = base64.StdEncoding.EncodeToString(hashBytes[:])

	return hash, salt, nil
}

// VerifyPassword проверяет пароль по хешу и соли
func VerifyPassword(password, hash, salt string) bool {
	hashBytes := sha256.Sum256([]byte(password + salt))
	computedHash := base64.StdEncoding.EncodeToString(hashBytes[:])
	return computedHash == hash
}
