package app_test

import (
	"os"
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/config"
	"github.com/ryabkov82/gophkeeper/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewAppServices_Success(t *testing.T) {
	tempLogDir := t.TempDir()

	cfg := &config.ClientConfig{
		ServerAddress: "localhost:50051",
		UseTLS:        false,
		LogLevel:      "debug",
		Timeout:       2 * time.Second,
	}

	services, err := app.NewAppServices(cfg, tempLogDir)
	assert.NoError(t, err, "NewAppServices не должен возвращать ошибку")
	assert.NotNil(t, services, "services должен быть создан")

	// Проверяем, что зависимости созданы
	assert.NotNil(t, services.Logger, "Logger должен быть инициализирован")
	assert.NotNil(t, services.AuthManager, "AuthManager должен быть инициализирован")
	assert.NotNil(t, services.ConnManager, "ConnManager должен быть инициализирован")

	// Проверяем, что каталог логов существует
	_, err = os.Stat(tempLogDir)
	assert.NoError(t, err, "Каталог логов должен существовать")

	// Закрываем ресурсы
	services.Close()
	logger.Close()
}

func TestNewAppServices_MkdirFail(t *testing.T) {
	cfg := &config.ClientConfig{
		LogLevel: "debug",
	}

	// Некорректный путь
	badDir := string([]byte{0})

	services, err := app.NewAppServices(cfg, badDir)
	assert.Error(t, err, "ожидалась ошибка при создании директории")
	assert.Nil(t, services, "services должен быть nil при ошибке")
}

func TestNewAppServices_LoggerInitFail(t *testing.T) {
	// Чтобы сломать инициализацию логгера, можно передать некорректный уровень логирования
	cfg := &config.ClientConfig{
		LogLevel: "invalid_level", // заведомо некорректно
	}

	tempLogDir := t.TempDir()

	services, err := app.NewAppServices(cfg, tempLogDir)
	assert.Error(t, err, "ожидалась ошибка при неправильном уровне логирования")
	assert.Nil(t, services, "services должен быть nil при ошибке")
}
