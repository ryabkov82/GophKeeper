// Package paths предоставляет функции для определения стандартных
// кроссплатформенных путей для хранения данных клиента GophKeeper,
// таких как файлы логов, токенов и ключей шифрования.
//
// Функции учитывают особенности операционных систем Windows, macOS и Linux,
// создают необходимые каталоги и возвращают полные пути для удобного использования.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// DefaultLogDir возвращает кроссплатформенный путь к каталогу для хранения логов.
//
// Windows: %APPDATA%\GophKeeper\logs
// macOS: ~/Library/Logs/GophKeeper
// Linux/Unix: ~/.local/share/gophkeeper/logs
//
// При необходимости создает каталог с правами 0755.
//
// Возвращает полный путь к каталогу и ошибку, если возникли проблемы с определением пути или созданием каталога.
func DefaultLogDir() (string, error) {
	var logDir string
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable is not set")
		}
		logDir = filepath.Join(appData, "GophKeeper", "logs")
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		logDir = filepath.Join(homeDir, "Library", "Logs", "GophKeeper")
	default: // linux и прочие unix-подобные
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		logDir = filepath.Join(homeDir, ".local", "share", "gophkeeper", "logs")
	}

	// Создаем каталог, если не существует
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return "", err
	}

	return logDir, nil
}

// DefaultTokenFilePath возвращает стандартный путь для хранения файла с токеном.
//
// Использует системную директорию конфигурации пользователя (например, ~/.config/gophkeeper/.token).
//
// Возвращает полный путь к файлу токена и ошибку в случае неудачи.
func DefaultTokenFilePath() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg, "gophkeeper", ".token"), nil
}

// DefaultKeyFilePath возвращает стандартный путь для хранения файла с ключом шифрования.
//
// Обычно на Linux и macOS это ~/.config/gophkeeper/key.json,
// на Windows — соответствующий путь в AppData.
//
// Возвращает полный путь к файлу ключа и ошибку при неудаче.
func DefaultKeyFilePath() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg, "gophkeeper", "key.json"), nil
}
