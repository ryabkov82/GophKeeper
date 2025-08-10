package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
	domainService "github.com/ryabkov82/gophkeeper/internal/domain/service"
)

// CredentialService реализует интерфейс service.CredentialService
type CredentialService struct {
	repo repository.CredentialRepository
}

// NewCredentialService создаёт новый сервис с указанным репозиторием
func NewCredentialService(repo repository.CredentialRepository) domainService.CredentialService {
	return &CredentialService{repo: repo}
}

// Create создаёт новую запись учётных данных с генерацией UUID и датой создания
func (s *CredentialService) Create(ctx context.Context, cred *model.Credential) error {
	if cred.ID == "" {
		cred.ID = uuid.NewString()
	}
	now := time.Now()
	cred.CreatedAt = now
	cred.UpdatedAt = now
	return s.repo.Create(ctx, cred)
}

// GetByID возвращает учётные данные по их уникальному идентификатору
func (s *CredentialService) GetByID(ctx context.Context, id string) (*model.Credential, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return s.repo.GetByID(ctx, id)
}

// GetByUserID возвращает все учётные данные пользователя
func (s *CredentialService) GetByUserID(ctx context.Context, userID string) ([]model.Credential, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	return s.repo.GetByUserID(ctx, userID)
}

// Update обновляет существующую запись учётных данных с обновлением времени
func (s *CredentialService) Update(ctx context.Context, cred *model.Credential) error {
	if cred.ID == "" {
		return errors.New("id is required for update")
	}
	cred.UpdatedAt = time.Now()
	return s.repo.Update(ctx, cred)
}

// Delete удаляет запись учётных данных по идентификатору
func (s *CredentialService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required for delete")
	}
	return s.repo.Delete(ctx, id)
}
