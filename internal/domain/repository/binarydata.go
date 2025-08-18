package repository

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// BinaryDataRepository определяет контракт для работы с бинарными данными.
// Интерфейс реализуется на стороне хранилища (Postgres, S3, локальная ФС и т.д.).
type BinaryDataRepository interface {
	// Save сохраняет новую запись бинарных данных.
	Save(ctx context.Context, data *model.BinaryData) error

	// GetByID возвращает бинарные данные по их идентификатору и владельцу.
	GetByID(ctx context.Context, userID, id string) (*model.BinaryData, error)

	// ListByUser возвращает все бинарные данные конкретного пользователя.
	ListByUser(ctx context.Context, userID string) ([]*model.BinaryData, error)

	// Delete удаляет запись по идентификатору и владельцу.
	Delete(ctx context.Context, userID, id string) error
}
