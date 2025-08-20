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

func TestAuthPerRPCCredentials_GetRequestMetadata(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name       string
		token      string
		method     string
		expectedMD map[string]string
	}{
		{
			name:       "WithToken_NotExcludedMethod",
			token:      "mock-token",
			method:     "/pkg.Service/Method",
			expectedMD: map[string]string{"authorization": "Bearer mock-token"},
		},
		{
			name:       "WithToken_ExcludedLogin",
			token:      "mock-token",
			method:     "/pkg.Service/Login",
			expectedMD: nil,
		},
		{
			name:       "WithToken_ExcludedRegister",
			token:      "mock-token",
			method:     "/pkg.Service/Register",
			expectedMD: nil,
		},
		{
			name:       "NoToken_NotExcludedMethod",
			token:      "",
			method:     "/pkg.Service/Method",
			expectedMD: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := NewAuthPerRPCCredentials(&mockAuthManager{token: tt.token}, logger, false)
			md, err := creds.GetRequestMetadata(context.Background(), tt.method)
			require.NoError(t, err)
			if tt.expectedMD == nil {
				require.Nil(t, md)
			} else {
				require.Equal(t, tt.expectedMD, md)
			}
		})
	}
}

func TestAuthPerRPCCredentials_RequireTransportSecurity(t *testing.T) {
	logger := zap.NewNop()
	credsTLS := NewAuthPerRPCCredentials(&mockAuthManager{}, logger, true)
	require.True(t, credsTLS.RequireTransportSecurity())
	credsNoTLS := NewAuthPerRPCCredentials(&mockAuthManager{}, logger, false)
	require.False(t, credsNoTLS.RequireTransportSecurity())
}

func TestAuthPerRPCCreds_shouldSkip(t *testing.T) {
	logger := zap.NewNop()
	creds := NewAuthPerRPCCredentials(&mockAuthManager{token: "token"}, logger, false).(*authPerRPCCreds)

	cases := []struct {
		method string
		want   bool
	}{
		{"/pkg.Service/Login", true},
		{"/pkg.Service/Register", true},
		{"/pkg.Service/Other", false},
		{"", false},
	}

	for _, c := range cases {
		t.Run(c.method, func(t *testing.T) {
			require.Equal(t, c.want, creds.shouldSkip(c.method))
		})
	}
}
