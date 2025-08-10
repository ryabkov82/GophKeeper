package auth

import (
	"os"
	"path/filepath"
)

// FileTokenStorage реализует интерфейс TokenStorage,
// используя файл на локальной файловой системе.
type FileTokenStorage struct {
	path       string
	writeFile  func(string, []byte, os.FileMode) error
	readFile   func(string) ([]byte, error)
	removeFile func(string) error
	mkdirAll   func(string, os.FileMode) error
}

// NewFileTokenStorage создаёт новый экземпляр FileTokenStorage.
// По умолчанию использует стандартные функции из пакета os.
func NewFileTokenStorage(path string) *FileTokenStorage {
	return &FileTokenStorage{
		path:       path,
		writeFile:  os.WriteFile,
		readFile:   os.ReadFile,
		removeFile: os.Remove,
		mkdirAll:   os.MkdirAll,
	}
}

// Save сохраняет токен в файл с правами доступа 0600 (только для владельца).
// Если родительская директория отсутствует, она будет создана с правами 0700.
func (f *FileTokenStorage) Save(token string) error {
	dir := filepath.Dir(f.path)
	if err := f.mkdirAll(dir, 0700); err != nil {
		return err
	}
	return f.writeFile(f.path, []byte(token), 0600)
}

// Load читает токен из файла.
// Возвращает строку токена или ошибку, если файл отсутствует или недоступен.
func (f *FileTokenStorage) Load() (string, error) {
	data, err := f.readFile(f.path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Clear удаляет файл, в котором хранится токен.
// Возвращает ошибку, если файл не удалось удалить.
func (f *FileTokenStorage) Clear() error {
	return f.removeFile(f.path)
}
