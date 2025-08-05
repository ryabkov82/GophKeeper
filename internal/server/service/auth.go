package service

import (
	"context"
	"errors"

	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
	domainService "github.com/ryabkov82/gophkeeper/internal/domain/service"
	"github.com/ryabkov82/gophkeeper/internal/pkg/crypto"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtutils"
)

// AuthService — реализация domainService.AuthService
type authService struct {
	userRepo     repository.UserRepository
	tokenManager *jwtutils.TokenManager
}

// NewAuthService — конструктор, возвращает интерфейс domainService.AuthService
func NewAuthService(userRepo repository.UserRepository, tm *jwtutils.TokenManager) domainService.AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenManager: tm,
	}
}

func (s *authService) Register(ctx context.Context, login, password string) error {
	if login == "" || password == "" {
		return errors.New("login and password must not be empty")
	}

	hash, salt, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}

	return s.userRepo.CreateUser(ctx, login, hash, salt)
}

func (s *authService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	if !crypto.VerifyPassword(password, user.PasswordHash, user.Salt) {
		return "", errors.New("invalid credentials")
	}

	return s.tokenManager.GenerateToken(user.ID, login)
}
