package logger_test

import (
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/pkg/logger"
)

func TestInitialize_ValidLevel(t *testing.T) {
	err := logger.Initialize("debug", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Проверяем, что глобальный логер теперь не является no-op
	if logger.Log == nil {
		t.Fatal("expected logger.Log to be initialized")
	}

	// Попытка залогировать что-то — не вызовет панику
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("logging caused panic: %v", r)
		}
	}()

	logger.Log.Info("test info log")
}

func TestInitialize_InvalidLevel(t *testing.T) {
	err := logger.Initialize("invalid_level", "")
	if err == nil {
		t.Fatal("expected error for invalid log level, got nil")
	}
}
