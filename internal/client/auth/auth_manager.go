package auth

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/client/connection"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
)

// AuthManager управляет авторизацией пользователя, включая хранение токена,
// взаимодействие с сервером через gRPC и логирование.
type AuthManager struct {
	token       string       // Текущий токен (в памяти)
	tokenStore  TokenStorage // Постоянное хранилище (файл, keychain и т.д.)
	connManager *connection.Manager
	Logger      *zap.Logger
}

// TokenStorage описывает интерфейс для сохранения, загрузки и очистки токена.
type TokenStorage interface {
	Save(token string) error
	Load() (string, error)
	Clear() error
}

// NewAuthManager создаёт новый экземпляр AuthManager.
//
// connManager — менеджер подключения к gRPC-серверу,
// store — реализация хранения токена,
// logger — логгер.
func NewAuthManager(connManager *connection.Manager, store TokenStorage, logger *zap.Logger) *AuthManager {
	return &AuthManager{
		tokenStore:  store,
		connManager: connManager,
		Logger:      logger,
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

// Login выполняет аутентификацию пользователя через gRPC,
// получает access token и сохраняет его в хранилище.
func (a *AuthManager) Login(ctx context.Context, login, password string) error {
	a.Logger.Info("Attempting login", zap.String("login", login))

	conn, err := a.connManager.Connect(ctx)
	if err != nil {
		a.Logger.Error("Connection failed", zap.Error(err))
		return fmt.Errorf("connection failed: %w", err)
	}

	req := &proto.LoginRequest{}
	req.SetLogin(login)
	req.SetPassword(password)

	client := proto.NewAuthServiceClient(conn)
	resp, err := client.Login(ctx, req)
	if err != nil {
		a.Logger.Error("Login RPC failed", zap.Error(err))
		return fmt.Errorf("login RPC failed: %w", err)
	}

	if err := a.SetToken(resp.GetAccessToken()); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	a.Logger.Info("Login successful", zap.String("login", login))
	return nil
}

// Register выполняет регистрацию пользователя через gRPC.
// После успешной регистрации автоматически выполняет вход.
func (a *AuthManager) Register(ctx context.Context, login, password string) error {
	a.Logger.Info("Attempting registration", zap.String("login", login))

	conn, err := a.connManager.Connect(ctx)
	if err != nil {
		a.Logger.Error("Connection failed", zap.Error(err))
		return fmt.Errorf("connection failed: %w", err)
	}

	req := &proto.RegisterRequest{}
	req.SetLogin(login)
	req.SetPassword(password)

	client := proto.NewAuthServiceClient(conn)
	_, err = client.Register(ctx, req)
	if err != nil {
		a.Logger.Error("Register RPC failed", zap.Error(err))
		return fmt.Errorf("register RPC failed: %w", err)
	}

	a.Logger.Info("Registration successful, proceeding to login", zap.String("login", login))
	return a.Login(ctx, login, password)
}
