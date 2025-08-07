package main

import (
	"context"
	"fmt"
	"log"
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

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application exited with error: %v", err)
	}
}

// run выполняет основную инициализацию клиентского приложения:
// - загружает конфигурацию
// - инициализирует логгер
// - создаёт менеджер gRPC соединений
// - создает менеджер авторизации
// - запускает TUI с обработкой системных сигналов
//
// Возвращает ошибку, если какая-либо из инициализаций или запуск TUI завершился с ошибкой.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logDir := "logs"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	if err := logger.InitializeWithTimestamp(cfg.LogLevel, logDir); err != nil {
		return fmt.Errorf("logger initialize failed: %w", err)
	}
	defer logger.Log.Sync()

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
	return tui.Run(ctx, services)
}
