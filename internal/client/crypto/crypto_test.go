package crypto_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/crypto"
	"github.com/stretchr/testify/assert"
)

func TestDeriveKey_Success(t *testing.T) {
	password := "testpassword"
	salt := []byte("random_salt_123456")

	key, params, err := crypto.DeriveKey(password, salt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(key) != int(params.KeyLen) {
		t.Errorf("expected key length %d, got %d", params.KeyLen, len(key))
	}
}

func TestDeriveKey_EmptySalt(t *testing.T) {
	_, _, err := crypto.DeriveKey("password", []byte{})
	if err == nil {
		t.Fatal("expected error for empty salt, got nil")
	}
}

func TestDeriveKey_Deterministic(t *testing.T) {
	password := "same_password"
	salt := []byte("same_salt")

	key1, _, err1 := crypto.DeriveKey(password, salt)
	if err1 != nil {
		t.Fatalf("unexpected error: %v", err1)
	}

	key2, _, err2 := crypto.DeriveKey(password, salt)
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("expected same keys for same password and salt, got different")
	}
}

func TestDeriveKey_DifferentSalt(t *testing.T) {
	password := "password"
	salt1 := []byte("salt_one")
	salt2 := []byte("salt_two")

	key1, _, _ := crypto.DeriveKey(password, salt1)
	key2, _, _ := crypto.DeriveKey(password, salt2)

	if bytes.Equal(key1, key2) {
		t.Error("expected different keys for different salts, got same")
	}
}

func TestEncryptDecryptAESGCM_Success(t *testing.T) {
	key := []byte("0123456789ABCDEF0123456789ABCDEF") // 32 байта — AES-256
	plaintext := []byte("Hello, AES-GCM!")

	ciphertext, err := crypto.EncryptAESGCM(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptAESGCM returned error: %v", err)
	}

	decrypted, err := crypto.DecryptAESGCM(ciphertext, key)
	if err != nil {
		t.Fatalf("DecryptAESGCM returned error: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted data mismatch: got %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptAESGCM_InvalidKeySize(t *testing.T) {
	_, err := crypto.EncryptAESGCM([]byte("data"), []byte("short"))
	if err == nil {
		t.Error("expected error for invalid key size, got nil")
	}
}

func TestDecryptAESGCM_InvalidKeySize(t *testing.T) {
	_, err := crypto.DecryptAESGCM([]byte("cipher"), []byte("short"))
	if err == nil {
		t.Error("expected error for invalid key size, got nil")
	}
}

func TestDecryptAESGCM_CiphertextTooShort(t *testing.T) {
	key := []byte("0123456789ABCDEF") // 16 байт — AES-128
	// Передаём заведомо короткий ciphertext
	_, err := crypto.DecryptAESGCM([]byte("short"), key)
	if err == nil {
		t.Error("expected error for short ciphertext, got nil")
	}
}

func TestEncryptDecryptAESGCM_WrongKey(t *testing.T) {
	key := []byte("0123456789ABCDEF")      // правильный ключ
	wrongKey := []byte("AAAAAAAAAAAAAAAA") // 16 байт, но другой

	plaintext := []byte("Secret data")

	ciphertext, err := crypto.EncryptAESGCM(plaintext, key)
	if err != nil {
		t.Fatalf("EncryptAESGCM returned error: %v", err)
	}

	_, err = crypto.DecryptAESGCM(ciphertext, wrongKey)
	if err == nil {
		t.Error("expected error when decrypting with wrong key, got nil")
	}
}

func TestEncryptDecryptStream_Success(t *testing.T) {
	key := make([]byte, 32) // AES-256
	_, err := rand.Read(key)
	assert.NoError(t, err)

	plaintext := []byte("Hello, this is a test of EncryptStream and DecryptStream!")

	var encrypted bytes.Buffer
	err = crypto.EncryptStream(bytes.NewReader(plaintext), &encrypted, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted.Bytes())

	var decrypted bytes.Buffer
	err = crypto.DecryptStream(&encrypted, &decrypted, key)
	assert.NoError(t, err)

	assert.Equal(t, plaintext, decrypted.Bytes())
}

func TestEncryptStream_InvalidKey(t *testing.T) {
	key := []byte("short") // некорректный размер ключа
	var buf bytes.Buffer
	err := crypto.EncryptStream(bytes.NewReader([]byte("data")), &buf, key)
	assert.Error(t, err)
}

func TestDecryptStream_InvalidKey(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	plaintext := []byte("test data")
	var encrypted bytes.Buffer
	err := crypto.EncryptStream(bytes.NewReader(plaintext), &encrypted, key)
	assert.NoError(t, err)

	badKey := make([]byte, 32)
	_, _ = rand.Read(badKey)
	var decrypted bytes.Buffer
	err = crypto.DecryptStream(&encrypted, &decrypted, badKey)
	assert.Error(t, err) // ошибка из-за неверного ключа
}

func TestDecryptStream_TruncatedData(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	plaintext := []byte("test data")
	var encrypted bytes.Buffer
	err := crypto.EncryptStream(bytes.NewReader(plaintext), &encrypted, key)
	assert.NoError(t, err)

	// Укорачиваем зашифрованные данные
	truncated := encrypted.Bytes()[:len(encrypted.Bytes())-5]
	var decrypted bytes.Buffer
	err = crypto.DecryptStream(bytes.NewReader(truncated), &decrypted, key)
	assert.Error(t, err)
}

func TestEncryptDecryptStream_LargeData(t *testing.T) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	plaintext := bytes.Repeat([]byte("A"), 1024*1024) // 1 MB данных

	var encrypted bytes.Buffer
	err := crypto.EncryptStream(bytes.NewReader(plaintext), &encrypted, key)
	assert.NoError(t, err)

	var decrypted bytes.Buffer
	err = crypto.DecryptStream(&encrypted, &decrypted, key)
	assert.NoError(t, err)

	assert.Equal(t, plaintext, decrypted.Bytes())
}
