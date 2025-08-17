package repository

// StorageFactory определяет фабрику, которая создаёт реализации всех репозиториев.
//
// Такой подход позволяет легко подменять хранилище данных
// (Postgres, InMemory, Mock и т.п.) без изменения бизнес-логики.
type StorageFactory interface {
	User() UserRepository
	Credential() CredentialRepository
	BankCard() BankCardRepository
	TextData() TextDataRepository
	// Если будут новые сущности — добавляем сюда
}
