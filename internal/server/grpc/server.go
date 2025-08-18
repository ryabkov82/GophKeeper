package grpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtutils"
	api "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/ryabkov82/gophkeeper/internal/server/config"
	"github.com/ryabkov82/gophkeeper/internal/server/grpc/handlers"
	"github.com/ryabkov82/gophkeeper/internal/server/grpc/interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// NewGRPCServer создаёт и настраивает gRPC-сервер с зарегистрированными хендлерами
func NewGRPCServer(cfg *config.Config, logger *zap.Logger, serviceFactory service.ServiceFactory) (*grpc.Server, error) {
	var opts []grpc.ServerOption

	jwtManager := jwtutils.New(cfg.JwtKey, 24*60*60)
	// Добавляем интерцепторы
	opts = append(opts,
		grpc.ChainUnaryInterceptor(
			interceptors.LoggingInterceptor(logger),
			interceptors.UnaryAuthInterceptor(jwtManager, logger),
			// можно добавить другие: AuthInterceptor, RecoveryInterceptor и т.п.
		),
	)

	if cfg.EnableTLS {
		creds, err := credentials.NewServerTLSFromFile(cfg.SSLCertFile, cfg.SSLKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certs: %w", err)
		}
		logger.Info("TLS is enabled for gRPC server")
		opts = append(opts, grpc.Creds(creds))
	} else {
		logger.Info("TLS is disabled for gRPC server")
	}

	s := grpc.NewServer(opts...)

	// Регистрируем Auth хендлер
	authHandler := handlers.NewAuthHandler(serviceFactory.Auth(), logger)
	api.RegisterAuthServiceServer(s, authHandler)

	// Регистрируем Credential хендлер
	credHandler := handlers.NewCredentialHandler(serviceFactory.Credential(), logger)
	api.RegisterCredentialServiceServer(s, credHandler)

	// Регистрируем BankCard хендлер
	bankcardHandler := handlers.NewBankCardHandler(serviceFactory.BankCard(), logger)
	api.RegisterBankCardServiceServer(s, bankcardHandler)

	// Регистрируем TextData хендлер
	textDataHandler := handlers.NewTextDataHandler(serviceFactory.TextData(), logger)
	api.RegisterTextDataServiceServer(s, textDataHandler)

	// Регистрируем TextData хендлер
	binaryDataHandler := handlers.NewBinaryDataHandler(serviceFactory.BinaryData(), logger)
	api.RegisterBinaryDataServiceServer(s, binaryDataHandler)

	return s, nil
}

// функция запуска gRPC сервера с graceful shutdown
func ServeGRPC(
	s *grpc.Server,
	lis net.Listener,
	signals <-chan os.Signal,
	logger *zap.Logger,
	svcFactory service.ServiceFactory,
) {
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Fatal("gRPC serve error", zap.Error(err))
		}
	}()
	logger.Info("gRPC server started", zap.String("addr", lis.Addr().String()))

	<-signals // ждём сигнала завершения

	logger.Info("Shutting down gRPC server...")

	// Закрываем бинарное хранилище перед остановкой сервера
	svcFactory.BinaryData().Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Server stopped gracefully")
	case <-ctx.Done():
		logger.Warn("Graceful shutdown timed out")
		s.Stop()
	}
}

// StartGRPCServer запускает gRPC-сервер с учетом конфигурации.
//
// Функция выполняет следующие действия:
//   - Открывает TCP-порт, указанный в cfg.GRPCServerAddr;
//   - При включенном cfg.EnableTLS использует TLS-сертификаты из cfg.SSLCertFile и cfg.SSLKeyFile;
//   - Регистрирует обработчики gRPC-сервисов (в частности, AuthService);
//   - Запускает сервер в отдельной горутине и логирует адрес запуска;
//   - Обрабатывает системные сигналы завершения (SIGINT, SIGTERM);
//   - При завершении инициирует корректное отключение через GracefulStop()
//     с таймаутом 5 секунд. В случае превышения таймаута — выполняется
//     принудительная остановка через Stop();
//
// Параметры:
//   - log: логгер zap.Logger для вывода ошибок и состояния сервера.
//   - cfg: конфигурация сервера, включая адрес, TLS-настройки и пути к сертификатам.
//   - services: контейнер зарегистрированных бизнес-сервисов.
//
// Возвращает ошибку при невозможности запуска или в процессе graceful shutdown.
func StartGRPCServer(log *zap.Logger, cfg *config.Config, serviceFactory service.ServiceFactory) error {
	lis, err := net.Listen("tcp", cfg.GRPCServerAddr)
	if err != nil {
		log.Error("failed to listen", zap.Error(err))
		return err
	}

	srv, err := NewGRPCServer(cfg, log, serviceFactory)
	if err != nil {
		log.Error("failed to create gRPC server", zap.Error(err))
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ServeGRPC(srv, lis, sigChan, log, serviceFactory)
	return nil
}
