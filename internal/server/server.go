package server

import (
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtutils"
	"github.com/ryabkov82/gophkeeper/internal/server/config"
	"github.com/ryabkov82/gophkeeper/internal/server/grpc"
	"github.com/ryabkov82/gophkeeper/internal/server/service"
	"github.com/ryabkov82/gophkeeper/internal/server/storage"
	"github.com/ryabkov82/gophkeeper/internal/server/storage/postgres"
	"go.uber.org/zap"
)

func StartServer(log *zap.Logger, cfg *config.Config) {
	// 1. Инициализация БД
	db, err := postgres.Init(cfg.DBConnect)
	if err != nil {
		log.Fatal("Failed to initialize storage", zap.Error(err))
	}

	// 2. Слои: repository -> services
	repos := storage.NewRepositories(db)
	jwtManager := jwtutils.New(cfg.JwtKey, 24*60*60)
	services := service.NewServices(repos, jwtManager)

	// 3. Запуск gRPC сервера с набором сервисов
	if err := grpc.StartGRPCServer(log, cfg, services); err != nil {
		log.Fatal("gRPC server failed", zap.Error(err))
	}
}
