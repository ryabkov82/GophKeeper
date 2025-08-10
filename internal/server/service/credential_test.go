package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/server/service"
)

// Мок репозитория
type MockCredentialRepository struct {
	mock.Mock
}

func (m *MockCredentialRepository) Create(ctx context.Context, cred *model.Credential) error {
	args := m.Called(ctx, cred)
	return args.Error(0)
}

func (m *MockCredentialRepository) GetByID(ctx context.Context, id string) (*model.Credential, error) {
	args := m.Called(ctx, id)
	if cred, ok := args.Get(0).(*model.Credential); ok {
		return cred, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCredentialRepository) GetByUserID(ctx context.Context, userID string) ([]model.Credential, error) {
	args := m.Called(ctx, userID)
	if creds, ok := args.Get(0).([]model.Credential); ok {
		return creds, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCredentialRepository) Update(ctx context.Context, cred *model.Credential) error {
	args := m.Called(ctx, cred)
	return args.Error(0)
}

func (m *MockCredentialRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCredentialService_Create(t *testing.T) {
	mockRepo := new(MockCredentialRepository)
	svc := service.NewCredentialService(mockRepo)

	cred := &model.Credential{UserID: "user1"}

	// Ожидание вызова Create и возврат nil (успех)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Credential")).Return(nil)

	err := svc.Create(context.Background(), cred)
	assert.NoError(t, err)
	assert.NotEmpty(t, cred.ID) // должен сгенерироваться UUID
	assert.False(t, cred.CreatedAt.IsZero())
	assert.False(t, cred.UpdatedAt.IsZero())

	mockRepo.AssertExpectations(t)
}

func TestCredentialService_GetByID(t *testing.T) {
	mockRepo := new(MockCredentialRepository)
	svc := service.NewCredentialService(mockRepo)

	testID := uuid.NewString()
	expectedCred := &model.Credential{ID: testID, UserID: "user1"}

	mockRepo.On("GetByID", mock.Anything, testID).Return(expectedCred, nil)

	cred, err := svc.GetByID(context.Background(), testID)
	assert.NoError(t, err)
	assert.Equal(t, expectedCred, cred)

	mockRepo.AssertExpectations(t)
}

func TestCredentialService_GetByID_EmptyID(t *testing.T) {
	mockRepo := new(MockCredentialRepository)
	svc := service.NewCredentialService(mockRepo)

	cred, err := svc.GetByID(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, cred)
}

func TestCredentialService_GetByUserID(t *testing.T) {
	mockRepo := new(MockCredentialRepository)
	svc := service.NewCredentialService(mockRepo)

	userID := "user1"
	expectedCreds := []model.Credential{
		{ID: uuid.NewString(), UserID: userID, Title: "title1"},
		{ID: uuid.NewString(), UserID: userID, Title: "title2"},
	}

	mockRepo.On("GetByUserID", mock.Anything, userID).Return(expectedCreds, nil)

	creds, err := svc.GetByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Equal(t, expectedCreds, creds)

	mockRepo.AssertExpectations(t)
}

func TestCredentialService_GetByUserID_EmptyUserID(t *testing.T) {
	mockRepo := new(MockCredentialRepository)
	svc := service.NewCredentialService(mockRepo)

	creds, err := svc.GetByUserID(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, creds)
}

func TestCredentialService_Update(t *testing.T) {
	mockRepo := new(MockCredentialRepository)
	svc := service.NewCredentialService(mockRepo)

	cred := &model.Credential{ID: uuid.NewString(), UserID: "user1"}

	mockRepo.On("Update", mock.Anything, cred).Return(nil)

	err := svc.Update(context.Background(), cred)
	assert.NoError(t, err)
	assert.False(t, cred.UpdatedAt.IsZero())

	mockRepo.AssertExpectations(t)
}

func TestCredentialService_Update_EmptyID(t *testing.T) {
	mockRepo := new(MockCredentialRepository)
	svc := service.NewCredentialService(mockRepo)

	err := svc.Update(context.Background(), &model.Credential{ID: ""})
	assert.Error(t, err)
}

func TestCredentialService_Delete(t *testing.T) {
	mockRepo := new(MockCredentialRepository)
	svc := service.NewCredentialService(mockRepo)

	testID := uuid.NewString()

	mockRepo.On("Delete", mock.Anything, testID).Return(nil)

	err := svc.Delete(context.Background(), testID)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestCredentialService_Delete_EmptyID(t *testing.T) {
	mockRepo := new(MockCredentialRepository)
	svc := service.NewCredentialService(mockRepo)

	err := svc.Delete(context.Background(), "")
	assert.Error(t, err)
}
