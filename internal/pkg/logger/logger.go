// Package logger предоставляет централизованную систему логирования для приложения
// на основе zap.Logger. Реализует паттерн синглтона для глобального доступа к логеру.
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Log - глобальный экземпляр логера, инициализированный no-op логером по умолчанию.
// No-op логер не производит никакого вывода и не аллоцирует ресурсы.
var Log *zap.Logger = zap.NewNop()

// ljLogger — держим ссылку на lumberjack.Logger для Close()
var ljLogger *lumberjack.Logger

//var logFile *os.File // держим дескриптор, чтобы можно было закрыть

// Initialize настраивает глобальный логгер zap с указанным уровнем логирования и опциональной записью в файл с ротацией.
//
// Параметры:
//   - level: строка, определяющая уровень логирования (debug, info, warn, error, dpanic, panic, fatal).
//   - logFilePath: путь к лог-файлу. Если пустая строка, логирование ведётся в stdout.
//
// Возвращает:
//   - error: если уровень логирования некорректен или при создании логгера произошла ошибка.
//
// Особенности:
//   - При указании пути к файлу используется lumberjack.Logger для ротации логов (максимальный размер файла, количество бэкапов, максимальный возраст).
//   - При пустом пути создаётся стандартный production-логгер zap с выводом в stdout.
//
// Пример использования:
//
//	err := logger.Initialize("debug", "")
//	if err != nil {
//	    // обработка ошибки инициализации
//	}
//	logger.Log.Info("Логгер успешно инициализирован")
func Initialize(level string, logFilePath string) error {
	// Преобразование строкового уровня в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	// Если путь к файлу пустой — используем обычный Build() путь (stdout/stderr).
	if logFilePath == "" {
		// Конфигурация логера в production-стиле (JSON-формат, stacktrace для ошибок)
		cfg := zap.NewProductionConfig()
		// Установка уровня логирования
		cfg.Level = lvl
		// Создание логера на основе конфигурации
		zl, err := cfg.Build()
		if err != nil {
			return err
		}
		// Замена глобального логера
		Log = zl
		return nil
	}

	/*
		// Открываем файл сами, чтобы позже иметь возможность его закрыть.
		// Создаём с флагами: create, append, write-only
		if err := os.MkdirAll(filepath.Dir(logFilePath), 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return err
		}

		// Сохраняем дескриптор глобально, чтобы Close() мог его закрыть.
		logFile = f

		// Создаём zap core, encoder и logger вручную, чтобы записывать в наш файл.
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "ts"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(f),
			lvl,
		)
	*/
	// Используем lumberjack для ротации
	ljLogger = &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	writer := zapcore.AddSync(ljLogger)

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.LevelKey = "level"
	encCfg.CallerKey = "caller"

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		writer,
		lvl,
	)

	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
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

	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	filename := filepath.Join(logDir, time.Now().Format("client_2006-01-02_15-04")+".log")
	return Initialize(level, filename)
}

// Close корректно завершает работу логгера, освобождая ресурсы.
// Всегда следует вызывать перед завершением процесса или завершением теста,
// чтобы избежать блокировки файлов логов и гарантировать, что все буферы
// были сброшены и закрыты.
//
// Возвращает ошибку, если при закрытии логгера возникли проблемы.
func Close() error {
	// Сначала сбрасываем буферы внутри zap
	if Log != nil {
		_ = Log.Sync()
	}
	if ljLogger != nil {
		err := ljLogger.Close()
		ljLogger = nil
		if err != nil {
			return fmt.Errorf("failed to close lumberjack logger: %w", err)
		}
	}
	return nil
}
