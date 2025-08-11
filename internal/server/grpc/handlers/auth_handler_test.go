package handlers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/ryabkov82/gophkeeper/internal/server/grpc/handlers"
)

// --- мок AuthService ---
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, login, password string) error {
	args := m.Called(ctx, login, password)
	return args.Error(0)
}

func (m *mockAuthService) Login(ctx context.Context, login, password string) (string, []byte, error) {
	args := m.Called(ctx, login, password)
	var salt []byte
	if s, ok := args.Get(1).([]byte); ok {
		salt = s
	}
	return args.String(0), salt, args.Error(2)
}

func TestAuthHandler_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockSvc.On("Register", ctx, "testuser", "testpass").Return(nil)

		handler := handlers.NewAuthHandler(mockSvc, zap.NewNop())
		req := &api.RegisterRequest{}
		req.SetLogin("testuser")
		req.SetPassword("testpass")

		resp, err := handler.Register(ctx, req)
		require.NoError(t, err)
		require.Equal(t, "user registered successfully", resp.GetMessage())

		mockSvc.AssertExpectations(t)
	})

	t.Run("error from service", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockSvc.On("Register", ctx, "baduser", "badpass").
			Return(errors.New("user exists"))

		handler := handlers.NewAuthHandler(mockSvc, zap.NewNop())
		req := &api.RegisterRequest{}
		req.SetLogin("baduser")
		req.SetPassword("badpass")

		resp, err := handler.Register(ctx, req)
		require.Nil(t, resp)

		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
		require.Contains(t, st.Message(), "user exists")

		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockSvc.On("Login", ctx, "testuser", "testpass").
			Return("token123", []byte("mysalt"), nil)

		handler := handlers.NewAuthHandler(mockSvc, zap.NewNop())
		req := &api.LoginRequest{}
		req.SetLogin("testuser")
		req.SetPassword("testpass")

		resp, err := handler.Login(ctx, req)
		require.NoError(t, err)
		require.Equal(t, "token123", resp.GetAccessToken())

		mockSvc.AssertExpectations(t)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockSvc.On("Login", ctx, "baduser", "wrongpass").
			Return("", nil, errors.New("invalid credentials"))

		handler := handlers.NewAuthHandler(mockSvc, zap.NewNop())
		req := &api.LoginRequest{}
		req.SetLogin("baduser")
		req.SetPassword("wrongpass")

		resp, err := handler.Login(ctx, req)
		require.Nil(t, resp)

		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.Unauthenticated, st.Code())
		require.Contains(t, st.Message(), "invalid credentials")

		mockSvc.AssertExpectations(t)
	})
}
