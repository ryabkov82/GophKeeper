package app_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/config"
	"github.com/ryabkov82/gophkeeper/internal/client/connection"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/logger"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

func TestNewAppServices_Success(t *testing.T) {
	tempLogDir := t.TempDir()

	cfg := &config.ClientConfig{
		ServerAddress: "localhost:50051",
		UseTLS:        false,
		LogLevel:      "debug",
		Timeout:       2 * time.Second,
	}

	services, err := app.NewAppServices(cfg, tempLogDir)
	assert.NoError(t, err, "NewAppServices не должен возвращать ошибку")
	assert.NotNil(t, services, "services должен быть создан")

	// Проверяем, что зависимости созданы
	assert.NotNil(t, services.Logger, "Logger должен быть инициализирован")
	assert.NotNil(t, services.AuthManager, "AuthManager должен быть инициализирован")
	assert.NotNil(t, services.ConnManager, "ConnManager должен быть инициализирован")

	// Проверяем, что каталог логов существует
	_, err = os.Stat(tempLogDir)
	assert.NoError(t, err, "Каталог логов должен существовать")

	// Закрываем ресурсы
	services.Close()
	logger.Close()
}

func TestNewAppServices_MkdirFail(t *testing.T) {
	cfg := &config.ClientConfig{
		LogLevel: "debug",
	}

	// Некорректный путь
	badDir := string([]byte{0})

	services, err := app.NewAppServices(cfg, badDir)
	assert.Error(t, err, "ожидалась ошибка при создании директории")
	assert.Nil(t, services, "services должен быть nil при ошибке")
}

func TestNewAppServices_LoggerInitFail(t *testing.T) {
	// Чтобы сломать инициализацию логгера, можно передать некорректный уровень логирования
	cfg := &config.ClientConfig{
		LogLevel: "invalid_level", // заведомо некорректно
	}

	tempLogDir := t.TempDir()

	services, err := app.NewAppServices(cfg, tempLogDir)
	assert.Error(t, err, "ожидалась ошибка при неправильном уровне логирования")
	assert.Nil(t, services, "services должен быть nil при ошибке")
}

type mockGrpcConn struct{}

func (m *mockGrpcConn) Close() error {
	return nil
}

func (m *mockGrpcConn) GetState() connectivity.State {
	return connectivity.Ready
}

func (m *mockGrpcConn) Connect() {
	// Заглушка, ничего не делает
}

func (m *mockGrpcConn) WaitForStateChange(ctx context.Context, s connectivity.State) bool {
	// Просто возвращаем true — имитируем изменение состояния
	return true
}

// grpc.ClientConnInterface требует реализации методов Invoke и NewStream
func (m *mockGrpcConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	return nil
}

