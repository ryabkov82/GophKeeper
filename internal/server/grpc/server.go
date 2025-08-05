// internal/server/grpc/server.go
package grpc

import (
	"net"

	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	api "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/ryabkov82/gophkeeper/internal/server/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func StartGRPCServer(log *zap.Logger, cfg *config.Config, services *service.Services) error {
	lis, err := net.Listen("tcp", cfg.GRPCServerAddr)
	if err != nil {
		log.Error("failed to listen", zap.Error(err))
		return err
	}

	s := grpc.NewServer()

	authHandler := NewAuthHandler(services.Auth)
	api.RegisterAuthServiceServer(s, authHandler)

	// Other handlers: proto.RegisterDataServiceServer(s, handler.NewDataHandler(services.Data))

	log.Info("gRPC server started", zap.String("addr", cfg.GRPCServerAddr))
	return s.Serve(lis)
}
