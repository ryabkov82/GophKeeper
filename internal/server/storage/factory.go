package storage

import (
	"database/sql"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
	"github.com/ryabkov82/gophkeeper/internal/server/storage/postgres"
)

// postgresFactory — реализация StorageFactory для работы с PostgreSQL.
type postgresFactory struct {
	userRepo       repository.UserRepository
	credentialRepo repository.CredentialRepository
	bankCardRepo   repository.BankCardRepository
	textDataRepo   repository.TextDataRepository
}

// NewPostgresFactory создаёт фабрику, использующую Postgres как источник данных.
func NewPostgresFactory(db *sql.DB) repository.StorageFactory {
	return &postgresFactory{
		userRepo:       postgres.NewUserStorage(db),
		credentialRepo: postgres.NewCredentialStorage(db),
		bankCardRepo:   postgres.NewBankCardStorage(db),
		textDataRepo:   postgres.NewTextDataStorage(db),
	}
}

func (f *postgresFactory) User() repository.UserRepository {
	return f.userRepo
}

func (f *postgresFactory) Credential() repository.CredentialRepository {
	return f.credentialRepo
}

func (f *postgresFactory) BankCard() repository.BankCardRepository {
	return f.bankCardRepo
}

func (f *postgresFactory) TextData() repository.TextDataRepository {
	return f.textDataRepo
}

// NewStorageFactory создает фабрику репозиториев для указанного драйвера БД.
//
// Параметры:
//   - driver: имя драйвера ("postgres", "pgx", "inmemory" и т.п.)
//   - db: подключение к базе данных (*sql.DB) — обязательно для SQL-реализаций.
//
// Возвращает:
//   - экземпляр StorageFactory с нужными реализациями репозиториев
//   - ошибку, если передан неизвестный драйвер
func NewStorageFactory(driver string, db *sql.DB) (repository.StorageFactory, error) {
	switch driver {
	case "pgx", "postgres":
		return NewPostgresFactory(db), nil
	default:
		return nil, fmt.Errorf("неизвестный тип хранилища: %s", driver)
	}
}
