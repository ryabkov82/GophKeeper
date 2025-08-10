package repository

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// CredentialRepository определяет интерфейс для работы с сущностями Credential
// на уровне хранилища данных (например, базы данных).
//
// Интерфейс инкапсулирует CRUD-операции для учётных данных пользователя:
//   - Create — создание новой пары логин/пароль;
//   - GetByID — получение конкретной записи по её идентификатору;
//   - GetByUserID — получение всех записей, принадлежащих определённому пользователю;
//   - Update — обновление существующей записи;
//   - Delete — удаление записи по идентификатору.
//
// Конкретная реализация может использовать PostgreSQL, SQLite, in-memory-структуры
// или иные механизмы хранения данных.
type CredentialRepository interface {
	// Create сохраняет новую запись учётных данных в хранилище.
	Create(ctx context.Context, cred *model.Credential) error

	// GetByID возвращает запись учётных данных по её уникальному идентификатору.
	GetByID(ctx context.Context, id string) (*model.Credential, error)

	// GetByUserID возвращает все записи учётных данных,
	// принадлежащие указанному пользователю.
	GetByUserID(ctx context.Context, userID string) ([]model.Credential, error)

	// Update изменяет существующую запись учётных данных.
	Update(ctx context.Context, cred *model.Credential) error

	// Delete удаляет запись учётных данных по её уникальному идентификатору.
	Delete(ctx context.Context, id string) error
}
