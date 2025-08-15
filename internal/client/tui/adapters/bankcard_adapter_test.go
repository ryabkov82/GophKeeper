package adapters_test

import (
	"context"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/tui/adapters"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBankCardService мок для BankCardService
type MockBankCardService struct {
	mock.Mock
}

func (m *MockBankCardService) CreateBankCard(ctx context.Context, card *model.BankCard) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}

func (m *MockBankCardService) GetBankCardByID(ctx context.Context, id string) (*model.BankCard, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.BankCard), args.Error(1)
}

func (m *MockBankCardService) GetBankCards(ctx context.Context) ([]model.BankCard, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.BankCard), args.Error(1)
}

func (m *MockBankCardService) UpdateBankCard(ctx context.Context, card *model.BankCard) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}

func (m *MockBankCardService) DeleteBankCard(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestBankCardAdapter_List(t *testing.T) {
	mockSvc := new(MockBankCardService)
	adapter := adapters.NewBankCardAdapter(mockSvc)

	cards := []model.BankCard{
		{ID: "1", Title: "Card 1"},
		{ID: "2", Title: "Card 2"},
	}

	mockSvc.On("GetBankCards", mock.Anything).Return(cards, nil)

	items, err := adapter.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Card 1", items[0].Title)
	mockSvc.AssertExpectations(t)
}

func TestBankCardAdapter_Get(t *testing.T) {
	mockSvc := new(MockBankCardService)
	adapter := adapters.NewBankCardAdapter(mockSvc)

	card := &model.BankCard{ID: "1", Title: "Main Card"}
	mockSvc.On("GetBankCardByID", mock.Anything, "1").Return(card, nil)

	got, err := adapter.Get(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, card, got)
	mockSvc.AssertExpectations(t)
}

func TestBankCardAdapter_Create(t *testing.T) {
	mockSvc := new(MockBankCardService)
	adapter := adapters.NewBankCardAdapter(mockSvc)

	card := &model.BankCard{ID: "1", Title: "New Card"}
	mockSvc.On("CreateBankCard", mock.Anything, card).Return(nil)

	err := adapter.Create(context.Background(), card)
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestBankCardAdapter_Update(t *testing.T) {
	mockSvc := new(MockBankCardService)
	adapter := adapters.NewBankCardAdapter(mockSvc)

	card := &model.BankCard{ID: "1", Title: "Updated Card"}
	mockSvc.On("UpdateBankCard", mock.Anything, card).Return(nil)

	err := adapter.Update(context.Background(), "1", card)
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestBankCardAdapter_Delete(t *testing.T) {
	mockSvc := new(MockBankCardService)
	adapter := adapters.NewBankCardAdapter(mockSvc)

	mockSvc.On("DeleteBankCard", mock.Anything, "1").Return(nil)

	err := adapter.Delete(context.Background(), "1")
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestBankCardAdapter_Create_InvalidType(t *testing.T) {
	mockSvc := new(MockBankCardService)
	adapter := adapters.NewBankCardAdapter(mockSvc)

	err := adapter.Create(context.Background(), "not a bank card")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected *model.BankCard")
}

func TestBankCardAdapter_Update_InvalidType(t *testing.T) {
	mockSvc := new(MockBankCardService)
	adapter := adapters.NewBankCardAdapter(mockSvc)

	err := adapter.Update(context.Background(), "1", "not a bank card")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected *model.BankCard")
}

func TestBankCardAdapter_Update_IDMismatch(t *testing.T) {
	mockSvc := new(MockBankCardService)
	adapter := adapters.NewBankCardAdapter(mockSvc)

	card := &model.BankCard{ID: "2", Title: "Card"}
	mockSvc.On("UpdateBankCard", mock.Anything, mock.MatchedBy(func(c *model.BankCard) bool {
		return c.ID == "1" // Проверяем, что ID был обновлён
	})).Return(nil)

	err := adapter.Update(context.Background(), "1", card)
	assert.NoError(t, err)
	assert.Equal(t, "1", card.ID) // Проверяем, что ID изменился
	mockSvc.AssertExpectations(t)
}
