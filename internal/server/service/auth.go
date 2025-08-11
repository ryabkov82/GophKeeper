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

// Register выполняет регистрацию нового пользователя.
//
// Выполняет валидацию входных данных (логин и пароль не должны быть пустыми),
// хеширует пароль с солью и сохраняет пользователя в хранилище.
//
// Параметры:
//   - ctx: контекст выполнения (может содержать таймаут или отмену);
//   - login: логин пользователя;
//   - password: пароль пользователя (в открытом виде).
//
// Возвращает ошибку, если:
//   - логин или пароль пустые;
//   - произошла ошибка при хешировании пароля;
//   - не удалось создать пользователя в хранилище.
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

// Login выполняет аутентификацию пользователя.
//
// Получает пользователя из хранилища по логину, сравнивает сохранённый хеш пароля
// с введённым паролем, и в случае успеха — генерирует JWT-токен.
//
// Параметры:
//   - ctx: контекст выполнения (может содержать таймаут или отмену);
//   - login: логин пользователя;
//   - password: введённый пользователем пароль (в открытом виде).
//
// Возвращает:
//   - строку с JWT-токеном в случае успеха;
//   - срез байт с солью пользователя ([]byte);
//   - ошибку, если пользователь не найден, пароль не совпадает,
//     либо возникли проблемы при генерации токена.
func (s *authService) Login(ctx context.Context, login, password string) (string, []byte, error) {
	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", nil, err
	}

	if !crypto.VerifyPassword(password, user.PasswordHash, user.Salt) {
		return "", nil, errors.New("invalid credentials")
	}

	token, err := s.tokenManager.GenerateToken(user.ID, login)
	if err != nil {
		return "", nil, err
	}

	return token, []byte(user.Salt), nil
}
