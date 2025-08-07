package auth

import (
	"os"
	"path/filepath"
)

// FileTokenStorage реализует интерфейс TokenStorage, используя файл
// на локальной файловой системе для хранения токена авторизации.
type FileTokenStorage struct {
	path string
}

// NewFileTokenStorage создаёт новый экземпляр FileTokenStorage.
//
// path — путь к файлу, где будет храниться токен.
func NewFileTokenStorage(path string) *FileTokenStorage {
	return &FileTokenStorage{path: path}
}

// Save сохраняет токен в файл с правами доступа 0600 (только для владельца).
// Если родительская директория отсутствует, она будет создана с правами 0700.
func (f *FileTokenStorage) Save(token string) error {
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(f.path, []byte(token), 0600)
}

// Load читает токен из файла.
// Возвращает строку токена или ошибку, если файл отсутствует или недоступен.
func (f *FileTokenStorage) Load() (string, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Clear удаляет файл, в котором хранится токен.
// Возвращает ошибку, если файл не удалось удалить.
func (f *FileTokenStorage) Clear() error {
	return os.Remove(f.path)
}
