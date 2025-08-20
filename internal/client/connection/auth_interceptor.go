package connection

import (
	"context"
	"strings"

	"github.com/ryabkov82/gophkeeper/internal/client/service/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

// NewAuthPerRPCCredentials создаёт PerRPCCredentials, которые
// добавляют Authorization: Bearer <token> во все RPC (unary и stream),
// кроме указанных в exclude (по имени метода: "Login", "Register").
func NewAuthPerRPCCredentials(
	authManager auth.AuthManagerIface,
	logger *zap.Logger,
	requireTLS bool, // true в проде; false можно для dev/plaintext
) credentials.PerRPCCredentials {

	ex := map[string]struct{}{
		"Login":    {},
		"Register": {},
	}

	return &authPerRPCCreds{
		am:         authManager,
		logger:     logger,
		exclude:    ex,
		requireTLS: requireTLS,
	}
}

type authPerRPCCreds struct {
	am         auth.AuthManagerIface
	logger     *zap.Logger
	exclude    map[string]struct{}
	requireTLS bool
}

// GetRequestMetadata вызывается на КАЖДЫЙ RPC/стрим.
// uri обычно содержит полный путь метода вида "/pkg.Service/Method".
func (c *authPerRPCCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	fullMethod := ""
	if len(uri) > 0 {
		fullMethod = uri[0]
	}
	if c.shouldSkip(fullMethod) {
		c.logger.Debug("Skipping auth token for excluded method", zap.String("method", fullMethod))
		return nil, nil // ничего не добавляем
	}

	token := c.am.GetToken()
	if token == "" {
		c.logger.Debug("No token; sending unauthenticated request", zap.String("method", fullMethod))
		return nil, nil
	}

	c.logger.Debug("Added authorization token", zap.String("method", fullMethod))
	return map[string]string{"authorization": "Bearer " + token}, nil
}

// RequireTransportSecurity сообщает, требуются ли TLS-соединения
// для отправки этих кредов.
func (c *authPerRPCCreds) RequireTransportSecurity() bool {
	return c.requireTLS
}

func (c *authPerRPCCreds) shouldSkip(fullMethod string) bool {
	// fullMethod: "/pkg.service/Method" -> берём "Method"
	if fullMethod == "" {
		return false
	}
	parts := strings.Split(fullMethod, "/")
	method := parts[len(parts)-1]
	_, skip := c.exclude[method]
	return skip
}
