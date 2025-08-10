// Package jwtauth предоставляет функции для интеграции JWT-токенов в контекст
// приложения и удобного извлечения информации о пользователе из контекста.
//
// В частности, пакет позволяет:
//   - Добавлять идентификатор пользователя (userID) в context.Context.
//   - Извлекать userID из контекста, с обработкой ошибок при отсутствии.
//
// Этот пакет удобен для использования в gRPC или HTTP middleware/interceptor,
// где после проверки JWT токена нужно "прикрепить" userID к контексту запроса
// и далее использовать его в бизнес-логике.
//
// Пример использования:
//
//	import "github.com/ryabkov82/gophkeeper/internal/pkg/jwtauth"
//
//	func SomeInterceptor(ctx context.Context, userID string) context.Context {
//	    return jwtauth.WithUserID(ctx, userID)
//	}
//
//	func SomeHandler(ctx context.Context) {
//	    userID, err := jwtauth.FromContext(ctx)
//	    if err != nil {
//	        // обработка отсутствия userID в контексте
//	    }
//	    // использование userID
//	}
package jwtauth

import (
	"context"
	"errors"
)

type contextKey string

const userIDKey = contextKey("userID")

// WithUserID возвращает новый контекст, в который добавлен идентификатор пользователя userID.
// Используется для "прикрепления" userID к контексту запроса после успешной аутентификации.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// FromContext извлекает идентификатор пользователя userID из контекста.
// Возвращает ошибку, если userID отсутствует или имеет неверный тип.
func FromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("userID not found in context")
	}
	return userID, nil
}
