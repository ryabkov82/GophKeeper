package storage

import (
	"database/sql"

	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
	"github.com/ryabkov82/gophkeeper/internal/server/storage/postgres"
)

func NewRepositories(db *sql.DB) *repository.Repositories {
	return &repository.Repositories{
		User:       postgres.NewUserStorage(db),
		Credential: postgres.NewCredentialStorage(db),
	}
}
