package crypto

import (
	"os"
	"path/filepath"
)

// DefaultKeyFilePath возвращает стандартный путь для хранения ключа на диске.
//
// Обычно это "~/.config/gophkeeper/key.json" на Linux и macOS.
// На Windows возвращает аналогичный путь из AppData.
func DefaultKeyFilePath() (string, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg, "gophkeeper", "key.json"), nil
}
