package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/config"
	"github.com/ryabkov82/gophkeeper/internal/client/tuiiface"
	"github.com/stretchr/testify/assert"
)

func TestRunWithServices(t *testing.T) {

	tempLogDir := t.TempDir()
	cfg := &config.ClientConfig{
		ServerAddress: "localhost:50051",
		UseTLS:        false,
		LogLevel:      "debug",
		Timeout:       2 * time.Second,
		LogDirPath:    tempLogDir,
	}

	var runTUICalled bool

	mockRunTUI := func(ctx context.Context, services *app.AppServices, progFactory tuiiface.ProgramFactory) error {
		runTUICalled = true

		assert.NotNil(t, ctx)
		assert.NotNil(t, services)
		// progFactory может быть nil, проверять не обязательно
		return nil
	}

	err := app.RunWithServices(cfg, mockRunTUI)
	assert.NoError(t, err)
	assert.True(t, runTUICalled)
}

func TestRunWithServices_MkdirFail(t *testing.T) {

	badDir := string([]byte{0})
	cfg := &config.ClientConfig{LogDirPath: badDir}

	mockRunTUI := func(ctx context.Context, services *app.AppServices, progFactory tuiiface.ProgramFactory) error {
		return nil
	}

	err := app.RunWithServices(cfg, mockRunTUI)
	assert.Error(t, err)
}

func TestRunWithServices_RunTUIFail(t *testing.T) {

	tempLogDir := t.TempDir()
	cfg := &config.ClientConfig{
		ServerAddress: "localhost:50051",
		UseTLS:        false,
		LogLevel:      "debug",
		Timeout:       2 * time.Second,
		LogDirPath:    tempLogDir,
	}

	mockRunTUI := func(ctx context.Context, services *app.AppServices, progFactory tuiiface.ProgramFactory) error {
		return errors.New("TUI error")
	}

	err := app.RunWithServices(cfg, mockRunTUI)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TUI error")
}
