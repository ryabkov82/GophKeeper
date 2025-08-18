package auth_test

import (
	"context"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/service/auth"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// Заглушка для TokenStorage
type mockTokenStorage struct {
	token string
}

func (m *mockTokenStorage) Save(token string) error {
	m.token = token
	return nil
}
func (m *mockTokenStorage) Load() (string, error) {
	return m.token, nil
}
func (m *mockTokenStorage) Clear() error {
	m.token = ""
	return nil
}

func TestAuthManager_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAuthServiceClient(ctrl)

	resp := &proto.LoginResponse{}
	resp.SetAccessToken("testtoken")
	resp.SetSalt([]byte("salt"))

	mockClient.EXPECT().
		Login(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(resp, nil).
		Times(1)

	store := &mockTokenStorage{}

	authMgr := auth.NewAuthManager(store, zap.NewNop())
	authMgr.Client = mockClient // инжектим мок клиента

	salt, err := authMgr.Login(context.Background(), "user", "pass")
	require.NoError(t, err)

	require.Equal(t, "testtoken", authMgr.GetToken())
	require.Equal(t, "salt", string(salt))
}

func TestAuthManager_SetToken(t *testing.T) {
	store := &mockTokenStorage{}
	authMgr := auth.NewAuthManager(store, zap.NewNop())

	err := authMgr.SetToken("mytoken")
	require.NoError(t, err)
	require.Equal(t, "mytoken", authMgr.GetToken())
	require.Equal(t, "mytoken", store.token)
}

func TestAuthManager_GetToken_LoadsFromStorage(t *testing.T) {
	store := &mockTokenStorage{token: "storedtoken"}
	authMgr := auth.NewAuthManager(store, zap.NewNop())

	token := authMgr.GetToken()
	require.Equal(t, "storedtoken", token)
}

func TestAuthManager_Clear(t *testing.T) {
	store := &mockTokenStorage{token: "someToken"}
	authMgr := auth.NewAuthManager(store, zap.NewNop())
	authMgr.SetToken("someToken")

	err := authMgr.Clear()
	require.NoError(t, err)
	require.Equal(t, "", authMgr.GetToken())
	require.Equal(t, "", store.token)
}

func TestAuthManager_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockAuthServiceClient(ctrl)
	mockClient.EXPECT().
		Register(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&proto.RegisterResponse{}, nil).
		Times(1)

	store := &mockTokenStorage{}

	authMgr := auth.NewAuthManager(store, zap.NewNop())
	authMgr.Client = mockClient

	err := authMgr.Register(context.Background(), "user", "pass")
	require.NoError(t, err)
}
