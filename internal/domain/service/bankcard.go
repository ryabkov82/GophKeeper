package service

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// BankCardService определяет контракт для работы с банковскими картами в системе.
// Сервис предоставляет методы для создания, чтения, обновления и удаления (CRUD)
// данных банковских карт с обязательной проверкой прав доступа.
type BankCardService interface {
	// Create создает новую запись банковской карты в системе.
	Create(ctx context.Context, card *model.BankCard) error

	// GetByID возвращает банковскую карту по её уникальному идентификатору.
	GetByID(ctx context.Context, id string) (*model.BankCard, error)

	// GetByUserID возвращает все банковские карты, принадлежащие указанному пользователю.
	GetByUserID(ctx context.Context, userID string) ([]model.BankCard, error)

	// Update обновляет существующую запись банковской карты.
	Update(ctx context.Context, card *model.BankCard) error

	// Delete удаляет запись банковской карты по идентификатору.
	Delete(ctx context.Context, id string) error
}
