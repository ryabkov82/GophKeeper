// Package retry предоставляет перехватчик gRPC-клиента, обеспечивающий
// автоматическое переподключение к серверу перед выполнением запроса,
// если текущее соединение не готово.
//
// Это особенно полезно в интерактивных приложениях, где клиент может
// временно потерять соединение с сервером (например, при потере сети)
// и должен попытаться восстановить его автоматически без перезапуска приложения.
//
// Перехватчик можно использовать при создании gRPC клиента:
//
//	conn, _ := grpc.Dial(
//	    address,
//	    grpc.WithUnaryInterceptor(retry.WithReconnect(manager)),
//	    ...
//	)
//
// Функции:
//   - WithReconnect: возвращает UnaryClientInterceptor, который проверяет состояние соединения
//     и, если оно не готово, пытается переподключиться через переданный connection.Manager.
package retry

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/client/connection"
	"google.golang.org/grpc"
)

// WithReconnect возвращает gRPC UnaryClientInterceptor, который автоматически
// проверяет готовность gRPC-соединения перед выполнением каждого запроса.
//
// Если соединение не находится в состоянии `connectivity.Ready`, перехватчик
// попытается переподключиться с помощью метода `Connect` менеджера соединений.
//
// Это поведение полезно в клиентских приложениях с нестабильной сетью —
// позволяет избежать ошибок из-за неготового соединения и выполнить попытку
// восстановления до начала запроса.
//
// Параметры:
//   - manager: указатель на connection.Manager, который управляет состоянием gRPC-соединения.
//
// Возвращает:
//   - grpc.UnaryClientInterceptor: перехватчик, который можно использовать с `grpc.WithUnaryInterceptor`.
func WithReconnect(manager *connection.Manager) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Автоматическое переподключение перед вызовом
		if !manager.IsReady() {
			if _, err := manager.Connect(ctx); err != nil {
				return err // Пользователь увидит ошибку и решит, что делать
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
