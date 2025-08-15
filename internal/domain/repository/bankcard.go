package repository

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// BankCardRepository — интерфейс для работы с таблицей bank_cards.
type BankCardRepository interface {
	Create(ctx context.Context, card *model.BankCard) error
	GetByID(ctx context.Context, id string) (*model.BankCard, error)
	GetByUser(ctx context.Context, userID string) ([]model.BankCard, error)
	Update(ctx context.Context, card *model.BankCard) error
	Delete(ctx context.Context, id string) error
}
