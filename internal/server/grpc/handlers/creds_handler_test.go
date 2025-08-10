package handlers_test

import (
	"context"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtauth"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/ryabkov82/gophkeeper/internal/server/grpc/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CredentialServiceMock struct {
	mock.Mock
}

func (m *CredentialServiceMock) Create(ctx context.Context, cred *model.Credential) error {
	args := m.Called(ctx, cred)
	return args.Error(0)
}

func (m *CredentialServiceMock) GetByID(ctx context.Context, id string) (*model.Credential, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Credential), args.Error(1)
}

func (m *CredentialServiceMock) GetByUserID(ctx context.Context, userID string) ([]model.Credential, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.Credential), args.Error(1)
}

func (m *CredentialServiceMock) Update(ctx context.Context, cred *model.Credential) error {
	args := m.Called(ctx, cred)
	return args.Error(0)
}

func (m *CredentialServiceMock) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func contextWithUserID(userID string) context.Context {
	return jwtauth.WithUserID(context.Background(), userID)
}

func TestCreateCredential_Success(t *testing.T) {
	userID := "user-123"
	mockService := new(CredentialServiceMock)
	h := handlers.NewCredentialHandler(mockService, zap.NewNop())

	credProto := &pb.Credential{}
	credProto.SetTitle("Gmail")
	credProto.SetLogin("user@gmail.com")
	credProto.SetPassword("encryptedpass")
	credProto.SetMetadata("some meta")

	req := &pb.CreateCredentialRequest{}
	req.SetCredential(credProto)

	ctx := contextWithUserID(userID)

	mockService.On("Create", ctx, mock.MatchedBy(func(c *model.Credential) bool {
		return c.UserID == userID && c.Title == "Gmail" && c.Login == "user@gmail.com"
	})).Return(nil)

	resp, err := h.CreateCredential(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockService.AssertExpectations(t)
}

func TestCreateCredential_Unauthenticated(t *testing.T) {
	mockService := new(CredentialServiceMock)
	h := handlers.NewCredentialHandler(mockService, zap.NewNop())

	req := &pb.CreateCredentialRequest{}

	ctx := context.Background() // Без userID

	resp, err := h.CreateCredential(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestGetCredentialByID_Success(t *testing.T) {
	userID := "user-123"
	credID := "cred-1"
	mockService := new(CredentialServiceMock)
	h := handlers.NewCredentialHandler(mockService, zap.NewNop())

	cred := &model.Credential{
		ID:     credID,
		UserID: userID,
		Title:  "GitHub",
		Login:  "user",
	}

	req := &pb.GetCredentialByIDRequest{}
	req.SetId(credID)

	ctx := contextWithUserID(userID)

	mockService.On("GetByID", ctx, credID).Return(cred, nil)

	resp, err := h.GetCredentialByID(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, credID, resp.GetCredential().GetId())
	mockService.AssertExpectations(t)
}

func TestGetCredentialByID_NotFound(t *testing.T) {
	userID := "user-123"
	credID := "cred-1"
	mockService := new(CredentialServiceMock)
	h := handlers.NewCredentialHandler(mockService, zap.NewNop())

	req := &pb.GetCredentialByIDRequest{}
	req.SetId(credID)

	ctx := contextWithUserID(userID)

	mockService.On("GetByID", ctx, credID).Return((*model.Credential)(nil), nil)

	resp, err := h.GetCredentialByID(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestGetCredentialsByUserID_Success(t *testing.T) {
	userID := "user-123"
	mockService := new(CredentialServiceMock)
	h := handlers.NewCredentialHandler(mockService, zap.NewNop())

	creds := []model.Credential{
		{ID: "1", UserID: userID, Title: "Site1", Login: "login1"},
		{ID: "2", UserID: userID, Title: "Site2", Login: "login2"},
	}

	req := &pb.GetCredentialsByUserIDRequest{}

	ctx := contextWithUserID(userID)

	mockService.On("GetByUserID", ctx, userID).Return(creds, nil)

	resp, err := h.GetCredentialsByUserID(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.GetCredentials(), 2)
	mockService.AssertExpectations(t)
}

func TestUpdateCredential_Success(t *testing.T) {
	userID := "user-123"
	credID := "cred-1"
	mockService := new(CredentialServiceMock)
	h := handlers.NewCredentialHandler(mockService, zap.NewNop())

	credProto := &pb.Credential{}
	credProto.SetId(credID)
	credProto.SetTitle("UpdatedTitle")
	credProto.SetLogin("newlogin")
	credProto.SetPassword("newpass")
	credProto.SetMetadata("new meta")

	req := &pb.UpdateCredentialRequest{}
	req.SetCredential(credProto)

	ctx := contextWithUserID(userID)

	existing := &model.Credential{
		ID:     credID,
		UserID: userID,
	}

	mockService.On("GetByID", ctx, credID).Return(existing, nil)
	mockService.On("Update", ctx, mock.AnythingOfType("*model.Credential")).Return(nil)

	resp, err := h.UpdateCredential(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockService.AssertExpectations(t)
}

func TestDeleteCredential_Success(t *testing.T) {
	userID := "user-123"
	credID := "cred-1"
	mockService := new(CredentialServiceMock)
	h := handlers.NewCredentialHandler(mockService, zap.NewNop())

	req := &pb.DeleteCredentialRequest{}
	req.SetId(credID)

	ctx := contextWithUserID(userID)

	existing := &model.Credential{
		ID:     credID,
		UserID: userID,
	}

	mockService.On("GetByID", ctx, credID).Return(existing, nil)
	mockService.On("Delete", ctx, credID).Return(nil)

	resp, err := h.DeleteCredential(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockService.AssertExpectations(t)
}
