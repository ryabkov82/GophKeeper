package filesystem_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/server/storage/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestBinaryDataStorage_SaveLoadDelete(t *testing.T) {
	ctx := context.Background()
	baseDir := t.TempDir()

	// Интервал очистки 100ms, maxAge 500ms для теста
	fs := filesystem.NewBinaryDataStorage(baseDir, 100*time.Millisecond, 500*time.Millisecond)
	defer fs.Close()

	userID := "user123"
	data := []byte("hello world")

	// --- Save ---
	storagePath, err := fs.Save(ctx, userID, bytes.NewReader(data))
	assert.NoError(t, err)
	assert.NotEmpty(t, storagePath)

	fullPath := filepath.Join(baseDir, storagePath)
	_, err = os.Stat(fullPath)
	assert.NoError(t, err, "файл должен существовать после Save")

	// --- Load ---
	rc, err := fs.Load(ctx, storagePath)
	assert.NoError(t, err)

	loadedData, err := io.ReadAll(rc)
	assert.NoError(t, err)
	assert.Equal(t, data, loadedData)

	rc.Close() // закрываем поток для Windows

	// --- Delete ---
	err = fs.Delete(ctx, storagePath)
	assert.NoError(t, err)

	_, err = os.Stat(fullPath)
	assert.True(t, os.IsNotExist(err), "файл должен быть удалён")
}

func TestBinaryDataStorage_DeleteNonExistentFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	fs := filesystem.NewBinaryDataStorage(tmpDir, 100*time.Millisecond, 500*time.Millisecond)
	defer fs.Close()

	// Удаляем несуществующий файл
	err := fs.Delete(ctx, "nonexistent/file.bin")
	assert.NoError(t, err)
}

func TestBinaryDataStorage_LoadNonExistentFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	fs := filesystem.NewBinaryDataStorage(tmpDir, 100*time.Millisecond, 500*time.Millisecond)
	defer fs.Close()

	// Пытаемся открыть несуществующий файл
	rc, err := fs.Load(ctx, "nonexistent/file.bin")
	assert.Error(t, err)
	assert.Nil(t, rc)
}

func TestStartTempFileCleaner(t *testing.T) {
	tmpDir := t.TempDir()

	// Создаём старый и новый временные файлы
	oldFile := filepath.Join(tmpDir, "old.tmp")
	newFile := filepath.Join(tmpDir, "new.tmp")

	err := os.WriteFile(oldFile, []byte("old"), 0o644)
	assert.NoError(t, err)
	err = os.WriteFile(newFile, []byte("new"), 0o644)
	assert.NoError(t, err)

	// Принудительно меняем время модификации старого файла на прошлое
	past := time.Now().Add(-2 * time.Second)
	err = os.Chtimes(oldFile, past, past)
	assert.NoError(t, err)

	stopCh := make(chan struct{})
	// Запускаем очистку каждые 100ms, удаляем файлы старше 1 секунды
	filesystem.StartTempFileCleaner(tmpDir, 100*time.Millisecond, 1*time.Second, stopCh)
	defer close(stopCh)

	// Ждём немного, чтобы горутина успела сработать
	time.Sleep(300 * time.Millisecond)

	// Проверяем, что старый файл удалён
	_, err = os.Stat(oldFile)
	assert.True(t, os.IsNotExist(err), "старый .tmp файл должен быть удалён")

	// Проверяем, что новый файл остался
	_, err = os.Stat(newFile)
	assert.NoError(t, err, "новый .tmp файл не должен быть удалён")
}
