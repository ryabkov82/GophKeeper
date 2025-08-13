package paths_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/paths"
)

func TestDefaultLogDir(t *testing.T) {
	// Для Windows тестируем отдельно отсутствие APPDATA
	if runtime.GOOS == "windows" {
		origAppData := os.Getenv("APPDATA")
		defer os.Setenv("APPDATA", origAppData)

		// Тест когда APPDATA не установлен — должно вернуть ошибку
		os.Unsetenv("APPDATA")
		_, err := paths.DefaultLogDir()
		if err == nil {
			t.Error("Expected error when APPDATA is not set on Windows, got nil")
		}

		// Восстановим APPDATA
		os.Setenv("APPDATA", origAppData)
	}

	// Общий тест на успешное получение каталога
	dir, err := paths.DefaultLogDir()
	if err != nil {
		t.Fatalf("DefaultLogDir() returned error: %v", err)
	}

	// Проверяем, что каталог существует и является директорией
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("os.Stat(%q) returned error: %v", dir, err)
	}
	if !info.IsDir() {
		t.Errorf("Expected %q to be a directory", dir)
	}

	// Проверим, что путь содержит "logs" (проверка простая, чтобы убедиться в корректности)
	if filepath.Base(dir) != "logs" {
		t.Errorf("Expected directory basename to be 'logs', got %q", filepath.Base(dir))
	}
}

// Тест для DefaultTokenFilePath
func TestDefaultTokenFilePath(t *testing.T) {
	path, err := paths.DefaultTokenFilePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filepath.Base(path) != ".token" {
		t.Errorf("expected token filename '.token', got %q", filepath.Base(path))
	}

	// Получаем домашнюю директорию (кроссплатформенно)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home dir: %v", err)
	}

	// Путь к gophkeeper в конфиге для Linux/macOS
	configPathUnix := filepath.Join(homeDir, ".config", "gophkeeper")

	// Путь к gophkeeper в APPDATA для Windows
	appData := os.Getenv("APPDATA")
	configPathWin := ""
	if appData != "" {
		configPathWin = filepath.Join(appData, "gophkeeper")
	}

	// Проверяем, что path внутри одной из этих директорий
	isInsideUnix := isSubpath(configPathUnix, path)
	isInsideWin := false
	if configPathWin != "" {
		isInsideWin = isSubpath(configPathWin, path)
	}

	if !isInsideUnix && !isInsideWin {
		t.Errorf("expected path to be inside gophkeeper config dir, got %q", path)
	}
}

// isSubpath проверяет, находится ли subPath внутри basePath
func isSubpath(basePath, subPath string) bool {
	rel, err := filepath.Rel(basePath, subPath)
	if err != nil {
		return false
	}
	// rel не должен начинаться с ".."
	return !strings.HasPrefix(rel, "..") && rel != ".."
}

// Тест для DefaultKeyFilePath
func TestDefaultKeyFilePath(t *testing.T) {
	path, err := paths.DefaultKeyFilePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if filepath.Base(path) != "key.json" {
		t.Errorf("expected key filename 'key.json', got %q", filepath.Base(path))
	}

	// Аналогично проверяем, что путь содержит gophkeeper
	if !containsIgnoreCase(path, "gophkeeper") {
		t.Errorf("expected path to contain 'gophkeeper', got %q", path)
	}
}

// Вспомогательная функция для проверки подстроки без учёта регистра
func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}
