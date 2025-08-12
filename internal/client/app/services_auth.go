package app

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
)

// ensureAuthClient гарантирует создание gRPC клиента для Auth сервиса и установку его в AuthManager.
//
// ctx — контекст запроса.
//
// Возвращает ошибку, если подключение не удалось установить.
func (s *AppServices) ensureAuthClient(ctx context.Context) error {
	conn, err := s.getGRPCConn(ctx)
	if err != nil {
		return err
	}
	client := proto.NewAuthServiceClient(conn)
	s.AuthManager.SetClient(client)
	return nil
}

// LoginUser выполняет аутентификацию пользователя с указанным логином и паролем.
//
// Выполняет подключение к Auth сервису, получает соль,
// генерирует и сохраняет криптографический ключ.
//
// ctx — контекст запроса.
// login — логин пользователя.
// password — пароль пользователя.
//
// Возвращает ошибку при неудачной аутентификации или генерации ключа.
func (s *AppServices) LoginUser(ctx context.Context, login, password string) error {

	if err := s.ensureAuthClient(ctx); err != nil {
		return err
	}

	salt, err := s.AuthManager.Login(ctx, login, password)
	if err != nil {
		return err
	}

	if len(salt) == 0 {
		return fmt.Errorf("no salt received from server")
	}

	if err := s.CryptoKeyManager.GenerateAndSaveKey(password, salt); err != nil {
		return fmt.Errorf("failed to generate encryption key: %w", err)
	}

	s.Logger.Info("User logged in and encryption key saved", zap.String("login", login))
	return nil
}

// RegisterUser регистрирует нового пользователя с заданным логином и паролем,
// а затем автоматически выполняет вход.
//
// ctx — контекст запроса.
// login — логин пользователя.
// password — пароль пользователя.
//
// Возвращает ошибку при неудачной регистрации или аутентификации.
func (s *AppServices) RegisterUser(ctx context.Context, login, password string) error {

	if err := s.ensureAuthClient(ctx); err != nil {
		return err
	}

	err := s.AuthManager.Register(ctx, login, password)
	if err != nil {
		return err
	}

	return s.LoginUser(ctx, login, password)

}
