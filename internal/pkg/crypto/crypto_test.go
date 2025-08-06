package crypto_test

import (
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/pkg/crypto"
)

func TestHashPassword(t *testing.T) {
	// Проверка ошибки при пустом пароле
	_, _, err := crypto.HashPassword("")
	if err == nil {
		t.Error("expected error for empty password, got nil")
	}

	password := "strongpassword123"

	// Проверка хеширования пароля
	hash, salt, err := crypto.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == "" || salt == "" {
		t.Fatal("hash or salt is empty")
	}

	// Проверка, что VerifyPassword вернёт true для правильного пароля
	if !crypto.VerifyPassword(password, hash, salt) {
		t.Error("VerifyPassword returned false for correct password")
	}

	// Проверка, что VerifyPassword вернёт false для неправильного пароля
	if crypto.VerifyPassword("wrongpassword", hash, salt) {
		t.Error("VerifyPassword returned true for incorrect password")
	}
}
