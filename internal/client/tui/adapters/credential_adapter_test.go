package adapters_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ryabkov82/gophkeeper/internal/client/tui/adapters"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// MockCredentialService мок для CredentialService
type MockCredentialService struct {
	mock.Mock
}

func (m *MockCredentialService) CreateCredential(ctx context.Context, cred *model.Credential) error {
	args := m.Called(ctx, cred)
	return args.Error(0)
}
func (m *MockCredentialService) GetCredentialByID(ctx context.Context, id string) (*model.Credential, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Credential), args.Error(1)
}
func (m *MockCredentialService) GetCredentials(ctx context.Context) ([]model.Credential, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Credential), args.Error(1)
}
func (m *MockCredentialService) UpdateCredential(ctx context.Context, cred *model.Credential) error {
	args := m.Called(ctx, cred)
	return args.Error(0)
}
func (m *MockCredentialService) DeleteCredential(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCredentialAdapter_List(t *testing.T) {
	mockSvc := new(MockCredentialService)
	adapter := adapters.NewCredentialAdapter(mockSvc)

	creds := []model.Credential{
		{ID: "1", Title: "cred1"},
		{ID: "2", Title: "cred2"},
	}

	mockSvc.On("GetCredentials", mock.Anything).Return(creds, nil)

	items, err := adapter.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "cred1", items[0].Title)
	mockSvc.AssertExpectations(t)
}

func TestCredentialAdapter_Get(t *testing.T) {
	mockSvc := new(MockCredentialService)
	adapter := adapters.NewCredentialAdapter(mockSvc)

	cred := &model.Credential{ID: "1", Title: "cred1"}
	mockSvc.On("GetCredentialByID", mock.Anything, "1").Return(cred, nil)

	got, err := adapter.Get(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, cred, got)
	mockSvc.AssertExpectations(t)
}

func TestCredentialAdapter_Create(t *testing.T) {
	mockSvc := new(MockCredentialService)
	adapter := adapters.NewCredentialAdapter(mockSvc)

	cred := &model.Credential{ID: "1", Title: "cred1"}
	mockSvc.On("CreateCredential", mock.Anything, cred).Return(nil)

	err := adapter.Create(context.Background(), cred)
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestCredentialAdapter_Update(t *testing.T) {
	mockSvc := new(MockCredentialService)
	adapter := adapters.NewCredentialAdapter(mockSvc)

	cred := &model.Credential{ID: "1", Title: "cred1"}
	mockSvc.On("UpdateCredential", mock.Anything, cred).Return(nil)

	err := adapter.Update(context.Background(), "1", cred)
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestCredentialAdapter_Delete(t *testing.T) {
	mockSvc := new(MockCredentialService)
	adapter := adapters.NewCredentialAdapter(mockSvc)

	mockSvc.On("DeleteCredential", mock.Anything, "1").Return(nil)

	err := adapter.Delete(context.Background(), "1")
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestCredentialAdapter_Create_InvalidType(t *testing.T) {
	mockSvc := new(MockCredentialService)
	adapter := adapters.NewCredentialAdapter(mockSvc)

	err := adapter.Create(context.Background(), "not a credential")
	assert.Error(t, err)
}

func TestCredentialAdapter_Update_InvalidType(t *testing.T) {
	mockSvc := new(MockCredentialService)
	adapter := adapters.NewCredentialAdapter(mockSvc)

	err := adapter.Update(context.Background(), "1", "not a credential")
	assert.Error(t, err)
}
