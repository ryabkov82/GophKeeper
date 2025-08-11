package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// EncryptAESGCM шифрует переданные данные plaintext с помощью ключа key,
// используя алгоритм AES-GCM.
//
// Возвращает зашифрованный с nonce в префиксе срез байт или ошибку.
//
// Ключ должен быть длины 16, 24 или 32 байта (AES-128/192/256).
func EncryptAESGCM(plaintext, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("invalid AES key size")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptAESGCM расшифровывает данные ciphertext с помощью ключа key,
// ожидая, что nonce записан в начале ciphertext.
//
// Возвращает расшифрованные данные или ошибку.
func DecryptAESGCM(ciphertext, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("invalid AES key size")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce := ciphertext[:gcm.NonceSize()]
	ct := ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, nil)
}
