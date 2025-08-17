package service

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// TextDataService определяет контракт для работы с произвольными текстовыми данными.
// Сервис предоставляет методы для создания, чтения, обновления и удаления (CRUD)
// записей с обязательной проверкой прав доступа.
type TextDataService interface {
	// Create создает новую запись TextData в системе.
	Create(ctx context.Context, data *model.TextData) error

	// GetByID возвращает полную запись TextData по её уникальному идентификатору.
	GetByID(ctx context.Context, userID, id string) (*model.TextData, error)

	// ListTitles возвращает список заголовков всех записей пользователя (ID + Title).
	ListTitles(ctx context.Context, userID string) ([]*model.TextData, error)

	// Update обновляет существующую запись TextData.
	Update(ctx context.Context, data *model.TextData) error

	// Delete удаляет запись TextData по идентификатору.
	Delete(ctx context.Context, userID, id string) error
}
