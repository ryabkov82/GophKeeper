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

// EncryptStream шифрует поток данных из r и пишет в w.
func EncryptStream(r io.Reader, w io.Writer, key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	if _, err := w.Write(nonce); err != nil {
		return err
	}

	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			ct := gcm.Seal(nil, nonce, buf[:n], nil)
			if _, err := w.Write(ct); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// DecryptStream расшифровывает поток данных из r и пишет результат в w.
func DecryptStream(r io.Reader, w io.Writer, key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(r, nonce); err != nil {
		return err
	}

	buf := make([]byte, 32*1024+gcm.Overhead())
	for {
		n, err := r.Read(buf)
		if n > 0 {
			plain, err := gcm.Open(nil, nonce, buf[:n], nil)
			if err != nil {
				return err
			}
			if _, err := w.Write(plain); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}
