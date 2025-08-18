// Package filesystem содержит реализацию BinaryDataStorage,
// которая сохраняет бинарные данные в локальной файловой системе.
package filesystem

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/domain/storage"
)

// binaryDataStorage реализует интерфейс storage.BinaryDataStorage
// и сохраняет данные в локальной файловой системе.
type binaryDataStorage struct {
	basePath string
	stopCh   chan struct{}
}

// NewBinaryDataStorage создаёт локальное хранилище с фоновым удалением
// старых .tmp файлов. interval — частота очистки, maxAge — файлы старше
// maxAge удаляются.
func NewBinaryDataStorage(basePath string, interval, maxAge time.Duration) storage.BinaryDataStorage {
	fs := &binaryDataStorage{
		basePath: basePath,
		stopCh:   make(chan struct{}),
	}
	StartTempFileCleaner(fs.basePath, interval, maxAge, fs.stopCh)
	return fs
}

// Save безопасно сохраняет бинарные данные в локальном хранилище и возвращает размер файла.
// Данные сначала пишутся во временный файл (*.tmp). Если запись завершилась успешно,
// файл переименовывается в финальное имя. В случае ошибки временный файл удаляется.
func (fs *binaryDataStorage) Save(ctx context.Context, userID string, r io.Reader) (storagePath string, size int64, err error) {
	// Путь к каталогу пользователя
	userDir := filepath.Join(fs.basePath, userID)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		return "", 0, fmt.Errorf("failed to create user dir: %w", err)
	}

	// Уникальное имя файла
	fileName := uuid.New().String() + ".bin"
	tmpFileName := fileName + ".tmp"
	tmpPath := filepath.Join(userDir, tmpFileName)
	finalPath := filepath.Join(userDir, fileName)

	// Создаём временный файл
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Если что-то пойдёт не так, удаляем временный файл
	defer func() {
		tmpFile.Close()
		if _, statErr := os.Stat(tmpPath); statErr == nil {
			_ = os.Remove(tmpPath)
		}
	}()

	// Копируем данные в временный файл и получаем размер
	size, err = io.Copy(tmpFile, r)
	if err != nil {
		return "", 0, fmt.Errorf("failed to write data: %w", err)
	}

	// Закрываем файл перед переименованием
	if err := tmpFile.Close(); err != nil {
		return "", 0, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Переименовываем временный файл в финальный
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return "", 0, fmt.Errorf("failed to finalize file: %w", err)
	}

	// Возвращаем относительный путь и размер
	storagePath = filepath.Join(userID, fileName)
	return storagePath, size, nil
}

// Load открывает файл для чтения по относительному пути storagePath.
func (fs *binaryDataStorage) Load(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	filePath := filepath.Join(fs.basePath, storagePath)

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return f, nil
}

// Delete удаляет файл по относительному пути storagePath.
func (fs *binaryDataStorage) Delete(ctx context.Context, storagePath string) error {
	filePath := filepath.Join(fs.basePath, storagePath)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			// Если файл уже удалён — не считаем это ошибкой
			return nil
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Close останавливает фоновый очиститель.
func (fs *binaryDataStorage) Close() {
	close(fs.stopCh)
}

// StartTempFileCleaner запускает фоновую горутину, которая периодически
// удаляет старые временные файлы (*.tmp) из basePath.
// interval — частота проверки (например, 1 час).
// maxAge — файлы старше этой продолжительности будут удаляться.
func StartTempFileCleaner(basePath string, interval, maxAge time.Duration, stopCh <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				cleanOldTempFiles(basePath, maxAge)
			case <-stopCh:
				log.Println("Temp file cleaner stopped")
				return
			}
		}
	}()
}

// cleanOldTempFiles ищет и удаляет все файлы с расширением .tmp,
// старше maxAge.
func cleanOldTempFiles(basePath string, maxAge time.Duration) {
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Не можем прочитать файл — пропускаем
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".tmp" {
			return nil
		}

		if time.Since(info.ModTime()) > maxAge {
			if err := os.Remove(path); err != nil {
				log.Printf("Failed to remove temp file %s: %v", path, err)
			} else {
				log.Printf("Removed old temp file: %s", path)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking temp files: %v", err)
	}
}
