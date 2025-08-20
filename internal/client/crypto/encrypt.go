package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
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

// размер исходного (plaintext) чанка
const chunkSize = 32 * 1024 // 32 KiB

// EncryptStream шифрует r -> w при помощи AES-GCM.
// Формат: [baseNonce(12)] { [len(ct):u32][ct] }*
func EncryptStream(r io.Reader, w io.Writer, key []byte) error {

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if nonceSize < 8 {
		return fmt.Errorf("nonce too small: %d", nonceSize)
	}

	// общий базовый nonce
	base := make([]byte, nonceSize)
	if _, err := rand.Read(base); err != nil {
		return fmt.Errorf("rand base nonce: %w", err)
	}
	if _, err := w.Write(base); err != nil {
		return fmt.Errorf("write base nonce: %w", err)
	}

	buf := make([]byte, chunkSize)
	var index uint64 = 0

	for {
		n, rerr := r.Read(buf)
		if n > 0 {
			// производим nonce_i = base || counter(index) в последних 8 байтах
			nonce := make([]byte, nonceSize)
			copy(nonce, base)
			binary.BigEndian.PutUint64(nonce[nonceSize-8:], index)

			ct := gcm.Seal(nil, nonce, buf[:n], nil)

			// пишем длину и сам шифрокусок
			if len(ct) > int(^uint32(0)) {
				return fmt.Errorf("ciphertext too large")
			}
			var clen = uint32(len(ct))
			if err := binary.Write(w, binary.BigEndian, clen); err != nil {
				return fmt.Errorf("write clen: %w", err)
			}
			if _, err := w.Write(ct); err != nil {
				return fmt.Errorf("write ct: %w", err)
			}

			index++
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return fmt.Errorf("read plaintext: %w", rerr)
		}
	}

	return nil
}

// DecryptStream расшифровывает поток формата EncryptStream.
func DecryptStream(r io.Reader, w io.Writer, key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("gcm: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if nonceSize < 8 {
		return fmt.Errorf("nonce too small: %d", nonceSize)
	}

	// читаем общий базовый nonce
	base := make([]byte, nonceSize)
	if _, err := io.ReadFull(r, base); err != nil {
		if err == io.EOF {
			return nil // пустой поток
		}
		return fmt.Errorf("read base nonce: %w", err)
	}

	var index uint64 = 0
	for {
		// читаем длину зашифрованного чанка
		var clen uint32
		if err := binary.Read(r, binary.BigEndian, &clen); err != nil {
			if err == io.EOF {
				break // конец потока
			}
			return fmt.Errorf("read clen: %w", err)
		}
		if clen == 0 {
			// допустим пустой чанк и просто идём дальше
			index++
			continue
		}

		ct := make([]byte, clen)
		if _, err := io.ReadFull(r, ct); err != nil {
			return fmt.Errorf("read ct: %w", err)
		}

		nonce := make([]byte, nonceSize)
		copy(nonce, base)
		binary.BigEndian.PutUint64(nonce[nonceSize-8:], index)

		pt, err := gcm.Open(nil, nonce, ct, nil)
		if err != nil {
			return fmt.Errorf("open: %w", err) // тут раньше падало "cipher: message authentication failed"
		}
		if _, err := w.Write(pt); err != nil {
			return fmt.Errorf("write plaintext: %w", err)
		}

		index++
	}

	return nil
}
