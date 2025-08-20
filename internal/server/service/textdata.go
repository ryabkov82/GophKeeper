package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
	"github.com/ryabkov82/gophkeeper/internal/domain/service"
)

// TextDataServiceImpl реализует интерфейс TextDataService
type TextDataServiceImpl struct {
	repo repository.TextDataRepository
}

// NewTextDataService создаёт новый сервис с указанным репозиторием
func NewTextDataService(repo repository.TextDataRepository) service.TextDataService {
	return &TextDataServiceImpl{repo: repo}
}

// Create создаёт новую запись TextData с генерацией UUID и датой создания
func (s *TextDataServiceImpl) Create(ctx context.Context, data *model.TextData) error {
	if data.ID == "" {
		data.ID = uuid.NewString()
	}
	now := time.Now()
	data.CreatedAt = now
	data.UpdatedAt = now

	// Базовая валидация обязательных полей
	if data.Title == "" {
		return errors.New("title is required")
	}
	if len(data.Content) == 0 {
		return errors.New("content is required")
	}

	return s.repo.Create(ctx, data)
}

// GetByID возвращает полную запись TextData по её уникальному идентификатору и userID
func (s *TextDataServiceImpl) GetByID(ctx context.Context, userID, id string) (*model.TextData, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	return s.repo.GetByID(ctx, userID, id)
}

// ListTitles возвращает список заголовков всех записей пользователя (ID + Title)
func (s *TextDataServiceImpl) ListTitles(ctx context.Context, userID string) ([]*model.TextData, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	return s.repo.ListTitles(ctx, userID)
}

// Update обновляет существующую запись TextData с обновлением времени
func (s *TextDataServiceImpl) Update(ctx context.Context, data *model.TextData) error {
	if data.ID == "" {
		return errors.New("id is required for update")
	}
	if data.UserID == "" {
		return errors.New("userID is required for update")
	}

	/* проверка существования внутри Update
	// Проверка существования записи перед обновлением
	existing, err := s.repo.GetByID(ctx, data.UserID, data.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("text data not found")
	}
	*/

	data.UpdatedAt = time.Now()
	return s.repo.Update(ctx, data)
}

// Delete удаляет запись TextData по идентификатору и userID
func (s *TextDataServiceImpl) Delete(ctx context.Context, userID, id string) error {
	if id == "" {
		return errors.New("id is required for delete")
	}
	if userID == "" {
		return errors.New("userID is required for delete")
	}

	// Проверка существования записи перед удалением
	existing, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("text data not found")
	}

	return s.repo.Delete(ctx, userID, id)
}
