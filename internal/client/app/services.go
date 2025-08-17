package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/config"
	"github.com/ryabkov82/gophkeeper/internal/client/connection"
	"github.com/ryabkov82/gophkeeper/internal/client/service/auth"
	"github.com/ryabkov82/gophkeeper/internal/client/service/bankcard"
	"github.com/ryabkov82/gophkeeper/internal/client/service/credential"
	"github.com/ryabkov82/gophkeeper/internal/client/service/cryptokey"
	"github.com/ryabkov82/gophkeeper/internal/client/service/textdata"
	"github.com/ryabkov82/gophkeeper/internal/client/storage"
	"github.com/ryabkov82/gophkeeper/internal/pkg/logger"
	"go.uber.org/zap"
)

// AppServices представляет контейнер всех основных сервисов и зависимостей,
// необходимых для работы клиента.
//
// Такая структура облегчает передачу зависимостей между слоями приложения (например, в TUI),
// повышает модульность кода и упрощает тестирование.
//
// Основные поля:
//   - AuthManager: управление регистрацией, аутентификацией и токенами доступа пользователей.
//   - CredentialManager: управление учётными данными (создание, получение, обновление, удаление).
//   - BankCardManager: управление банковскими картами (создание, получение, обновление, удаление).
//   - CryptoKeyManager: генерация, хранение и загрузка криптографических ключей для шифрования.
//   - ConnManager: управление gRPC подключениями к серверу.
//   - Logger: структурированный логгер для записи отладочной, диагностической и системной информации.
//
// Для корректного закрытия ресурсов (например, gRPC соединений) используется sync.Once.
type AppServices struct {
	AuthManager       auth.AuthManagerIface
	CredentialManager credential.CredentialManagerIface
	BankCardManager   bankcard.BankCardManagerIface
	TextDataManager   textdata.TextDataManagerIface
	CryptoKeyManager  cryptokey.CryptoKeyManagerIface
	ConnManager       connection.ConnManager
	Logger            *zap.Logger

	closeOnce sync.Once
}

// NewAppServices создаёт контейнер зависимостей клиента.
// cfg — конфигурация клиента.
func NewAppServices(cfg *config.ClientConfig) (*AppServices, error) {

	if err := logger.InitializeWithTimestamp(cfg.LogLevel, cfg.LogDirPath); err != nil {
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

	tokenStore := storage.NewFileTokenStorage(cfg.TokenFilePath)

	cryptoStore := storage.NewFileCryptoKeyStorage(cfg.KeyFilePath)
	cryptoKeyManager := cryptokey.NewCryptoKeyManager(cryptoStore, log)
	authManager := auth.NewAuthManager(tokenStore, log)

	// Создаем CredentialManager, передав logger
	credentialManager := credential.NewCredentialManager(log)

	// Создаем BankCardManager, передав logger
	bankcardManager := bankcard.NewBankCardManager(log)

	// Создаем BankCardManager, передав logger
	textdataManager := textdata.NewTextDataManager(log)

	connManager := connection.New(connConfig, log, authManager)

	return &AppServices{
		AuthManager:       authManager,
		CredentialManager: credentialManager,
		BankCardManager:   bankcardManager,
		TextDataManager:   textdataManager,
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
