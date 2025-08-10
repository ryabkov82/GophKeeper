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

		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		logger.Debug("Auth interceptor triggered", zap.String("method", info.FullMethod))

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("Missing metadata in context", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			logger.Warn("Authorization token not provided", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		authHeader := authHeaders[0]
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			logger.Warn("Invalid authorization header format", zap.String("header", authHeader), zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
		}
		tokenStr := parts[1]

		claims, err := tm.ParseToken(tokenStr)
		if err != nil {
			logger.Warn("Failed to parse token", zap.Error(err), zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "invalid token: "+err.Error())
		}

		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			logger.Warn("UserID not found in token claims", zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "userID not found in token")
		}

		logger.Debug("Token validated, userID extracted", zap.String("userID", userID), zap.String("method", info.FullMethod))

		newCtx := jwtauth.WithUserID(ctx, userID)
		return handler(newCtx, req)
	}
}

func isPublicMethod(method string) bool {
	publicMethods := map[string]bool{
		"/gophkeeper.proto.AuthService/Register": true,
		"/gophkeeper.proto.AuthService/Login":    true,
		// добавьте сюда другие публичные методы, не требующие аутентификации
	}
	return publicMethods[method]
}
