package interceptors

import (
	"context"
	"strings"

	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtauth"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtutils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryAuthInterceptor возвращает gRPC UnaryServerInterceptor, который выполняет
// аутентификацию запроса на основе JWT токена, переданного в метаданных запроса.
//
// Интерцептор:
//   - Извлекает метаданные из входящего контекста.
//   - Проверяет наличие и формат заголовка "authorization" с Bearer токеном.
//   - Парсит и валидирует JWT токен с помощью TokenManager.
//   - Извлекает userID из claims токена ("sub").
//   - Если аутентификация успешна, добавляет userID в контекст запроса и передаёт управление дальше.
//   - В случае ошибок возвращает ошибку с кодом Unauthenticated.
//
// Параметры:
//   - tm: менеджер токенов для проверки и парсинга JWT.
//
// Возвращаемое значение:
//   - grpc.UnaryServerInterceptor — функция интерцептора для gRPC.
//
// Пример использования:
//
//	grpcServer := grpc.NewServer(
//	    grpc.UnaryInterceptor(UnaryAuthInterceptor(tokenManager)),
//	)
func UnaryAuthInterceptor(tm *jwtutils.TokenManager, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, err := authenticateCtx(ctx, tm, logger, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

type ctxStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context возвращает контекст, в который был добавлен идентификатор пользователя.
func (s *ctxStream) Context() context.Context { return s.ctx }

// StreamAuthInterceptor возвращает gRPC StreamServerInterceptor, выполняющий
// проверку JWT-токена для потоковых RPC-запросов.
func StreamAuthInterceptor(tm *jwtutils.TokenManager, logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, err := authenticateCtx(ss.Context(), tm, logger, info.FullMethod)
		if err != nil {
			return err
		}
		// Подменяем контекст в stream, чтобы handler видел userID в stream.Context()
		return handler(srv, &ctxStream{ServerStream: ss, ctx: ctx})
	}
}

func authenticateCtx(ctx context.Context, tm *jwtutils.TokenManager, logger *zap.Logger, fullMethod string) (context.Context, error) {
	if isPublicMethod(fullMethod) {
		return ctx, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Warn("Missing metadata", zap.String("method", fullMethod))
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Ключи в metadata — lower-case
	authz := ""
	if vals := md.Get("authorization"); len(vals) > 0 {
		authz = vals[0]
	}
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		logger.Warn("Invalid authorization header", zap.String("header", authz), zap.String("method", fullMethod))
		return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
	}
	tokenStr := parts[1]

	claims, err := tm.ParseToken(tokenStr)
	if err != nil {
		logger.Warn("Failed to parse token", zap.Error(err), zap.String("method", fullMethod))
		return nil, status.Error(codes.Unauthenticated, "invalid token: "+err.Error())
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		logger.Warn("UserID not found in claims", zap.String("method", fullMethod))
		return nil, status.Error(codes.Unauthenticated, "userID not found in token")
	}

	logger.Debug("Token validated", zap.String("userID", userID), zap.String("method", fullMethod))

	ctx = jwtauth.WithUserID(ctx, userID)
	return ctx, nil
}

func isPublicMethod(method string) bool {
	publicMethods := map[string]bool{
		"/gophkeeper.proto.AuthService/Register": true,
		"/gophkeeper.proto.AuthService/Login":    true,
		// добавьте сюда другие публичные методы, не требующие аутентификации
	}
	return publicMethods[method]
}
