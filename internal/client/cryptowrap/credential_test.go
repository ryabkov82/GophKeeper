package cryptowrap

import (
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

func TestEncryptDecryptCredential(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 байта для AES-256

	original := &model.Credential{
		Login:    "user@example.com",
		Password: "mySecretPass",
		Metadata: "some additional info",
	}

	// Копируем исходные данные для проверки после дешифровки
	copyCred := *original

	// Шифруем
	if err := EncryptCredential(original, key); err != nil {
		t.Fatalf("EncryptCredential failed: %v", err)
	}

	// Проверяем, что данные изменились (зашифрованы и base64)
	if original.Login == copyCred.Login ||
		original.Password == copyCred.Password ||
		original.Metadata == copyCred.Metadata {
		t.Error("EncryptCredential: данные не изменились после шифрования")
	}

	// Дешифруем
	if err := DecryptCredential(original, key); err != nil {
		t.Fatalf("DecryptCredential failed: %v", err)
	}

	// Проверяем, что восстановили исходные значения
	if original.Login != copyCred.Login ||
		original.Password != copyCred.Password ||
		original.Metadata != copyCred.Metadata {
		t.Error("DecryptCredential: восстановленные данные не совпадают с исходными")
	}
}

func TestDecryptCredentialWithInvalidData(t *testing.T) {
	key := []byte("12345678901234567890123456789012")

	cred := &model.Credential{
		Login:    "invalid_base64@@@",
		Password: "also_invalid_base64@@@",
		Metadata: "12345",
	}

	err := DecryptCredential(cred, key)
	if err == nil {
		t.Error("DecryptCredential: ожидалась ошибка при дешифровке некорректных данных, но ошибки не было")
	}
}

func TestCredentialCryptoWrapper_EncryptDecrypt(t *testing.T) {
	key := []byte("12345678901234567890123456789012")

	cred := &model.Credential{
		Login:    "login",
		Password: "password",
		Metadata: "metadata",
	}

	wrapper := &CredentialCryptoWrapper{Credential: cred}

	if err := wrapper.Encrypt(key); err != nil {
		t.Fatalf("CredentialCryptoWrapper.Encrypt failed: %v", err)
	}

	// После шифрования данные должны измениться
	if wrapper.Login == "login" || wrapper.Password == "password" || wrapper.Metadata == "metadata" {
		t.Error("CredentialCryptoWrapper.Encrypt: данные не изменились после шифрования")
	}

	if err := wrapper.Decrypt(key); err != nil {
		t.Fatalf("CredentialCryptoWrapper.Decrypt failed: %v", err)
	}

	// После дешифрования данные должны вернуться к исходным
	if wrapper.Login != "login" || wrapper.Password != "password" || wrapper.Metadata != "metadata" {
		t.Error("CredentialCryptoWrapper.Decrypt: восстановленные данные не совпадают с исходными")
	}
}
