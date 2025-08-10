package app

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/config"
	"github.com/ryabkov82/gophkeeper/internal/client/connection"
	"github.com/ryabkov82/gophkeeper/internal/client/service/auth"
	"github.com/ryabkov82/gophkeeper/internal/pkg/logger"
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
	AuthManager auth.AuthManagerIface
	ConnManager connection.ConnManager
	Logger      *zap.Logger
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

	connManager := connection.New(connConfig, log)
	tokenStore := auth.NewFileTokenStorage(".token")
	authManager := auth.NewAuthManager(connManager, tokenStore, log)

	return &AppServices{
		AuthManager: authManager,
		ConnManager: connManager,
		Logger:      log,
	}, nil
}

// Close освобождает ресурсы
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
