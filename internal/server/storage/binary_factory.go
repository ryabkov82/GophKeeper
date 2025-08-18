package storage

import (
	"time"

	"github.com/ryabkov82/gophkeeper/internal/domain/storage"
	"github.com/ryabkov82/gophkeeper/internal/server/config"
	"github.com/ryabkov82/gophkeeper/internal/server/storage/filesystem"
)

// binaryDataFactory — фабрика для работы с BinaryDataStorage.
type binaryDataFactory struct {
	storage storage.BinaryDataStorage
}

// NewBinaryDataFactory создаёт фабрику бинарных данных.
// В зависимости от конфигурации можно вернуть локальное или облачное хранилище.
func NewBinaryDataFactory(cfg *config.Config) storage.BinaryDataStorageFactory {
	// пример: всегда возвращаем локальное хранилище
	return &binaryDataFactory{
		storage: filesystem.NewBinaryDataStorage(cfg.BinaryDataStorePath, 1*time.Hour, 24*time.Hour),
	}
}

// BinaryData возвращает конкретную реализацию BinaryDataStorage.
func (f *binaryDataFactory) BinaryData() storage.BinaryDataStorage {
	return f.storage
}
