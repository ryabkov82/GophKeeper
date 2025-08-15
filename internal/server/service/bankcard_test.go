package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/server/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок репозитория банковских карт
type mockBankCardRepo struct {
	mock.Mock
}

func (m *mockBankCardRepo) Create(ctx context.Context, card *model.BankCard) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}

func (m *mockBankCardRepo) GetByID(ctx context.Context, id string) (*model.BankCard, error) {
	args := m.Called(ctx, id)
	card := args.Get(0)
	if card == nil {
		return nil, args.Error(1)
	}
	return card.(*model.BankCard), args.Error(1)
}

func (m *mockBankCardRepo) GetByUser(ctx context.Context, userID string) ([]model.BankCard, error) {
	args := m.Called(ctx, userID)
	cards := args.Get(0)
	if cards == nil {
		return nil, args.Error(1)
	}
	return cards.([]model.BankCard), args.Error(1)
}

func (m *mockBankCardRepo) Update(ctx context.Context, card *model.BankCard) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}

func (m *mockBankCardRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestBankCardService_Create(t *testing.T) {
	mockRepo := new(mockBankCardRepo)
	svc := service.NewBankCardService(mockRepo)

	card := &model.BankCard{
		CardNumber:     "1234123412341234",
		CardholderName: "John Doe",
		ExpiryDate:     "12/30",
	}

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	err := svc.Create(context.Background(), card)
	assert.NoError(t, err)
	assert.NotEmpty(t, card.ID)
	assert.WithinDuration(t, time.Now(), card.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), card.UpdatedAt, time.Second)
}

func TestBankCardService_Create_ValidationError(t *testing.T) {
	mockRepo := new(mockBankCardRepo)
	svc := service.NewBankCardService(mockRepo)

	card := &model.BankCard{}

	err := svc.Create(context.Background(), card)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "card number is required")
}

func TestBankCardService_GetByID(t *testing.T) {
	mockRepo := new(mockBankCardRepo)
	svc := service.NewBankCardService(mockRepo)

	id := uuid.NewString()
	card := &model.BankCard{ID: id}

	mockRepo.On("GetByID", mock.Anything, id).Return(card, nil)

	result, err := svc.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)
}

func TestBankCardService_GetByID_Error(t *testing.T) {
	mockRepo := new(mockBankCardRepo)
	svc := service.NewBankCardService(mockRepo)

	mockRepo.On("GetByID", mock.Anything, "unknown").Return(nil, errors.New("not found"))

	_, err := svc.GetByID(context.Background(), "unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBankCardService_GetByUserID(t *testing.T) {
	mockRepo := new(mockBankCardRepo)
	svc := service.NewBankCardService(mockRepo)

	userID := uuid.NewString()
	cards := []model.BankCard{{ID: uuid.NewString()}, {ID: uuid.NewString()}}

	mockRepo.On("GetByUser", mock.Anything, userID).Return(cards, nil)

	result, err := svc.GetByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestBankCardService_Update(t *testing.T) {
	mockRepo := new(mockBankCardRepo)
	svc := service.NewBankCardService(mockRepo)

	card := &model.BankCard{ID: uuid.NewString()}
	mockRepo.On("GetByID", mock.Anything, card.ID).Return(card, nil)
	mockRepo.On("Update", mock.Anything, card).Return(nil)

	err := svc.Update(context.Background(), card)
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now(), card.UpdatedAt, time.Second)
}

func TestBankCardService_Update_NotFound(t *testing.T) {
	mockRepo := new(mockBankCardRepo)
	svc := service.NewBankCardService(mockRepo)

	card := &model.BankCard{ID: uuid.NewString()}
	mockRepo.On("GetByID", mock.Anything, card.ID).Return(nil, nil)

	err := svc.Update(context.Background(), card)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bank card not found")
}

func TestBankCardService_Delete(t *testing.T) {
	mockRepo := new(mockBankCardRepo)
	svc := service.NewBankCardService(mockRepo)

	id := uuid.NewString()
	card := &model.BankCard{ID: id}

	mockRepo.On("GetByID", mock.Anything, id).Return(card, nil)
	mockRepo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.Delete(context.Background(), id)
	assert.NoError(t, err)
}

func TestBankCardService_Delete_NotFound(t *testing.T) {
	mockRepo := new(mockBankCardRepo)
	svc := service.NewBankCardService(mockRepo)

	id := uuid.NewString()
	mockRepo.On("GetByID", mock.Anything, id).Return(nil, nil)

	err := svc.Delete(context.Background(), id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bank card not found")
}
