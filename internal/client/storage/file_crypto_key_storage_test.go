package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/crypto"
	"github.com/stretchr/testify/require"
)

func TestFileCryptoKeyStorage_SaveLoadClear(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "keyfile.json")

	params := crypto.Argon2Params{
		Memory:  64 * 1024,
		Time:    1,
		Threads: 4,
		KeyLen:  32,
	}
	key := []byte("test-key-bytes")

	store := NewFileCryptoKeyStorage(filePath)

	// Save
	err := store.Save(key, params)
	require.NoError(t, err)

	// Проверяем, что файл действительно создался
	info, err := os.Stat(filePath)
	require.NoError(t, err)
	require.False(t, info.IsDir())

	// Load
	loadedKey, loadedParams, err := store.Load()
	require.NoError(t, err)
	require.Equal(t, key, loadedKey)
	require.Equal(t, params, loadedParams)

	// Clear
	err = store.Clear()
	require.NoError(t, err)

	// После очистки файл должен отсутствовать
	_, err = os.Stat(filePath)
	require.True(t, os.IsNotExist(err))
}

func TestFileCryptoKeyStorage_Save_MkdirAllError(t *testing.T) {
	store := NewFileCryptoKeyStorage("/some/path/key.json")
	store.mkdirAll = func(path string, mode os.FileMode) error {
		return os.ErrPermission
	}
	err := store.Save([]byte("key"), crypto.Argon2Params{})
	require.ErrorIs(t, err, os.ErrPermission)
}

func TestFileCryptoKeyStorage_Save_WriteFileError(t *testing.T) {
	store := NewFileCryptoKeyStorage("/some/path/key.json")
	store.writeFile = func(filename string, data []byte, perm os.FileMode) error {
		return os.ErrPermission
	}
	err := store.Save([]byte("key"), crypto.Argon2Params{})
	require.ErrorIs(t, err, os.ErrPermission)
}

func TestFileCryptoKeyStorage_Load_ReadFileError(t *testing.T) {
	store := NewFileCryptoKeyStorage("/non/existent/file.json")
	_, _, err := store.Load()
	require.Error(t, err)
}

func TestFileCryptoKeyStorage_Load_UnmarshalError(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "keyfile.json")
	// Запишем в файл некорректный JSON
	err := os.WriteFile(filePath, []byte("invalid json"), 0600)
	require.NoError(t, err)

	store := NewFileCryptoKeyStorage(filePath)
	_, _, err = store.Load()
	require.Error(t, err)
}

func TestFileCryptoKeyStorage_Load_Base64DecodeError(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "keyfile.json")

	// Создадим JSON с неправильным base64 в поле KeyB64
	content := `{
		"kdf": "argon2id",
		"params": {
			"Memory": 65536,
			"Time": 1,
			"Threads": 4,
			"KeyLen": 32
		},
		"key_b64": "!!!not_base64!!!"
	}`
	err := os.WriteFile(filePath, []byte(content), 0600)
	require.NoError(t, err)

	store := NewFileCryptoKeyStorage(filePath)
	_, _, err = store.Load()
	require.Error(t, err)
}

func TestFileCryptoKeyStorage_Clear_RemoveFileError(t *testing.T) {
	store := NewFileCryptoKeyStorage("/some/path/key.json")
	store.removeFile = func(path string) error {
		return os.ErrPermission
	}
	err := store.Clear()
	require.ErrorIs(t, err, os.ErrPermission)
}
