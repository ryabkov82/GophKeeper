package storage

import (
	clientcrypto "github.com/ryabkov82/gophkeeper/internal/client/crypto"
)

// TokenStorage описывает интерфейс для хранения токена авторизации.
type TokenStorage interface {
	// Save сохраняет токен.
	Save(token string) error

	// Load загружает токен.
	// Возвращает ошибку, если токен отсутствует или недоступен.
	Load() (string, error)

	// Clear удаляет сохранённый токен.
	Clear() error
}

// CryptoKeyStorage описывает интерфейс для хранения ключа шифрования и параметров KDF.
type CryptoKeyStorage interface {
	// Save сохраняет ключ шифрования и параметры KDF.
	Save(key []byte, params clientcrypto.Argon2Params) error

	// Load загружает ключ шифрования и параметры KDF.
	// Возвращает ошибку, если ключ отсутствует или недоступен.
	Load() ([]byte, clientcrypto.Argon2Params, error)

	// Clear удаляет сохранённый ключ шифрования.
	Clear() error
}
