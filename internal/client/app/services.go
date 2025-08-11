package app

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/config"
	"github.com/ryabkov82/gophkeeper/internal/client/connection"
	"github.com/ryabkov82/gophkeeper/internal/client/crypto"
	"github.com/ryabkov82/gophkeeper/internal/client/service/auth"
	"github.com/ryabkov82/gophkeeper/internal/client/service/credential"
	"github.com/ryabkov82/gophkeeper/internal/client/service/cryptokey"
	"github.com/ryabkov82/gophkeeper/internal/client/storage"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/logger"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
)

// AppServices представляет контейнер всех зависимостей, необходимых для работы клиента.
// Это упрощает передачу сервисов в TUI и другие слои приложения, делает код более модульным
// и облегчает тестирование.
//
// Поля:
//   - AuthManager: отвечает за регистрацию, вход и управление токеном доступа пользователя.
//   - ConnManager: управляет подключением к gRPC-серверу.
//   - Logger: структурированный логгер для записи отладочной и диагностической информации.
type AppServices struct {
	AuthManager       auth.AuthManagerIface
	CredentialManager credential.CredentialManagerIface
	CryptoKeyManager  cryptokey.CryptoKeyManagerIface
	ConnManager       connection.ConnManager
	Logger            *zap.Logger
	// Будущие зависимости:
	// DataService *data.Service

	closeOnce sync.Once
}

// NewAppServices создаёт контейнер зависимостей клиента.
// logDir — директория для логов, cfg — конфигурация клиента.
func NewAppServices(cfg *config.ClientConfig, logDir string) (*AppServices, error) {
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	if err := logger.InitializeWithTimestamp(cfg.LogLevel, logDir); err != nil {
		return nil, fmt.Errorf("logger initialize failed: %w", err)
	}

	log := logger.Log

	connConfig := &connection.Config{
		ServerAddress:  cfg.ServerAddress,
		UseTLS:         cfg.UseTLS,
		TLSSkipVerify:  cfg.TLSSkipVerify,
		CACertPath:     cfg.CACertPath,
		ConnectTimeout: cfg.Timeout,
	}

	tokenStore := storage.NewFileTokenStorage(".token")

	keyFilePath, err := crypto.DefaultKeyFilePath()
	if err != nil {
		// обработка ошибки
	}

	cryptoStore := storage.NewFileCryptoKeyStorage(keyFilePath)
	cryptoKeyManager := cryptokey.NewCryptoKeyManager(cryptoStore, log)
	authManager := auth.NewAuthManager(tokenStore, log)

	// Создаем CredentialManager, передав logger
	credentialManager := credential.NewCredentialManager(log)

	connManager := connection.New(connConfig, log, authManager)

	return &AppServices{
		AuthManager:       authManager,
		CredentialManager: credentialManager,
		CryptoKeyManager:  cryptoKeyManager,
		ConnManager:       connManager,
		Logger:            log,
	}, nil
}

// getGRPCConn возвращает активное gRPC соединение, обеспечивая его создание и восстановление при необходимости.
//
// ctx — контекст запроса, используется для контроля таймаута и отмены операции.
//
// Возвращает готовое соединение GrpcConn или ошибку при неудаче подключения.
func (s *AppServices) getGRPCConn(ctx context.Context) (connection.GrpcConn, error) {
	conn, err := s.ConnManager.Connect(ctx)
	if err != nil {
		s.Logger.Error("Failed to connect gRPC", zap.Error(err))
		return nil, err
	}
	return conn, nil
}

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

// ensureCredentialClient гарантирует создание gRPC клиента для Credential сервиса и установку его в CredentialManager.
//
// ctx — контекст запроса.
//
// Возвращает ошибку при сбое подключения.
func (s *AppServices) ensureCredentialClient(ctx context.Context) error {
	conn, err := s.getGRPCConn(ctx)
	if err != nil {
		return err
	}

	client := proto.NewCredentialServiceClient(conn)
	s.CredentialManager.SetClient(client)
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

// CreateCredential создаёт новую учётную запись (credential) на сервере.
//
// ctx — контекст запроса.
// cred — данные учётных данных для создания.
//
// Возвращает ошибку при сбое RPC вызова.
func (s *AppServices) CreateCredential(ctx context.Context, cred *model.Credential) error {
	err := s.ensureCredentialClient(ctx)
	if err != nil {
		return err
	}
	return s.CredentialManager.CreateCredential(ctx, cred)
}

// GetCredentialByID получает учётные данные по их уникальному идентификатору.
//
// ctx — контекст запроса.
// id — идентификатор учётных данных.
//
// Возвращает найденные учётные данные или ошибку.
func (s *AppServices) GetCredentialByID(ctx context.Context, id string) (*model.Credential, error) {
	err := s.ensureCredentialClient(ctx)
	if err != nil {
		return nil, err
	}
	return s.CredentialManager.GetCredentialByID(ctx, id)
}

// GetCredentialsByUserID возвращает список учётных данных для заданного пользователя.
//
// ctx — контекст запроса.
// userID — идентификатор пользователя.
//
// Возвращает срез учётных данных или ошибку.
func (s *AppServices) GetCredentialsByUserID(ctx context.Context, userID string) ([]model.Credential, error) {
	err := s.ensureCredentialClient(ctx)
	if err != nil {
		return nil, err
	}
	return s.CredentialManager.GetCredentialsByUserID(ctx, userID)
}

// UpdateCredential обновляет существующую учётную запись.
//
// ctx — контекст запроса.
// cred — обновлённые данные учётных данных.
//
// Возвращает ошибку при сбое RPC вызова.
func (s *AppServices) UpdateCredential(ctx context.Context, cred *model.Credential) error {
	err := s.ensureCredentialClient(ctx)
	if err != nil {
		return err
	}
	return s.CredentialManager.UpdateCredential(ctx, cred)
}

// DeleteCredential удаляет учётные данные по идентификатору.
//
// ctx — контекст запроса.
// id — идентификатор учётных данных для удаления.
//
// Возвращает ошибку при сбое RPC вызова.
func (s *AppServices) DeleteCredential(ctx context.Context, id string) error {
	err := s.ensureCredentialClient(ctx)
	if err != nil {
		return err
	}
	return s.CredentialManager.DeleteCredential(ctx, id)
}

// Close корректно освобождает ресурсы, в частности закрывает gRPC соединение.
//
// Возвращает ошибку, если закрытие прошло с проблемами.
func (s *AppServices) Close() error {
	var err error
	s.closeOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			err = s.ConnManager.Close()
			close(done)
		}()

		select {
		case <-done:
			// закрытие успешно
		case <-ctx.Done():
			s.Logger.Warn("Превышено время ожидания закрытия ресурсов")
		}
	})
	return err
}
