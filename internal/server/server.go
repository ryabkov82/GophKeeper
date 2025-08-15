package server

import (
	"time"

	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtutils"
	"github.com/ryabkov82/gophkeeper/internal/server/config"
	"github.com/ryabkov82/gophkeeper/internal/server/grpc"
	"github.com/ryabkov82/gophkeeper/internal/server/service"
	"github.com/ryabkov82/gophkeeper/internal/server/storage"
	"github.com/ryabkov82/gophkeeper/internal/server/storage/postgres"
	"go.uber.org/zap"
)

// StartServer выполняет полную инициализацию и запуск gRPC-сервера приложения.
//
// Последовательно выполняются следующие шаги:
//  1. Инициализация хранилища данных (PostgreSQL) через Init;
//  2. Создание слоёв репозиториев и сервисов, включая JWT-менеджер;
//  3. Запуск gRPC-сервера с зарегистрированными сервисами.
//
// В случае ошибки на любом этапе, функция логирует критическую ошибку
// и завершает выполнение приложения.
//
// Параметры:
//   - log: логгер для вывода состояния и ошибок (zap.Logger);
//   - cfg: структура конфигурации, содержащая параметры подключения к БД,
//     TLS-настройки, JWT-ключ и другие параметры.
//
// Ошибки не возвращаются, функция завершает приложение через log.Fatal в случае сбоя.
func StartServer(log *zap.Logger, cfg *config.Config) {
	// 1. Инициализация БД
	db, err := postgres.Init(cfg.DBConnect)
	if err != nil {
		log.Fatal("Failed to initialize storage", zap.Error(err))
	}

	// 2. Слои: repository -> services
	storageFactory, err := storage.NewStorageFactory("postgres", db)
	if err != nil {
		log.Fatal("Failed to initialize storage", zap.Error(err))
	}

	jwtManager := jwtutils.New(cfg.JwtKey, 24*time.Hour)
	serviceFactory := service.NewServiceFactory(storageFactory, jwtManager)

	// 3. Запуск gRPC сервера с набором сервисов
	if err := grpc.StartGRPCServer(log, cfg, serviceFactory); err != nil {
		log.Fatal("gRPC server failed", zap.Error(err))
	}
}