func (m *mockGrpcConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

type mockConnManager struct {
	connectCalled bool
	closeCalled   bool
	connectErr    error
}

func (m *mockConnManager) Connect(ctx context.Context) (connection.GrpcConn, error) {
	m.connectCalled = true
	return &mockGrpcConn{}, m.connectErr
}

func (m *mockConnManager) Close() error {
	m.closeCalled = true
	return nil
}

type mockAuthManager struct {
	token           string
	loginCalled     bool
	registerCalled  bool
	registerErr     error
	loginErr        error
	saltToReturn    []byte
	setClientCalled bool
}

func (m *mockAuthManager) Register(ctx context.Context, login, password string) error {
	m.registerCalled = true
	return m.registerErr
}

func (m *mockAuthManager) Login(ctx context.Context, login, password string) ([]byte, error) {
	m.loginCalled = true
	return m.saltToReturn, m.loginErr
}

func (m *mockAuthManager) SetClient(client proto.AuthServiceClient) {
	m.setClientCalled = true
}

func (m *mockAuthManager) GetToken() string {
	return m.token
}

type mockCryptoKeyManager struct {
	generateErr error
	loadKeyData []byte
	loadErr     error
	clearErr    error

	generateCalled bool
	loadCalled     bool
	clearCalled    bool
}

func (m *mockCryptoKeyManager) GenerateAndSaveKey(password string, salt []byte) error {
	m.generateCalled = true
	return m.generateErr
}

func (m *mockCryptoKeyManager) LoadKey() ([]byte, error) {
	m.loadCalled = true
	return m.loadKeyData, m.loadErr
}

func (m *mockCryptoKeyManager) ClearKey() error {
	m.clearCalled = true
	return m.clearErr
}

func TestLoginUser_Success(t *testing.T) {
	authMgr := &mockAuthManager{saltToReturn: []byte("salt")}
	cryptoMgr := &mockCryptoKeyManager{}
	connMgr := &mockConnManager{}

	appSvc := &app.AppServices{
		AuthManager:      authMgr,
		CryptoKeyManager: cryptoMgr,
		ConnManager:      connMgr,
		Logger:           zap.NewNop(),
	}

	err := appSvc.LoginUser(context.Background(), "user", "pass")
	require.NoError(t, err)
	require.True(t, authMgr.setClientCalled)
	require.True(t, authMgr.loginCalled)
	require.True(t, cryptoMgr.generateCalled)
	require.True(t, connMgr.connectCalled)
}

func TestLoginUser_FailEmptySalt(t *testing.T) {
	authMgr := &mockAuthManager{saltToReturn: []byte{}}
	cryptoMgr := &mockCryptoKeyManager{}
	connMgr := &mockConnManager{}

	appSvc := &app.AppServices{
		AuthManager:      authMgr,
		CryptoKeyManager: cryptoMgr,
		ConnManager:      connMgr,
		Logger:           zap.NewNop(),
	}

	err := appSvc.LoginUser(context.Background(), "user", "pass")
	require.ErrorContains(t, err, "no salt received")
}

func TestLoginUser_GenerateKeyError(t *testing.T) {
	authMgr := &mockAuthManager{saltToReturn: []byte("salt")}
	cryptoMgr := &mockCryptoKeyManager{generateErr: fmt.Errorf("fail generate")}
	connMgr := &mockConnManager{}

	appSvc := &app.AppServices{
		AuthManager:      authMgr,
		CryptoKeyManager: cryptoMgr,
		ConnManager:      connMgr,
		Logger:           zap.NewNop(),
	}

	err := appSvc.LoginUser(context.Background(), "user", "pass")
	require.ErrorContains(t, err, "failed to generate encryption key")
}

func TestRegisterUser_Success(t *testing.T) {
	authMgr := &mockAuthManager{}
	cryptoMgr := &mockCryptoKeyManager{}
	connMgr := &mockConnManager{}

	appSvc := &app.AppServices{
		AuthManager:      authMgr,
		CryptoKeyManager: cryptoMgr,
		ConnManager:      connMgr,
		Logger:           zap.NewNop(),
	}

	// Мокаем Register: без ошибок
	authMgr.registerErr = nil
	// LoginUser внутри RegisterUser — используем тот же mockAuthManager, salt возвращаем нормальный
	authMgr.saltToReturn = []byte("salt")

	err := appSvc.RegisterUser(context.Background(), "user", "pass")
	require.NoError(t, err)
	require.True(t, authMgr.setClientCalled)
	require.True(t, authMgr.registerCalled)
	require.True(t, authMgr.loginCalled)
	require.True(t, cryptoMgr.generateCalled)
}

func TestRegisterUser_RegisterFail(t *testing.T) {
	authMgr := &mockAuthManager{registerErr: fmt.Errorf("register failed")}
	connMgr := &mockConnManager{}
	appSvc := &app.AppServices{
		AuthManager: authMgr,
		ConnManager: connMgr,
		Logger:      zap.NewNop(),
	}

	err := appSvc.RegisterUser(context.Background(), "user", "pass")
	require.ErrorContains(t, err, "register failed")
}

// mockCredentialManager - мок CredentialManagerIface
type mockCredentialManager struct {
	createErr         error
	getByIDResult     *model.Credential
	getByIDErr        error
	getByUserIDResult []model.Credential
	getByUserIDErr    error
	updateErr         error
	deleteErr         error
	setClientCalled   bool
}

func (m *mockCredentialManager) CreateCredential(ctx context.Context, cred *model.Credential) error {
	return m.createErr
}

func (m *mockCredentialManager) GetCredentialByID(ctx context.Context, id string) (*model.Credential, error) {
	return m.getByIDResult, m.getByIDErr
}

func (m *mockCredentialManager) GetCredentialsByUserID(ctx context.Context, userID string) ([]model.Credential, error) {
	return m.getByUserIDResult, m.getByUserIDErr
}

func (m *mockCredentialManager) UpdateCredential(ctx context.Context, cred *model.Credential) error {
	return m.updateErr
}

func (m *mockCredentialManager) DeleteCredential(ctx context.Context, id string) error {
	return m.deleteErr
}

func (m *mockCredentialManager) SetClient(client proto.CredentialServiceClient) {
	m.setClientCalled = true
}

func TestCreateCredential(t *testing.T) {
	ctx := context.Background()

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:       &mockConnManager{connectErr: errors.New("connect failed")},
		CredentialManager: &mockCredentialManager{},
		Logger:            zap.NewNop(),
	}
	err := appSvc.CreateCredential(ctx, &model.Credential{})
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешный вызов
	credMgr := &mockCredentialManager{}
	appSvc = &app.AppServices{
		ConnManager:       &mockConnManager{},
		CredentialManager: credMgr,
		Logger:            zap.NewNop(),
	}
	err = appSvc.CreateCredential(ctx, &model.Credential{})
	require.NoError(t, err)
	require.True(t, credMgr.setClientCalled)
}

