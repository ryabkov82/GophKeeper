// Package logger предоставляет централизованную систему логирования для приложения
// на основе zap.Logger. Реализует паттерн синглтона для глобального доступа к логеру.
package logger

import (
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// Log - глобальный экземпляр логера, инициализированный no-op логером по умолчанию.
// No-op логер не производит никакого вывода и не аллоцирует ресурсы.
var Log *zap.Logger = zap.NewNop()

// Initialize настраивает глобальный логер с указанным уровнем логирования.
//
// Параметры:
//   - level: строка, определяющая уровень логирования (debug, info, warn, error, dpanic, panic, fatal)
//   - logFilePath — путь к лог-файлу. Если пусто, используется stdout.
//
// Возвращает:
//   - error: ошибка, если передан некорректный уровень логирования или возникла проблема при создании логера
//
// Пример использования:
//
//	err := logger.Initialize("debug")
//	if err != nil {
//	    // обработка ошибки инициализации
//	}
//	logger.Log.Info("Логер успешно инициализирован")
func Initialize(level string, logFilePath string) error {
	// Преобразование строкового уровня в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	// Конфигурация логера в production-стиле (JSON-формат, stacktrace для ошибок)
	cfg := zap.NewProductionConfig()

	// Установка уровня логирования
	cfg.Level = lvl

	if logFilePath != "" {
		// Лог в файл
		cfg.OutputPaths = []string{logFilePath}
		cfg.ErrorOutputPaths = []string{logFilePath}
	} else {
		// Лог в stdout
		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stderr"}
	}

	// Создание логера на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	// Замена глобального логера
	Log = zl
	return nil
}

// InitializeWithTimestamp инициализирует глобальный логгер zap и сохраняет логи в файл с временной меткой.
//
// Лог-файл будет создан в указанной директории с именем вида:
//
//	client_YYYY-MM-DD_HH-MM.log
//
// Параметры:
//   - level: строка, определяющая уровень логирования (например, "debug", "info", "warn", "error").
//   - logDir: путь к директории, где будет создан лог-файл.
//
// Возвращает:
//   - error: ошибку при разборе уровня логирования, создании каталога или инициализации логгера.
//
// Пример использования:
//
//	err := logger.InitializeWithTimestamp("debug", "logs")
//	if err != nil {
//	    log.Fatalf("Не удалось инициализировать логгер: %v", err)
//	}
//	logger.Log.Info("Логгер инициализирован")
func InitializeWithTimestamp(level, logDir string) error {
	filename := filepath.Join(logDir, time.Now().Format("client_2006-01-02_15-04")+".log")
	return Initialize(level, filename)
}
