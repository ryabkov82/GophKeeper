package auth

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/client/storage"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
)

// AuthManager управляет авторизацией пользователя, включая хранение токена,
// взаимодействие с сервером через gRPC и логирование.
type AuthManager struct {
	token      string               // Текущий токен (в памяти)
	tokenStore storage.TokenStorage // Постоянное хранилище (файл, keychain и т.д.)
	Logger     *zap.Logger
	Client     proto.AuthServiceClient // добавлено для инъекции моков
}

// AuthManagerIface описывает интерфейс для управления аутентификацией и регистрацией пользователей.
// Включает методы для регистрации, входа и других операций, связанных с авторизацией.
// Используется для абстрагирования реальной реализации AuthManager,
// что облегчает подмену в тестах и повышает модульность кода.
type AuthManagerIface interface {
	// Register регистрирует нового пользователя с заданным логином и паролем.
	// Возвращает ошибку, если регистрация не удалась.
	Register(ctx context.Context, login, password string) error

	// Login выполняет аутентификацию пользователя с заданным логином и паролем.
	// Возвращает соль для последующей генерации ключа шифрования
	// Возвращает ошибку, если вход не удался.
	Login(ctx context.Context, login, password string) ([]byte, error)

	// SetClient задаёт gRPC клиента для AuthManager.
	SetClient(client proto.AuthServiceClient)

	// GetToken возвращает текущий токен.
	GetToken() string
}

// NewAuthManager создаёт новый экземпляр AuthManager.
//
// store — реализация хранения токена,
// logger — логгер.
func NewAuthManager(
	store storage.TokenStorage,
	logger *zap.Logger,
) *AuthManager {
	return &AuthManager{
		tokenStore: store,
		Logger:     logger,
	}
}

// SetToken сохраняет токен в память и постоянное хранилище.
// Используется после успешного входа или получения нового токена.
func (a *AuthManager) SetToken(token string) error {
	a.Logger.Debug("Saving access token")
	a.token = token
	if err := a.tokenStore.Save(token); err != nil {
		a.Logger.Error("Failed to save token", zap.Error(err))
		return err
	}
	return nil
}

// GetToken возвращает текущий токен.
// Если токен отсутствует в памяти, он будет загружен из хранилища.
func (a *AuthManager) GetToken() string {
	if a.token == "" {
		a.Logger.Debug("Token not in memory, attempting to load from storage")
		token, err := a.tokenStore.Load()
		if err != nil {
			a.Logger.Warn("Failed to load token from storage", zap.Error(err))
		}
		a.token = token
	}
	return a.token
}

// Clear удаляет токен из памяти и из постоянного хранилища.
func (a *AuthManager) Clear() error {
	a.Logger.Info("Clearing access token")
	a.token = ""
	if err := a.tokenStore.Clear(); err != nil {
		a.Logger.Error("Failed to clear token", zap.Error(err))
		return err
	}
	return nil
}

// SetClient задаёт gRPC клиента для AuthManager.
//
// Этот метод используется для установки экземпляра
// proto.AuthServiceClient, который будет использоваться
// для выполнения вызовов RPC внутри AuthManager.
//
// Обычно вызывается после установления gRPC соединения,
// чтобы внедрить готовый клиент в AuthManager.
//
// client — клиент, реализующий интерфейс proto.AuthServiceClient.
func (a *AuthManager) SetClient(client proto.AuthServiceClient) {
	a.Client = client
}

// Login выполняет аутентификацию пользователя через gRPC,
// получает access token и сохраняет его в хранилище.
func (a *AuthManager) Login(ctx context.Context, login, password string) ([]byte, error) {

	a.Logger.Info("Attempting login", zap.String("login", login))

	req := &proto.LoginRequest{}
	req.SetLogin(login)
	req.SetPassword(password)

	resp, err := a.Client.Login(ctx, req)
	if err != nil {
		a.Logger.Error("Login RPC failed", zap.Error(err))
		return nil, fmt.Errorf("login RPC failed: %w", err)
	}

	if err := a.SetToken(resp.GetAccessToken()); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	salt := resp.GetSalt()

	a.Logger.Info("Login successful",
		zap.String("login", login),
	)

	return salt, nil

}

// Register выполняет регистрацию пользователя через gRPC.
func (a *AuthManager) Register(ctx context.Context, login, password string) error {
	a.Logger.Info("Attempting registration", zap.String("login", login))

	req := &proto.RegisterRequest{}
	req.SetLogin(login)
	req.SetPassword(password)

	_, err := a.Client.Register(ctx, req)
	if err != nil {
		a.Logger.Error("Register RPC failed", zap.Error(err))
		return fmt.Errorf("register RPC failed: %w", err)
	}

	a.Logger.Info("Registration successful, proceeding to login", zap.String("login", login))
	return nil
}
