package credential

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CredentialManagerIface описывает интерфейс управления учётными данными (логины/пароли).
type CredentialManagerIface interface {
	CreateCredential(ctx context.Context, cred *model.Credential) error
	GetCredentialByID(ctx context.Context, id string) (*model.Credential, error)
	GetCredentials(ctx context.Context) ([]model.Credential, error)
	UpdateCredential(ctx context.Context, cred *model.Credential) error
	DeleteCredential(ctx context.Context, id string) error
	SetClient(client pb.CredentialServiceClient)
}

// CredentialManager управляет CRUD операциями с учётными данными,
// взаимодействует с сервером по gRPC и логирует операции.
type CredentialManager struct {
	logger *zap.Logger
	client pb.CredentialServiceClient // для инъекции моков в тестах
}

// NewCredentialManager создаёт новый CredentialManager.
func NewCredentialManager(logger *zap.Logger) *CredentialManager {
	return &CredentialManager{
		logger: logger,
	}
}

// SetClient позволяет установить кастомный (например, моковый) gRPC-клиент.
// Это полезно для тестирования, чтобы подменить реальный клиент на мок.
func (m *CredentialManager) SetClient(client pb.CredentialServiceClient) {
	m.client = client
}

// CreateCredential создаёт новую учётную запись (логин/пароль) на сервере.
func (m *CredentialManager) CreateCredential(ctx context.Context, cred *model.Credential) error {
	m.logger.Debug("CreateCredential request started",
		zap.String("userID", cred.UserID),
		zap.String("title", cred.Title),
	)

	req := &pb.CreateCredentialRequest{}
	req.SetCredential(toProtoCredential(cred))

	_, err := m.client.CreateCredential(ctx, req)
	if err != nil {
		m.logger.Error("CreateCredential RPC failed", zap.Error(err))
		return fmt.Errorf("CreateCredential RPC failed: %w", err)
	}

	m.logger.Info("CreateCredential succeeded",
		zap.String("userID", cred.UserID),
		zap.String("title", cred.Title),
	)
	return nil
}

// GetCredentialByID получает учётные данные по ID.
func (m *CredentialManager) GetCredentialByID(ctx context.Context, id string) (*model.Credential, error) {
	m.logger.Debug("GetCredentialByID request started",
		zap.String("credentialID", id),
	)

	var cred *model.Credential
	req := &pb.GetCredentialByIDRequest{}
	req.SetId(id)

	resp, err := m.client.GetCredentialByID(ctx, req)
	if err != nil {
		m.logger.Error("GetCredentialByID RPC failed", zap.Error(err))
		return nil, fmt.Errorf("GetCredentialByID RPC failed: %w", err)
	}

	cred = fromProtoCredential(resp.GetCredential())

	m.logger.Info("GetCredentialByID succeeded",
		zap.String("credentialID", id),
	)
	return cred, nil
}

// GetCredentialsByUserID получает список учётных данных по ID пользователя.
func (m *CredentialManager) GetCredentials(ctx context.Context) ([]model.Credential, error) {

	m.logger.Debug("GetCredentialsB request started")

	var creds []model.Credential

	resp, err := m.client.GetCredentials(ctx, &emptypb.Empty{})
	if err != nil {
		m.logger.Error("GetCredentialsByUserID RPC failed", zap.Error(err))
		return nil, fmt.Errorf("GetCredentialsByUserID RPC failed: %w", err)
	}

	creds = make([]model.Credential, 0, len(resp.GetCredentials()))
	for _, pbCred := range resp.GetCredentials() {
		creds = append(creds, *fromProtoCredential(pbCred))
	}
	m.logger.Info("GetCredentialsByUserID succeeded",
		zap.Int("count", len(creds)),
	)
	return creds, nil
}

// UpdateCredential обновляет существующую учётную запись на сервере.
func (m *CredentialManager) UpdateCredential(ctx context.Context, cred *model.Credential) error {
	m.logger.Debug("UpdateCredential request started",
		zap.String("credentialID", cred.ID),
	)

	req := &pb.UpdateCredentialRequest{}
	req.SetCredential(toProtoCredential(cred))

	_, err := m.client.UpdateCredential(ctx, req)
	if err != nil {
		m.logger.Error("UpdateCredential RPC failed", zap.Error(err))
		return fmt.Errorf("UpdateCredential RPC failed: %w", err)
	}

	m.logger.Info("UpdateCredential succeeded",
		zap.String("credentialID", cred.ID),
	)
	return nil
}

// DeleteCredential удаляет учётные данные по ID.
func (m *CredentialManager) DeleteCredential(ctx context.Context, id string) error {
	m.logger.Debug("DeleteCredential request started",
		zap.String("credentialID", id),
	)

	req := &pb.DeleteCredentialRequest{}
	req.SetId(id)

	_, err := m.client.DeleteCredential(ctx, req)
	if err != nil {
		m.logger.Error("DeleteCredential RPC failed", zap.Error(err))
		return fmt.Errorf("DeleteCredential RPC failed: %w", err)
	}

	m.logger.Info("DeleteCredential succeeded",
		zap.String("credentialID", id),
	)
	return nil
}

// Преобразования между model.Credential и pb.Credential
func toProtoCredential(c *model.Credential) *pb.Credential {
	cred := &pb.Credential{}
	cred.SetId(c.ID)
	cred.SetUserId(c.UserID)
	cred.SetTitle(c.Title)
	cred.SetLogin(c.Login)
	cred.SetPassword(c.Password)
	cred.SetMetadata(c.Metadata)
	cred.SetCreatedAt(timestamppb.New(c.CreatedAt))
	cred.SetUpdatedAt(timestamppb.New(c.UpdatedAt))
	return cred
}

func fromProtoCredential(pbCred *pb.Credential) *model.Credential {
	return &model.Credential{
		ID:        pbCred.GetId(),
		UserID:    pbCred.GetUserId(),
		Title:     pbCred.GetTitle(),
		Login:     pbCred.GetLogin(),
		Password:  pbCred.GetPassword(),
		Metadata:  pbCred.GetMetadata(),
		CreatedAt: pbCred.GetCreatedAt().AsTime(),
		UpdatedAt: pbCred.GetUpdatedAt().AsTime(),
	}
}
