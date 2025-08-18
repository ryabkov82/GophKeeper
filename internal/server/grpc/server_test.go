package grpc_test

import (
	"context"
	"io"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
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

// --- мок BinaryDataService ---
type mockBinaryDataService struct {
	mock.Mock
}

func (m *mockBinaryDataService) Create(ctx context.Context, userID string, title string, metadata string, r io.Reader) (*model.BinaryData, error) {
	args := m.Called(ctx, userID, title, metadata, r)
	var bd *model.BinaryData
	if v := args.Get(0); v != nil {
		bd = v.(*model.BinaryData)
	}
	return bd, args.Error(1)
}

func (m *mockBinaryDataService) Update(ctx context.Context, userID, id, title, metadata string, r io.Reader) (*model.BinaryData, error) {
	args := m.Called(ctx, userID, id, title, metadata, r)
	var bd *model.BinaryData
	if v := args.Get(0); v != nil {
		bd = v.(*model.BinaryData)
	}
	return bd, args.Error(1)
}

func (m *mockBinaryDataService) Get(ctx context.Context, userID, id string) (*model.BinaryData, io.ReadCloser, error) {
	args := m.Called(ctx, userID, id)
	var bd *model.BinaryData
	if v := args.Get(0); v != nil {
		bd = v.(*model.BinaryData)
	}
	var rc io.ReadCloser
	if v := args.Get(1); v != nil {
		rc = v.(io.ReadCloser)
	}
	return bd, rc, args.Error(2)
}

func (m *mockBinaryDataService) List(ctx context.Context, userID string) ([]*model.BinaryData, error) {
	args := m.Called(ctx, userID)
	var list []*model.BinaryData
	if v := args.Get(0); v != nil {
		list = v.([]*model.BinaryData)
	}
	return list, args.Error(1)
}

func (m *mockBinaryDataService) Delete(ctx context.Context, userID, id string) error {
	return m.Called(ctx, userID, id).Error(0)
}

func (m *mockBinaryDataService) Close() {
	m.Called()
}

// --- мок ServiceFactory ---
type mockServiceFactory struct {
	binarySvc *mockBinaryDataService
}

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

func (m *mockServiceFactory) BinaryData() service.BinaryDataService {
	return m.binarySvc
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
	binarySvc := &mockBinaryDataService{}
	binarySvc.On("Close").Return() // метод ничего не возвращает

	serviceFactory := &mockServiceFactory{binarySvc: binarySvc}

	srv, err := grpc.NewGRPCServer(cfg, logger, serviceFactory)
	require.NoError(t, err)

	sigChan := make(chan os.Signal, 1)

	go func() {
		time.Sleep(1 * time.Second)
		sigChan <- syscall.SIGINT
	}()

	grpc.ServeGRPC(srv, lis, sigChan, logger, serviceFactory)

	// Проверяем, что Close() был вызван
	binarySvc.AssertCalled(t, "Close")
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
