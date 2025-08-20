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

func TestInitializeWithTimestamp(t *testing.T) {
	tempDir := t.TempDir()

	if err := logger.InitializeWithTimestamp("debug", tempDir); err != nil {
		t.Fatalf("InitializeWithTimestamp failed: %v", err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			t.Errorf("error closing logger: %v", err)
		}
	}()

	logger.Log.Info("from timestamped logger")

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("failed to read log dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected log file to be created")
	}
}

func TestCloseWithoutInitialize(t *testing.T) {
	// Ensure closing without prior initialization does not error.
	if err := logger.Close(); err != nil {
		t.Fatalf("unexpected error closing logger: %v", err)
	}
}
