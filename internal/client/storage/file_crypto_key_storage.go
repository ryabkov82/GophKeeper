package storage

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"

	clientcrypto "github.com/ryabkov82/gophkeeper/internal/client/crypto"
)

// keyFileStruct — структура для сериализации ключа + параметров.
type keyFileStruct struct {
	KDF    string                    `json:"kdf"`
	Params clientcrypto.Argon2Params `json:"params"`
	KeyB64 string                    `json:"key_b64"`
}

// FileCryptoKeyStorage — файловая реализация CryptoKeyStorage с поддержкой KDF параметров.
type FileCryptoKeyStorage struct {
	path       string
	writeFile  func(string, []byte, os.FileMode) error
	readFile   func(string) ([]byte, error)
	removeFile func(string) error
	mkdirAll   func(string, os.FileMode) error
}

// NewFileCryptoKeyStorage создаёт файловое хранилище ключа шифрования
// по указанному пути.
func NewFileCryptoKeyStorage(path string) *FileCryptoKeyStorage {
	return &FileCryptoKeyStorage{
		path:       path,
		writeFile:  os.WriteFile,
		readFile:   os.ReadFile,
		removeFile: os.Remove,
		mkdirAll:   os.MkdirAll,
	}
}

// Save сохраняет ключ и параметры KDF в JSON-файл.
func (f *FileCryptoKeyStorage) Save(key []byte, params clientcrypto.Argon2Params) error {
	dir := filepath.Dir(f.path)
	if err := f.mkdirAll(dir, 0700); err != nil {
		return err
	}

	kfs := keyFileStruct{
		KDF:    "argon2id",
		Params: params,
		KeyB64: base64.StdEncoding.EncodeToString(key),
	}

	data, err := json.MarshalIndent(kfs, "", "  ")
	if err != nil {
		return err
	}

	tmp := f.path + ".tmp"
	if err := f.writeFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, f.path)
}

// Load загружает ключ и параметры KDF из JSON-файла.
func (f *FileCryptoKeyStorage) Load() ([]byte, clientcrypto.Argon2Params, error) {
	data, err := f.readFile(f.path)
	if err != nil {
		return nil, clientcrypto.Argon2Params{}, err
	}
	var kfs keyFileStruct
	if err := json.Unmarshal(data, &kfs); err != nil {
		return nil, clientcrypto.Argon2Params{}, err
	}
	key, err := base64.StdEncoding.DecodeString(kfs.KeyB64)
	if err != nil {
		return nil, clientcrypto.Argon2Params{}, err
	}
	return key, kfs.Params, nil
}

// Clear удаляет файл с ключом шифрования.
func (f *FileCryptoKeyStorage) Clear() error {
	return f.removeFile(f.path)
}
