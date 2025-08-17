package grpc_test

import (
	"context"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	"github.com/ryabkov82/gophkeeper/internal/server/config"
	"github.com/ryabkov82/gophkeeper/internal/server/grpc"
)

// --- мок AuthService ---
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, login, password string) error {
	args := m.Called(ctx, login, password)
	return args.Error(0)
}

func (m *mockAuthService) Login(ctx context.Context, login, password string) (string, []byte, error) {
	args := m.Called(ctx, login, password)

	var salt []byte
	if s, ok := args.Get(1).([]byte); ok {
		salt = s
	}

	return args.String(0), salt, args.Error(2)
}

func getFreePort(t *testing.T) string {
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().String()
}

type mockServiceFactory struct{}

func (m *mockServiceFactory) Auth() service.AuthService {
	return &mockAuthService{}
}
func (m *mockServiceFactory) Credential() service.CredentialService {
	return nil
}
func (m *mockServiceFactory) BankCard() service.BankCardService {
	return nil
}

func (m *mockServiceFactory) TextData() service.TextDataService {
	return nil
}

func TestGRPCServer_StartAndGracefulShutdown(t *testing.T) {
	addr := getFreePort(t)

	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	defer lis.Close()

	cfg := &config.Config{
		GRPCServerAddr: addr,
		EnableTLS:      false,
		JwtKey:         "supersecretjwtkeystringwith32bytes!!",
	}

	logger := zap.NewNop()
	serviceFactory := &mockServiceFactory{}

	srv, err := grpc.NewGRPCServer(cfg, logger, serviceFactory)
	require.NoError(t, err)

	sigChan := make(chan os.Signal, 1)

	go func() {
		time.Sleep(1 * time.Second)
		sigChan <- syscall.SIGINT
	}()

	grpc.ServeGRPC(srv, lis, sigChan, logger)
}

func TestNewGRPCServer_NoTLS(t *testing.T) {
	cfg := &config.Config{
		EnableTLS: false,
		JwtKey:    "supersecretjwtkeystringwith32bytes!!",
	}

	logger := zap.NewNop()
	serviceFactory := &mockServiceFactory{}

	srv, err := grpc.NewGRPCServer(cfg, logger, serviceFactory)
	require.NoError(t, err)
	require.NotNil(t, srv)
}

func TestNewGRPCServer_WithInvalidTLSFiles(t *testing.T) {
	cfg := &config.Config{
		EnableTLS:   true,
		SSLCertFile: "nonexistent.crt",
		SSLKeyFile:  "nonexistent.key",
	}

	logger := zap.NewNop()
	serviceFactory := &mockServiceFactory{}

	_, err := grpc.NewGRPCServer(cfg, logger, serviceFactory)
	require.Error(t, err)
}
