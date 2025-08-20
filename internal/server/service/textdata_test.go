package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/server/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Мок репозитория TextData
type mockTextDataRepo struct {
	mock.Mock
}

func (m *mockTextDataRepo) Create(ctx context.Context, data *model.TextData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *mockTextDataRepo) GetByID(ctx context.Context, userID, id string) (*model.TextData, error) {
	args := m.Called(ctx, userID, id)
	data := args.Get(0)
	if data == nil {
		return nil, args.Error(1)
	}
	return data.(*model.TextData), args.Error(1)
}

func (m *mockTextDataRepo) ListTitles(ctx context.Context, userID string) ([]*model.TextData, error) {
	args := m.Called(ctx, userID)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.([]*model.TextData), args.Error(1)
}

func (m *mockTextDataRepo) Update(ctx context.Context, data *model.TextData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *mockTextDataRepo) Delete(ctx context.Context, userID, id string) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

// ----------------------- Тесты -----------------------

func TestTextDataService_Create(t *testing.T) {
	mockRepo := new(mockTextDataRepo)
	svc := service.NewTextDataService(mockRepo)

	data := &model.TextData{
		UserID:  uuid.NewString(),
		Title:   "Test",
		Content: []byte("secret"),
	}

	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	err := svc.Create(context.Background(), data)
	assert.NoError(t, err)
	assert.NotEmpty(t, data.ID)
	assert.WithinDuration(t, time.Now(), data.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), data.UpdatedAt, time.Second)
}

func TestTextDataService_Create_ValidationError(t *testing.T) {
	mockRepo := new(mockTextDataRepo)
	svc := service.NewTextDataService(mockRepo)

	data := &model.TextData{UserID: uuid.NewString()}

	err := svc.Create(context.Background(), data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title is required")
}

func TestTextDataService_GetByID(t *testing.T) {
	mockRepo := new(mockTextDataRepo)
	svc := service.NewTextDataService(mockRepo)

	id := uuid.NewString()
	userID := uuid.NewString()
	data := &model.TextData{ID: id, UserID: userID}

	mockRepo.On("GetByID", mock.Anything, userID, id).Return(data, nil)

	result, err := svc.GetByID(context.Background(), userID, id)
	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, userID, result.UserID)
}

func TestTextDataService_GetByID_Error(t *testing.T) {
	mockRepo := new(mockTextDataRepo)
	svc := service.NewTextDataService(mockRepo)

	mockRepo.On("GetByID", mock.Anything, "user", "unknown").Return(nil, errors.New("not found"))

	_, err := svc.GetByID(context.Background(), "user", "unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestTextDataService_ListTitles(t *testing.T) {
	mockRepo := new(mockTextDataRepo)
	svc := service.NewTextDataService(mockRepo)

	userID := uuid.NewString()
	dataList := []*model.TextData{
		{ID: uuid.NewString(), Title: "One"},
		{ID: uuid.NewString(), Title: "Two"},
	}

	mockRepo.On("ListTitles", mock.Anything, userID).Return(dataList, nil)

	result, err := svc.ListTitles(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "One", result[0].Title)
}

func TestTextDataService_Update(t *testing.T) {
	mockRepo := new(mockTextDataRepo)
	svc := service.NewTextDataService(mockRepo)

	data := &model.TextData{ID: uuid.NewString(), UserID: uuid.NewString()}
	mockRepo.On("GetByID", mock.Anything, data.UserID, data.ID).Return(data, nil)
	mockRepo.On("Update", mock.Anything, data).Return(nil)

	err := svc.Update(context.Background(), data)
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now(), data.UpdatedAt, time.Second)
}

func TestTextDataService_Update_NotFound(t *testing.T) {
	mockRepo := new(mockTextDataRepo)
	svc := service.NewTextDataService(mockRepo)

	data := &model.TextData{ID: uuid.NewString(), UserID: uuid.NewString()}
	mockRepo.On("GetByID", mock.Anything, data.UserID, data.ID).Return(nil, nil)
	mockRepo.On("Update", mock.Anything, data).Return(fmt.Errorf("text data not found"))

	err := svc.Update(context.Background(), data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text data not found")
}

func TestTextDataService_Delete(t *testing.T) {
	mockRepo := new(mockTextDataRepo)
	svc := service.NewTextDataService(mockRepo)

	id := uuid.NewString()
	userID := uuid.NewString()
	data := &model.TextData{ID: id, UserID: userID}

	mockRepo.On("GetByID", mock.Anything, userID, id).Return(data, nil)
	mockRepo.On("Delete", mock.Anything, userID, id).Return(nil)

	err := svc.Delete(context.Background(), userID, id)
	assert.NoError(t, err)
}

func TestTextDataService_Delete_NotFound(t *testing.T) {
	mockRepo := new(mockTextDataRepo)
	svc := service.NewTextDataService(mockRepo)

	id := uuid.NewString()
	userID := uuid.NewString()

	mockRepo.On("GetByID", mock.Anything, userID, id).Return(nil, nil)

	err := svc.Delete(context.Background(), userID, id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "text data not found")
}

func TestTextDataService_ListTitles_Empty(t *testing.T) {
	mockRepo := new(mockTextDataRepo)
	svc := service.NewTextDataService(mockRepo)

	userID := uuid.NewString()

	// Возвращаем пустой слайс
	mockRepo.On("ListTitles", mock.Anything, userID).Return([]*model.TextData{}, nil)

	result, err := svc.ListTitles(context.Background(), userID)
	assert.NoError(t, err)
	assert.Empty(t, result) // Должен быть пустой
}

func TestTextDataService_ListTitles_Error(t *testing.T) {
	mockRepo := new(mockTextDataRepo)
	svc := service.NewTextDataService(mockRepo)

	userID := uuid.NewString()

	// Возвращаем ошибку из репозитория
	mockRepo.On("ListTitles", mock.Anything, userID).Return(nil, errors.New("db error"))

	result, err := svc.ListTitles(context.Background(), userID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "db error")
}