func TestGetCredentialByID(t *testing.T) {
	ctx := context.Background()
	expectedCred := &model.Credential{ID: "123"}

	appSvc := &app.AppServices{
		ConnManager: &mockConnManager{},
		CredentialManager: &mockCredentialManager{
			getByIDResult: expectedCred,
		},
		Logger: zap.NewNop(),
	}

	cred, err := appSvc.GetCredentialByID(ctx, "123")
	require.NoError(t, err)
	require.Equal(t, expectedCred, cred)

	// Ошибка подключения
	appSvc.ConnManager = &mockConnManager{connectErr: errors.New("connect error")}
	cred, err = appSvc.GetCredentialByID(ctx, "123")
	require.Error(t, err)
	require.Nil(t, cred)
	require.EqualError(t, err, "connect error")
}

func TestGetCredentialsByUserID(t *testing.T) {
	ctx := context.Background()
	expectedList := []model.Credential{
		{ID: "1"},
		{ID: "2"},
	}

	appSvc := &app.AppServices{
		ConnManager: &mockConnManager{},
		CredentialManager: &mockCredentialManager{
			getByUserIDResult: expectedList,
		},
		Logger: zap.NewNop(),
	}

	creds, err := appSvc.GetCredentialsByUserID(ctx, "user1")
	require.NoError(t, err)
	require.Equal(t, expectedList, creds)

	// Ошибка подключения
	appSvc.ConnManager = &mockConnManager{connectErr: errors.New("connect error")}
	creds, err = appSvc.GetCredentialsByUserID(ctx, "user1")
	require.Error(t, err)
	require.Nil(t, creds)
	require.EqualError(t, err, "connect error")
}

func TestUpdateCredential(t *testing.T) {
	ctx := context.Background()

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:       &mockConnManager{connectErr: errors.New("connect failed")},
		CredentialManager: &mockCredentialManager{},
		Logger:            zap.NewNop(),
	}
	err := appSvc.UpdateCredential(ctx, &model.Credential{})
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешное обновление
	credMgr := &mockCredentialManager{}
	appSvc = &app.AppServices{
		ConnManager:       &mockConnManager{},
		CredentialManager: credMgr,
		Logger:            zap.NewNop(),
	}
	err = appSvc.UpdateCredential(ctx, &model.Credential{})
	require.NoError(t, err)
	require.True(t, credMgr.setClientCalled)
}

func TestDeleteCredential(t *testing.T) {
	ctx := context.Background()

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:       &mockConnManager{connectErr: errors.New("connect failed")},
		CredentialManager: &mockCredentialManager{},
		Logger:            zap.NewNop(),
	}
	err := appSvc.DeleteCredential(ctx, "id")
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешное удаление
	credMgr := &mockCredentialManager{}
	appSvc = &app.AppServices{
		ConnManager:       &mockConnManager{},
		CredentialManager: credMgr,
		Logger:            zap.NewNop(),
	}
	err = appSvc.DeleteCredential(ctx, "id")
	require.NoError(t, err)
	require.True(t, credMgr.setClientCalled)
}
