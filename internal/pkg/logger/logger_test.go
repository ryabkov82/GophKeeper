package logger_test

import (
	"os"
	"path/filepath"
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

func TestInitialize_WithLogFile(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	err := logger.Initialize("debug", logFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	defer func() {
		if err := logger.Close(); err != nil {
			t.Errorf("error closing logger: %v", err)
		}
	}()

	if logger.Log == nil {
		t.Fatal("expected logger.Log to be initialized")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("logging caused panic: %v", r)
		}
	}()

	logger.Log.Info("test info log")

	info, err := os.Stat(logFile)
	if err != nil {
		t.Fatalf("expected log file to be created, but got error: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected log file to be non-empty after logging")
	}
}
