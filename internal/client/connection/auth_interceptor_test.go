package connection

import (
	"context"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/service/auth"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAuthUnaryInterceptor(t *testing.T) {
	logger := zap.NewNop()

	// Мок authManager с заданным токеном
	authManagerWithToken := &mockAuthManager{
		token: "mock-token",
	}

	// Мок authManager без токена
	authManagerNoToken := &mockAuthManager{
		token: "",
	}

	// Мок invoker: проверяет, что в контексте присутствует токен, если должен быть
	mockInvoker := func(expectedToken string) grpc.UnaryInvoker {
		return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, _ := metadata.FromOutgoingContext(ctx)
			authHeaders := md.Get("authorization")
			if expectedToken == "" {
				if len(authHeaders) != 0 {
					t.Errorf("expected no authorization header, got %v", authHeaders)
				}
			} else {
				if len(authHeaders) == 0 || authHeaders[0] != "Bearer "+expectedToken {
					t.Errorf("expected authorization header 'Bearer %s', got %v", expectedToken, authHeaders)
				}
			}
			return nil
		}
	}

	tests := []struct {
		name          string
		authManager   auth.AuthManagerIface
		method        string
		expectedToken string
	}{
		{"WithToken_NotExcludedMethod", authManagerWithToken, "/package.Service/DoSomething", "mock-token"},
		{"WithToken_ExcludedMethodLogin", authManagerWithToken, "/package.Service/Login", ""},
		{"NoToken_NotExcludedMethod", authManagerNoToken, "/package.Service/DoSomething", ""},
		{"NoToken_ExcludedMethodRegister", authManagerNoToken, "/package.Service/Register", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := AuthUnaryInterceptor(tt.authManager, logger)

			err := interceptor(
				context.Background(),
				tt.method,
				nil,
				nil,
				nil,
				mockInvoker(tt.expectedToken),
			)
			require.NoError(t, err)
		})
	}
}

type mockAuthManager struct {
	token string
}

func (m *mockAuthManager) Register(ctx context.Context, login, password string) error {
	// Можно заглушку, если не нужен в тестах
	return nil
}

func (m *mockAuthManager) Login(ctx context.Context, login, password string) ([]byte, error) {
	// Можно возвращать фиктивную соль или nil, nil если не тестируем логику входа
	return []byte("fake_salt"), nil
}

func (m *mockAuthManager) SetClient(client proto.AuthServiceClient) {
	// Заглушка, ничего не делаем
}

func (m *mockAuthManager) GetToken() string {
	return m.token
}
