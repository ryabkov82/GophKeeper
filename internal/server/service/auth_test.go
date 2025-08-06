package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/crypto"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtutils"
	"github.com/ryabkov82/gophkeeper/internal/server/service"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) CreateUser(ctx context.Context, login, hash, salt string) error {
	args := m.Called(ctx, login, hash, salt)
	return args.Error(0)
}

func (m *mockUserRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	args := m.Called(ctx, login)
	if user, ok := args.Get(0).(*model.User); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	tm := jwtutils.New("testsecretstringthatlongenough!!!", time.Minute)
	mockRepo := new(mockUserRepository)
	svc := service.NewAuthService(mockRepo, tm)

	t.Run("empty login or password", func(t *testing.T) {
		err := svc.Register(context.Background(), "", "pass")
		require.Error(t, err)
		err = svc.Register(context.Background(), "user", "")
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		login := "user"
		password := "pass1234"
		// игнорируем, т.к. нам нужен just call tracking
		mockRepo.On("CreateUser", mock.Anything, login, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Once()

		err := svc.Register(ctx, login, password)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	tm := jwtutils.New("testsecretstringthatlongenough!!!", time.Minute)
	mockRepo := new(mockUserRepository)
	svc := service.NewAuthService(mockRepo, tm)

	ctx := context.Background()
	login := "user"
	password := "password123"
	hash, salt, _ := crypto.HashPassword(password)
	user := &model.User{
		ID:           "123",
		Login:        login,
		PasswordHash: hash,
		Salt:         salt,
	}

	t.Run("user not found", func(t *testing.T) {
		mockRepo.On("GetUserByLogin", mock.Anything, login).Return(nil, errors.New("not found")).Once()
		_, err := svc.Login(ctx, login, password)
		require.Error(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		mockRepo.On("GetUserByLogin", mock.Anything, login).Return(user, nil).Once()
		_, err := svc.Login(ctx, login, "wrongpass")
		require.Error(t, err)
		require.EqualError(t, err, "invalid credentials")
		mockRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo.On("GetUserByLogin", mock.Anything, login).Return(user, nil).Once()
		token, err := svc.Login(ctx, login, password)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		mockRepo.AssertExpectations(t)
	})
}
