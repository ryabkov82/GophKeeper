package repository

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// TextDataRepository — интерфейс для работы с таблицей text_data.
type TextDataRepository interface {
	Create(ctx context.Context, data *model.TextData) error
	GetByID(ctx context.Context, userID, id string) (*model.TextData, error)
	Update(ctx context.Context, data *model.TextData) error
	Delete(ctx context.Context, userID, id string) error
	ListTitles(ctx context.Context, userID string) ([]*model.TextData, error) // возвращает только ID и Title
}
