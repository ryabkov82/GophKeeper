package app_test

import (
	"context"
	"io"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

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

func (m *mockCredentialManager) GetCredentials(ctx context.Context) ([]model.Credential, error) {
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

// mockBankCardManager - мок BankcardManagerIface
type mockBankCardManager struct {
	setClientCalled bool
	client          proto.BankCardServiceClient

	createErr     error
	getByIDResult *model.BankCard
	getByIDErr    error
	getAllResult  []model.BankCard
	getAllErr     error
	updateErr     error
	deleteErr     error
}

func (m *mockBankCardManager) SetClient(client proto.BankCardServiceClient) {
	m.setClientCalled = true
	m.client = client
}

func (m *mockBankCardManager) CreateBankCard(ctx context.Context, card *model.BankCard) error {
	return m.createErr
}

func (m *mockBankCardManager) GetBankCardByID(ctx context.Context, id string) (*model.BankCard, error) {
	return m.getByIDResult, m.getByIDErr
}

func (m *mockBankCardManager) GetBankCards(ctx context.Context) ([]model.BankCard, error) {
	return m.getAllResult, m.getAllErr
}

func (m *mockBankCardManager) UpdateBankCard(ctx context.Context, card *model.BankCard) error {
	return m.updateErr
}

func (m *mockBankCardManager) DeleteBankCard(ctx context.Context, id string) error {
	return m.deleteErr
}

// mockTextDataManager - мок TextDataManagerIface
type mockTextDataManager struct {
	createErr       error
	getByIDResult   *model.TextData
	getByIDErr      error
	getTitlesResult []*model.TextData
	getTitlesErr    error
	updateErr       error
	deleteErr       error
	setClientCalled bool
}

func (m *mockTextDataManager) CreateTextData(ctx context.Context, td *model.TextData) error {
	return m.createErr
}

func (m *mockTextDataManager) GetTextDataByID(ctx context.Context, id string) (*model.TextData, error) {
	return m.getByIDResult, m.getByIDErr
}

func (m *mockTextDataManager) GetTextDataTitles(ctx context.Context) ([]*model.TextData, error) {
	return m.getTitlesResult, m.getTitlesErr
}

func (m *mockTextDataManager) UpdateTextData(ctx context.Context, td *model.TextData) error {
	return m.updateErr
}

func (m *mockTextDataManager) DeleteTextData(ctx context.Context, id string) error {
	return m.deleteErr
}

func (m *mockTextDataManager) SetClient(client proto.TextDataServiceClient) {
	m.setClientCalled = true
}

// mockBinaryDataManager - мок BinaryDataManagerIface
type mockBinaryDataManager struct {
	setClientCalled bool
	client          proto.BinaryDataServiceClient

	uploadErr     error
	updateErr     error
	updateInfoErr error
	createInfoErr error
	downloadFn    func(ctx context.Context, id string) (io.ReadCloser, error)
	deleteErr     error
	listResult    []model.BinaryData
	listErr       error
	getInfoFn     func(ctx context.Context, id string) (*model.BinaryData, error)
}

func (m *mockBinaryDataManager) SetClient(client proto.BinaryDataServiceClient) {
	m.setClientCalled = true
	m.client = client
}

func (m *mockBinaryDataManager) Upload(ctx context.Context, data *model.BinaryData, content io.Reader) error {
	return m.uploadErr
}

func (m *mockBinaryDataManager) Update(ctx context.Context, data *model.BinaryData, content io.Reader) error {
	return m.updateErr
}

func (m *mockBinaryDataManager) UpdateInfo(ctx context.Context, data *model.BinaryData) error {
	return m.updateInfoErr
}

func (m *mockBinaryDataManager) CreateInfo(ctx context.Context, data *model.BinaryData) error {
	return m.createInfoErr
}

func (m *mockBinaryDataManager) Download(ctx context.Context, id string) (io.ReadCloser, error) {
	if m.downloadFn != nil {
		return m.downloadFn(ctx, id)
	}
	return nil, nil
}

func (m *mockBinaryDataManager) Delete(ctx context.Context, id string) error {
	return m.deleteErr
}

func (m *mockBinaryDataManager) List(ctx context.Context) ([]model.BinaryData, error) {
	return m.listResult, m.listErr
}

func (m *mockBinaryDataManager) GetInfo(ctx context.Context, id string) (*model.BinaryData, error) {
	if m.getInfoFn != nil {
		return m.getInfoFn(ctx, id)
	}
	return nil, nil
}

func TestNewAppServices_Success(t *testing.T) {
	tempLogDir := t.TempDir()

	cfg := &config.ClientConfig{
		ServerAddress: "localhost:50051",
		UseTLS:        false,
		LogLevel:      "debug",
		Timeout:       2 * time.Second,
		LogDirPath:    tempLogDir,
	}

	services, err := app.NewAppServices(cfg)
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

	// Некорректный путь
	badDir := string([]byte{0})
	cfg := &config.ClientConfig{
		LogLevel:   "debug",
		LogDirPath: badDir,
	}

	services, err := app.NewAppServices(cfg)
	assert.Error(t, err, "ожидалась ошибка при создании директории")
	assert.Nil(t, services, "services должен быть nil при ошибке")
}

func TestNewAppServices_LoggerInitFail(t *testing.T) {

	tempLogDir := t.TempDir()
	// Чтобы сломать инициализацию логгера, можно передать некорректный уровень логирования
	cfg := &config.ClientConfig{
		LogLevel:   "invalid_level", // заведомо некорректно
		LogDirPath: tempLogDir,
	}

	services, err := app.NewAppServices(cfg)
	assert.Error(t, err, "ожидалась ошибка при неправильном уровне логирования")
	assert.Nil(t, services, "services должен быть nil при ошибке")
}
