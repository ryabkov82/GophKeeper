// Package cryptokey предоставляет функциональность для управления
// симметричным ключом шифрования на клиенте:
// - генерация ключа из пароля и соли с использованием Argon2id,
// - сохранение и загрузка ключа с параметрами KDF в безопасном хранилище,
// - очистка ключа из памяти и хранилища.
//
// Этот пакет служит абстракцией над механизмами хранения и генерации
// ключа, используемого для шифрования приватных данных перед
// отправкой на сервер и после получения с сервера.
package cryptokey

import (
	"github.com/ryabkov82/gophkeeper/internal/client/crypto"
	"github.com/ryabkov82/gophkeeper/internal/client/storage"
	"go.uber.org/zap"
)

// CryptoKeyManagerIface описывает поведение менеджера симметричного ключа.
type CryptoKeyManagerIface interface {
	// GenerateAndSaveKey генерирует и сохраняет ключ по паролю и соли.
	GenerateAndSaveKey(password string, salt []byte) error

	// LoadKey загружает ключ в память (если он есть в хранилище) и возвращает его.
	LoadKey() ([]byte, error)

	// ClearKey удаляет ключ из памяти и хранилища.
	ClearKey() error
}

// CryptoKeyManager управляет жизненным циклом симметричного
// ключа шифрования: генерацией, сохранением, загрузкой и очисткой.
//
// Включает хранение ключа в памяти и взаимодействие с постоянным
// хранилищем (например, файловым).
type CryptoKeyManager struct {
	key      []byte
	params   crypto.Argon2Params
	keyStore storage.CryptoKeyStorage
	logger   *zap.Logger
}

// NewCryptoKeyManager создаёт новый экземпляр CryptoKeyManager.
//
// keyStore — реализация интерфейса для постоянного хранения ключа,
// logger — логгер для отладки и информации.
func NewCryptoKeyManager(store storage.CryptoKeyStorage, logger *zap.Logger) *CryptoKeyManager {
	return &CryptoKeyManager{
		keyStore: store,
		logger:   logger,
	}
}

// GenerateAndSaveKey генерирует симметричный ключ из пароля и соли,
// используя Argon2id, и сохраняет ключ вместе с параметрами в хранилище.
//
// password — пароль пользователя,
// salt — соль, полученная с сервера.
//
// Возвращает ошибку, если генерация или сохранение ключа не удались.
func (c *CryptoKeyManager) GenerateAndSaveKey(password string, salt []byte) error {
	key, params, err := crypto.DeriveKey(password, salt)
	if err != nil {
		return err
	}
	if err := c.keyStore.Save(key, params); err != nil {
		return err
	}
	c.key = key
	c.params = params
	c.logger.Info("Crypto key generated and saved", zap.Int("key_len", len(key)))
	return nil
}

// LoadKey загружает ключ и параметры из постоянного хранилища,
// а также сохраняет ключ в памяти.
//
// Возвращает ключ и ошибку при неудаче загрузки.
func (c *CryptoKeyManager) LoadKey() ([]byte, error) {
	if len(c.key) != 0 {
		return c.key, nil
	}
	key, params, err := c.keyStore.Load()
	if err != nil {
		c.logger.Warn("Failed to load crypto key", zap.Error(err))
		return nil, err
	}
	c.key = key
	c.params = params
	return key, nil
}

// ClearKey удаляет ключ из памяти и постоянного хранилища.
//
// Используется при выходе пользователя или смене учётных данных.
//
// Возвращает ошибку, если удаление из хранилища не удалось.
func (c *CryptoKeyManager) ClearKey() error {
	c.key = nil
	c.params = crypto.Argon2Params{}
	return c.keyStore.Clear()
}
