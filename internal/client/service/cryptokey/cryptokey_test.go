package cryptokey

import (
	"errors"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/crypto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Мок для storage.CryptoKeyStorage
type mockCryptoKeyStorage struct {
	saveKey    []byte
	saveParams crypto.Argon2Params
	saveErr    error

	loadKey    []byte
	loadParams crypto.Argon2Params
	loadErr    error

	clearErr error
}

func (m *mockCryptoKeyStorage) Save(key []byte, params crypto.Argon2Params) error {
	m.saveKey = key
	m.saveParams = params
	return m.saveErr
}

func (m *mockCryptoKeyStorage) Load() ([]byte, crypto.Argon2Params, error) {
	return m.loadKey, m.loadParams, m.loadErr
}

func (m *mockCryptoKeyStorage) Clear() error {
	return m.clearErr
}

func TestGenerateAndSaveKey_Success(t *testing.T) {
	mockStore := &mockCryptoKeyStorage{}
	manager := NewCryptoKeyManager(mockStore, zap.NewNop())

	password := "password123"
	salt := []byte("somesalt")

	err := manager.GenerateAndSaveKey(password, salt)
	assert.NoError(t, err)
	assert.NotNil(t, manager.key)
	assert.NotZero(t, len(manager.key))
	assert.Equal(t, manager.key, mockStore.saveKey)
	assert.Equal(t, manager.params, mockStore.saveParams)
}

func TestGenerateAndSaveKey_SaveError(t *testing.T) {
	mockStore := &mockCryptoKeyStorage{
		saveErr: errors.New("save failed"),
	}
	manager := NewCryptoKeyManager(mockStore, zap.NewNop())

	password := "password123"
	salt := []byte("somesalt")

	err := manager.GenerateAndSaveKey(password, salt)
	assert.Error(t, err)
	assert.EqualError(t, err, "save failed")
}

func TestLoadKey_AlreadyInMemory(t *testing.T) {
	mockStore := &mockCryptoKeyStorage{
		loadKey:    []byte("storedkey"),
		loadParams: crypto.Argon2Params{Memory: 64, Time: 2, Threads: 1},
	}
	manager := NewCryptoKeyManager(mockStore, zap.NewNop())

	// Сначала загрузим ключ из хранилища, чтобы положить в память
	_, err := manager.LoadKey()
	assert.NoError(t, err)
	assert.Equal(t, mockStore.loadKey, manager.key)

	// Теперь при повторном вызове ключ должен быть взят из памяти (mockStore.Load не вызывается)
	key, err := manager.LoadKey()
	assert.NoError(t, err)
	assert.Equal(t, mockStore.loadKey, key)
}

func TestLoadKey_LoadError(t *testing.T) {
	mockStore := &mockCryptoKeyStorage{
		loadErr: errors.New("load failed"),
	}
	manager := NewCryptoKeyManager(mockStore, zap.NewNop())

	key, err := manager.LoadKey()
	assert.Nil(t, key)
	assert.Error(t, err)
	assert.EqualError(t, err, "load failed")
}

func TestClearKey_Success(t *testing.T) {
	mockStore := &mockCryptoKeyStorage{}
	manager := NewCryptoKeyManager(mockStore, zap.NewNop())

	// Имитируем что ключ в памяти
	manager.key = []byte("somekey")
	manager.params = crypto.Argon2Params{Memory: 64}

	err := manager.ClearKey()
	assert.NoError(t, err)
	assert.Nil(t, manager.key)
	assert.Equal(t, crypto.Argon2Params{}, manager.params)
}

func TestClearKey_ClearError(t *testing.T) {
	mockStore := &mockCryptoKeyStorage{
		clearErr: errors.New("clear failed"),
	}
	manager := NewCryptoKeyManager(mockStore, zap.NewNop())

	manager.key = []byte("somekey")
	manager.params = crypto.Argon2Params{Memory: 64}

	err := manager.ClearKey()
	assert.Error(t, err)
	assert.EqualError(t, err, "clear failed")
}
