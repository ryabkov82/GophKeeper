package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/config"
	"github.com/ryabkov82/gophkeeper/internal/client/tui"
	"github.com/stretchr/testify/assert"
)

func TestRunWithServices(t *testing.T) {
	// Мок конфигурации
	cfg := &config.ClientConfig{
		ServerAddress: "localhost:50051",
		UseTLS:        false,
		LogLevel:      "debug",
		Timeout:       2 * time.Second,
	}

	tempLogDir := t.TempDir() // создаем временную директорию для логов

	// Переменная для проверки вызова runTUI
	var runTUICalled bool

	mockRunTUI := func(ctx context.Context, services *tui.AppServices) error {
		runTUICalled = true
		return nil // имитируем успешный запуск
	}

	err := app.RunWithServices(cfg, tempLogDir, mockRunTUI)
	assert.NoError(t, err, "runWithServices должен завершиться без ошибки")
	assert.True(t, runTUICalled, "runTUI должен быть вызван")
}

func TestRunWithServices_MkdirFail(t *testing.T) {
	cfg := &config.ClientConfig{}
	// Указываем директорию, которая не может быть создана (например, корневая Windows)
	badDir := string([]byte{0})

	mockRunTUI := func(ctx context.Context, services *tui.AppServices) error {
		return nil
	}

	err := app.RunWithServices(cfg, badDir, mockRunTUI)
	assert.Error(t, err)
}

func TestRunWithServices_RunTUIFail(t *testing.T) {
	cfg := &config.ClientConfig{
		ServerAddress: "localhost:50051",
		UseTLS:        false,
		LogLevel:      "debug",
		Timeout:       2 * time.Second,
	}

	tempLogDir := t.TempDir()

	mockRunTUI := func(ctx context.Context, services *tui.AppServices) error {
		return errors.New("TUI error")
	}

	err := app.RunWithServices(cfg, tempLogDir, mockRunTUI)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TUI error")
}
