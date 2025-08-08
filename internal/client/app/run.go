package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ryabkov82/gophkeeper/internal/client/auth"
	"github.com/ryabkov82/gophkeeper/internal/client/config"
	"github.com/ryabkov82/gophkeeper/internal/client/connection"
	"github.com/ryabkov82/gophkeeper/internal/client/tui"
	"github.com/ryabkov82/gophkeeper/internal/pkg/logger"
	"go.uber.org/zap"
)

// runWithServices выполняет инициализацию и запуск TUI,
// принимает готовую конфигурацию, директорию для логов и функцию запуска TUI.
func RunWithServices(
	cfg *config.ClientConfig,
	logDir string,
	runTUI func(ctx context.Context, services *tui.AppServices) error,
) error {

	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	if err := logger.InitializeWithTimestamp(cfg.LogLevel, logDir); err != nil {
		return fmt.Errorf("logger initialize failed: %w", err)
	}
	defer logger.Close()

	log := logger.Log
	log.Info("Конфигурация загружена", zap.Any("config", cfg))

	connConfig := &connection.Config{
		ServerAddress:  cfg.ServerAddress,
		UseTLS:         cfg.UseTLS,
		TLSSkipVerify:  cfg.TLSSkipVerify,
		CACertPath:     cfg.CACertPath,
		ConnectTimeout: cfg.Timeout,
	}

	log.Debug("Создание менеджера соединений", zap.Any("connConfig", connConfig))

	connManager := connection.New(connConfig, log)
	tokenStore := auth.NewFileTokenStorage(".token")
	authManager := auth.NewAuthManager(connManager, tokenStore, log)

	// Контекст обработки сигналов
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	services := &tui.AppServices{
		AuthManager: authManager,
		ConnManager: connManager,
		Logger:      log,
	}

	log.Info("Запуск TUI...")
	return runTUI(ctx, services)
}
