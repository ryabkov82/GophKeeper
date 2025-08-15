package adapters

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// BankCardAdapter — адаптер, превращающий BankCardService в DataService
type BankCardAdapter struct {
	svc contracts.BankCardService
}

// NewBankCardAdapter создаёт адаптер для банковских карт
func NewBankCardAdapter(svc contracts.BankCardService) *BankCardAdapter {
	return &BankCardAdapter{svc: svc}
}

// List — возвращает список банковских карт в виде []ListItem
func (a *BankCardAdapter) List(ctx context.Context) ([]contracts.ListItem, error) {
	cards, err := a.svc.GetBankCards(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list bank cards: %w", err)
	}

	items := make([]contracts.ListItem, 0, len(cards))
	for _, c := range cards {
		items = append(items, contracts.ListItem{
			ID:    c.ID,
			Title: c.Title,
		})
	}
	return items, nil
}

// Get — возвращает BankCard по ID
func (a *BankCardAdapter) Get(ctx context.Context, id string) (interface{}, error) {
	card, err := a.svc.GetBankCardByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank card: %w", err)
	}
	return card, nil
}

// Create — создаёт новую BankCard
func (a *BankCardAdapter) Create(ctx context.Context, v interface{}) error {
	card, ok := v.(*model.BankCard)
	if !ok {
		return fmt.Errorf("invalid type for Create: expected *model.BankCard, got %T", v)
	}
	return a.svc.CreateBankCard(ctx, card)
}

// Update — обновляет BankCard
func (a *BankCardAdapter) Update(ctx context.Context, id string, v interface{}) error {
	card, ok := v.(*model.BankCard)
	if !ok {
		return fmt.Errorf("invalid type for Update: expected *model.BankCard, got %T", v)
	}
	if card.ID != id {
		card.ID = id
	}
	return a.svc.UpdateBankCard(ctx, card)
}

// Delete — удаляет BankCard
func (a *BankCardAdapter) Delete(ctx context.Context, id string) error {
	return a.svc.DeleteBankCard(ctx, id)
}
