package storage

// BinaryDataStorageFactory предоставляет реализацию BinaryDataStorage.
// Это позволяет менять бэкенд без изменения сервисного кода.
type BinaryDataStorageFactory interface {
	BinaryData() BinaryDataStorage
}
