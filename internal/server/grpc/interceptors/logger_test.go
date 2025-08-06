package interceptors_test

import (
	"context"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/server/grpc/interceptors"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLoggingInterceptor_HandlerCalledAndLogs(t *testing.T) {
	// Создаём тестовый логгер
	logger := zaptest.NewLogger(t)

	// Переменные для проверки вызова
	called := false

	// Мокаем handler
	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "response", nil
	}

	// Мокаем UnaryServerInfo
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	// Создаём интерцептор
	interceptor := interceptors.LoggingInterceptor(logger)

	// Вызываем
	resp, err := interceptor(context.Background(), "request", info, mockHandler)

	// Проверяем
	assert.True(t, called, "handler should have been called")
	assert.Equal(t, "response", resp)
	assert.NoError(t, err)
}

func TestLoggingInterceptor_WithError(t *testing.T) {
	logger := zaptest.NewLogger(t)

	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.InvalidArgument, "invalid input")
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/WithError",
	}

	interceptor := interceptors.LoggingInterceptor(logger)

	resp, err := interceptor(context.Background(), "bad-request", info, mockHandler)

	assert.Nil(t, resp)
	assert.Error(t, err)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}
