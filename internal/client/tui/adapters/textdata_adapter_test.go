package adapters_test

import (
	"context"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/tui/adapters"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTextDataService — мок для TextDataService
type MockTextDataService struct {
	mock.Mock
}

func (m *MockTextDataService) CreateTextData(ctx context.Context, text *model.TextData) error {
	args := m.Called(ctx, text)
	return args.Error(0)
}

func (m *MockTextDataService) GetTextDataByID(ctx context.Context, id string) (*model.TextData, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.TextData), args.Error(1)
}

func (m *MockTextDataService) GetTextDataTitles(ctx context.Context) ([]*model.TextData, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*model.TextData), args.Error(1)
}

func (m *MockTextDataService) UpdateTextData(ctx context.Context, text *model.TextData) error {
	args := m.Called(ctx, text)
	return args.Error(0)
}

func (m *MockTextDataService) DeleteTextData(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTextDataAdapter_List(t *testing.T) {
	mockSvc := new(MockTextDataService)
	adapter := adapters.NewTextDataAdapter(mockSvc)

	texts := []*model.TextData{
		{ID: "1", Title: "Text 1"},
		{ID: "2", Title: "Text 2"},
	}

	mockSvc.On("GetTextDataTitles", mock.Anything).Return(texts, nil)

	items, err := adapter.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Text 1", items[0].Title)
	mockSvc.AssertExpectations(t)
}

func TestTextDataAdapter_Get(t *testing.T) {
	mockSvc := new(MockTextDataService)
	adapter := adapters.NewTextDataAdapter(mockSvc)

	text := &model.TextData{ID: "1", Title: "My Text"}
	mockSvc.On("GetTextDataByID", mock.Anything, "1").Return(text, nil)

	got, err := adapter.Get(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, text, got)
	mockSvc.AssertExpectations(t)
}

func TestTextDataAdapter_Create(t *testing.T) {
	mockSvc := new(MockTextDataService)
	adapter := adapters.NewTextDataAdapter(mockSvc)

	text := &model.TextData{ID: "1", Title: "New Text"}
	mockSvc.On("CreateTextData", mock.Anything, text).Return(nil)

	err := adapter.Create(context.Background(), text)
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestTextDataAdapter_Update(t *testing.T) {
	mockSvc := new(MockTextDataService)
	adapter := adapters.NewTextDataAdapter(mockSvc)

	text := &model.TextData{ID: "1", Title: "Updated Text"}
	mockSvc.On("UpdateTextData", mock.Anything, text).Return(nil)

	err := adapter.Update(context.Background(), "1", text)
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestTextDataAdapter_Delete(t *testing.T) {
	mockSvc := new(MockTextDataService)
	adapter := adapters.NewTextDataAdapter(mockSvc)

	mockSvc.On("DeleteTextData", mock.Anything, "1").Return(nil)

	err := adapter.Delete(context.Background(), "1")
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestTextDataAdapter_Create_InvalidType(t *testing.T) {
	mockSvc := new(MockTextDataService)
	adapter := adapters.NewTextDataAdapter(mockSvc)

	err := adapter.Create(context.Background(), "not a text data")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected *model.TextData")
}

func TestTextDataAdapter_Update_InvalidType(t *testing.T) {
	mockSvc := new(MockTextDataService)
	adapter := adapters.NewTextDataAdapter(mockSvc)

	err := adapter.Update(context.Background(), "1", "not a text data")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected *model.TextData")
}

func TestTextDataAdapter_Update_IDMismatch(t *testing.T) {
	mockSvc := new(MockTextDataService)
	adapter := adapters.NewTextDataAdapter(mockSvc)

	text := &model.TextData{ID: "2", Title: "Some Text"}
	mockSvc.On("UpdateTextData", mock.Anything, mock.MatchedBy(func(t *model.TextData) bool {
		return t.ID == "1" // Проверяем, что ID был обновлён
	})).Return(nil)

	err := adapter.Update(context.Background(), "1", text)
	assert.NoError(t, err)
	assert.Equal(t, "1", text.ID) // Проверяем, что ID изменился
	mockSvc.AssertExpectations(t)
}
