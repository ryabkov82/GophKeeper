package connection

import (
	"context"
	"strings"

	"github.com/ryabkov82/gophkeeper/internal/client/service/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthUnaryInterceptor возвращает UnaryClientInterceptor,
// который добавляет токен авторизации в метаданные исходящего контекста,
// за исключением методов, перечисленных в exclude.
func AuthUnaryInterceptor(authManager auth.AuthManagerIface, logger *zap.Logger) grpc.UnaryClientInterceptor {
	exclude := map[string]struct{}{
		"Login":    {},
		"Register": {},
	}

	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		parts := strings.Split(method, "/")
		methodName := ""
		if len(parts) == 3 {
			methodName = parts[2]
		}

		if _, skip := exclude[methodName]; !skip {
			token := authManager.GetToken()
			if token != "" {
				md, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					md = metadata.New(nil)
				} else {
					md = md.Copy()
				}
				md.Set("authorization", "Bearer "+token)
				ctx = metadata.NewOutgoingContext(ctx, md)
				logger.Debug("Added authorization token to gRPC request", zap.String("method", method))
			} else {
				logger.Debug("No token found, sending unauthenticated gRPC request", zap.String("method", method))
			}
		} else {
			logger.Debug("Skipping auth token for excluded method", zap.String("method", method))
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
