package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ryabkov82/gophkeeper/internal/client/config"
	"github.com/ryabkov82/gophkeeper/internal/client/tuiiface"
	"github.com/ryabkov82/gophkeeper/internal/pkg/logger"
	"go.uber.org/zap"
)

// RunWithServices выполняет инициализацию всех необходимых сервисов и
// запускает TUI-приложение с учётом обработки системных сигналов.
//
// Аргументы:
//   - cfg: готовая конфигурация клиента, содержащая настройки подключения,
//     логирования и другие параметры.
//   - logDir: путь к директории, где будут сохраняться логи приложения.
//   - runTUI: функция запуска TUI, принимающая контекст, контейнер сервисов,
//     фабрику создания TUI-программы и канал системных сигналов.
//
// Логика:
//  1. Создаёт контейнер сервисов (AppServices) с инициализацией логгера,
//     менеджера соединений и других зависимостей.
//  2. Создаёт контекст, который будет отменён при получении сигнала
//     завершения (SIGINT, SIGTERM).
//  3. Создаёт фабрику progFactory для создания TUI-программы на базе bubbletea,
//     с включённым альтернативным экраном.
//  4. Вызывает runTUI, передавая контекст, сервисы, фабрику программы.
//  5. Обеспечивает корректное освобождение ресурсов (закрытие сервисов) при завершении.
//
// Возвращает:
//   - ошибку, возникшую при инициализации сервисов или при выполнении runTUI.
func RunWithServices(
	cfg *config.ClientConfig,
	logDir string,
	runTUI func(ctx context.Context, services *AppServices, progFactory tuiiface.ProgramFactory) error,
) error {
	services, err := NewAppServices(cfg, logDir)
	if err != nil {
		return err
	}
	defer services.Close()
	defer logger.Close()

	// Контекст с автоматической отменой при получении SIGINT или SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	services.Logger.Info("Конфигурация загружена", zap.Any("config", cfg))
	services.Logger.Info("Запуск TUI...")

	return runTUI(ctx, services, nil)
}
