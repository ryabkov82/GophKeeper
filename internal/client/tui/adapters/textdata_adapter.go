package adapters

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// TextDataAdapter — адаптер, превращающий TextDataService в общий интерфейс DataService
type TextDataAdapter struct {
	svc contracts.TextDataService
}

// NewTextDataAdapter создаёт адаптер для текстовых данных
func NewTextDataAdapter(svc contracts.TextDataService) *TextDataAdapter {
	return &TextDataAdapter{svc: svc}
}

// List — возвращает список текстовых данных (только заголовки) в виде []ListItem
func (a *TextDataAdapter) List(ctx context.Context) ([]contracts.ListItem, error) {
	texts, err := a.svc.GetTextDataTitles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list text data: %w", err)
	}

	items := make([]contracts.ListItem, 0, len(texts))
	for _, t := range texts {
		items = append(items, contracts.ListItem{
			ID:    t.ID,
			Title: t.Title,
		})
	}
	return items, nil
}

// Get — возвращает TextData по ID
func (a *TextDataAdapter) Get(ctx context.Context, id string) (interface{}, error) {
	text, err := a.svc.GetTextDataByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get text data: %w", err)
	}
	return text, nil
}

// Create — создаёт новый TextData
func (a *TextDataAdapter) Create(ctx context.Context, v interface{}) error {
	text, ok := v.(*model.TextData)
	if !ok {
		return fmt.Errorf("invalid type for Create: expected *model.TextData, got %T", v)
	}
	return a.svc.CreateTextData(ctx, text)
}

// Update — обновляет TextData
func (a *TextDataAdapter) Update(ctx context.Context, id string, v interface{}) error {
	text, ok := v.(*model.TextData)
	if !ok {
		return fmt.Errorf("invalid type for Update: expected *model.TextData, got %T", v)
	}
	if text.ID != id {
		text.ID = id
	}
	return a.svc.UpdateTextData(ctx, text)
}

// Delete — удаляет TextData по ID
func (a *TextDataAdapter) Delete(ctx context.Context, id string) error {
	return a.svc.DeleteTextData(ctx, id)
}
