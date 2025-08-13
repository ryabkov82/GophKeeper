package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileTokenStorage_SaveLoadClear(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.txt")

	store := NewFileTokenStorage(tokenFile)

	// Проверка сохранения токена
	err := store.Save("my_secret_token")
	require.NoError(t, err)

	// Проверка загрузки токена
	token, err := store.Load()
	require.NoError(t, err)
	require.Equal(t, "my_secret_token", token)

	// Проверка удаления токена
	err = store.Clear()
	require.NoError(t, err)

	// После удаления загрузка должна возвращать ошибку
	_, err = store.Load()
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestFileTokenStorage_Save_CreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")
	tokenFile := filepath.Join(nestedDir, "token.txt")

	store := NewFileTokenStorage(tokenFile)

	err := store.Save("token")
	require.NoError(t, err)

	// Проверяем, что файл действительно существует
	info, err := os.Stat(tokenFile)
	require.NoError(t, err)
	require.False(t, info.IsDir())
}

func TestFileTokenStorage_Save_Error(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.txt")

	store := NewFileTokenStorage(tokenFile)

	// Эмулируем ошибку MkdirAll
	store.mkdirAll = func(path string, perm os.FileMode) error {
		return os.ErrPermission
	}

	err := store.Save("token")
	require.ErrorIs(t, err, os.ErrPermission)
}

func TestFileTokenStorage_Load_Error(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.txt")

	store := NewFileTokenStorage(tokenFile)

	// Эмулируем ошибку ReadFile
	store.readFile = func(path string) ([]byte, error) {
		return nil, os.ErrNotExist
	}

	_, err := store.Load()
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestFileTokenStorage_Clear_Error(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.txt")

	store := NewFileTokenStorage(tokenFile)

	// Эмулируем ошибку Remove
	store.removeFile = func(path string) error {
		return os.ErrPermission
	}

	err := store.Clear()
	require.ErrorIs(t, err, os.ErrPermission)
}
