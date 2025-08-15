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

// BankCardService реализует интерфейс service.BankCardService
type BankCardService struct {
	repo repository.BankCardRepository
}

// NewBankCardService создаёт новый сервис с указанным репозиторием
func NewBankCardService(repo repository.BankCardRepository) domainService.BankCardService {
	return &BankCardService{repo: repo}
}

// Create создаёт новую запись банковской карты с генерацией UUID и датой создания
func (s *BankCardService) Create(ctx context.Context, card *model.BankCard) error {
	if card.ID == "" {
		card.ID = uuid.NewString()
	}
	now := time.Now()
	card.CreatedAt = now
	card.UpdatedAt = now

	// Базовая валидация обязательных полей
	if card.CardNumber == "" {
		return errors.New("card number is required")
	}
	if card.ExpiryDate == "" {
		return errors.New("expiry date is required")
	}
	if card.CardholderName == "" {
		return errors.New("cardholder name is required")
	}

	return s.repo.Create(ctx, card)
}

// GetByID возвращает банковскую карту по её уникальному идентификатору
func (s *BankCardService) GetByID(ctx context.Context, id string) (*model.BankCard, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	return s.repo.GetByID(ctx, id)
}

// GetByUserID возвращает все банковские карты пользователя
func (s *BankCardService) GetByUserID(ctx context.Context, userID string) ([]model.BankCard, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	return s.repo.GetByUser(ctx, userID)
}

// Update обновляет существующую запись банковской карты с обновлением времени
func (s *BankCardService) Update(ctx context.Context, card *model.BankCard) error {
	if card.ID == "" {
		return errors.New("id is required for update")
	}

	// Проверка существования карты перед обновлением
	existing, err := s.repo.GetByID(ctx, card.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("bank card not found")
	}

	card.UpdatedAt = time.Now()
	return s.repo.Update(ctx, card)
}

// Delete удаляет запись банковской карты по идентификатору
func (s *BankCardService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required for delete")
	}

	// Проверка существования карты перед удалением
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("bank card not found")
	}

	return s.repo.Delete(ctx, id)
}
