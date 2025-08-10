package service

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// CredentialService описывает контракт сервиса для работы с учётными данными (логин/пароль).
type CredentialService interface {
	// Create создаёт новую запись учётных данных.
	Create(ctx context.Context, cred *model.Credential) error

	// GetByID возвращает учётные данные по их уникальному идентификатору.
	GetByID(ctx context.Context, id string) (*model.Credential, error)

	// GetByUserID возвращает все учётные данные пользователя.
	GetByUserID(ctx context.Context, userID string) ([]model.Credential, error)

	// Update обновляет существующую запись учётных данных.
	Update(ctx context.Context, cred *model.Credential) error

	// Delete удаляет запись учётных данных по идентификатору.
	Delete(ctx context.Context, id string) error
}
