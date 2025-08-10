package credential

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/client/connection"
	"github.com/ryabkov82/gophkeeper/internal/client/service/auth"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CredentialManagerIface описывает интерфейс управления учётными данными (логины/пароли).
type CredentialManagerIface interface {
	CreateCredential(ctx context.Context, cred *model.Credential) error
	GetCredentialByID(ctx context.Context, id string) (*model.Credential, error)
	GetCredentialsByUserID(ctx context.Context, userID string) ([]model.Credential, error)
	UpdateCredential(ctx context.Context, cred *model.Credential) error
	DeleteCredential(ctx context.Context, id string) error
}

// CredentialManager управляет CRUD операциями с учётными данными,
// взаимодействует с сервером по gRPC и логирует операции.
type CredentialManager struct {
	connManager connection.ConnManager
	logger      *zap.Logger
	auth        auth.AuthManagerIface
	client      pb.CredentialServiceClient // для инъекции моков в тестах
}

// NewCredentialManager создаёт новый CredentialManager.
func NewCredentialManager(connManager connection.ConnManager, authManager auth.AuthManagerIface, logger *zap.Logger) *CredentialManager {
	return &CredentialManager{
		connManager: connManager,
		auth:        authManager,
		logger:      logger,
	}
}

// withClient упрощает работу с gRPC клиентом:
// открывает соединение и вызывает переданную функцию с клиентом
func (m *CredentialManager) withClient(ctx context.Context, fn func(client pb.CredentialServiceClient) error) error {
	conn, err := m.connManager.Connect(ctx)
	if err != nil {
		m.logger.Error("Connection failed", zap.Error(err))
		return fmt.Errorf("connection failed: %w", err)
	}

	client := m.getClient(conn)
	return fn(client)
}

// getClient возвращает gRPC клиента.
// Использует встроенный client для тестов, если он установлен,
// иначе создаёт нового клиента на основе соединения.
func (m *CredentialManager) getClient(conn grpc.ClientConnInterface) pb.CredentialServiceClient {
	if m.client != nil {
		return m.client
	}
	return pb.NewCredentialServiceClient(conn)
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

	ctx = m.auth.ContextWithToken(ctx)

	return m.withClient(ctx, func(client pb.CredentialServiceClient) error {
		req := &pb.CreateCredentialRequest{}
		req.SetCredential(toProtoCredential(cred))

		_, err := client.CreateCredential(ctx, req)
		if err != nil {
			m.logger.Error("CreateCredential RPC failed", zap.Error(err))
			return fmt.Errorf("CreateCredential RPC failed: %w", err)
		}

		m.logger.Info("CreateCredential succeeded",
			zap.String("userID", cred.UserID),
			zap.String("title", cred.Title),
		)
		return nil
	})
}

// GetCredentialByID получает учётные данные по ID.
func (m *CredentialManager) GetCredentialByID(ctx context.Context, id string) (*model.Credential, error) {
	m.logger.Debug("GetCredentialByID request started",
		zap.String("credentialID", id),
	)

	ctx = m.auth.ContextWithToken(ctx)
	var cred *model.Credential
	err := m.withClient(ctx, func(client pb.CredentialServiceClient) error {
		req := &pb.GetCredentialByIDRequest{}
		req.SetId(id)

		resp, err := client.GetCredentialByID(ctx, req)
		if err != nil {
			m.logger.Error("GetCredentialByID RPC failed", zap.Error(err))
			return fmt.Errorf("GetCredentialByID RPC failed: %w", err)
		}

		cred = fromProtoCredential(resp.GetCredential())
		return nil
	})

	if err != nil {
		return nil, err
	}

	m.logger.Info("GetCredentialByID succeeded",
		zap.String("credentialID", id),
	)
	return cred, nil
}

// GetCredentialsByUserID получает список учётных данных по ID пользователя.
func (m *CredentialManager) GetCredentialsByUserID(ctx context.Context, userID string) ([]model.Credential, error) {
	m.logger.Debug("GetCredentialsByUserID request started",
		zap.String("userID", userID),
	)

	ctx = m.auth.ContextWithToken(ctx)
	var creds []model.Credential
	err := m.withClient(ctx, func(client pb.CredentialServiceClient) error {
		req := &pb.GetCredentialsByUserIDRequest{}
		req.SetUserId(userID)

		resp, err := client.GetCredentialsByUserID(ctx, req)
		if err != nil {
			m.logger.Error("GetCredentialsByUserID RPC failed", zap.Error(err))
			return fmt.Errorf("GetCredentialsByUserID RPC failed: %w", err)
		}

		creds = make([]model.Credential, 0, len(resp.GetCredentials()))
		for _, pbCred := range resp.GetCredentials() {
			creds = append(creds, *fromProtoCredential(pbCred))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	m.logger.Info("GetCredentialsByUserID succeeded",
		zap.String("userID", userID),
		zap.Int("count", len(creds)),
	)
	return creds, nil
}

// UpdateCredential обновляет существующую учётную запись на сервере.
func (m *CredentialManager) UpdateCredential(ctx context.Context, cred *model.Credential) error {
	m.logger.Debug("UpdateCredential request started",
		zap.String("credentialID", cred.ID),
	)

	ctx = m.auth.ContextWithToken(ctx)
	return m.withClient(ctx, func(client pb.CredentialServiceClient) error {
		req := &pb.UpdateCredentialRequest{}
		req.SetCredential(toProtoCredential(cred))

		_, err := client.UpdateCredential(ctx, req)
		if err != nil {
			m.logger.Error("UpdateCredential RPC failed", zap.Error(err))
			return fmt.Errorf("UpdateCredential RPC failed: %w", err)
		}

		m.logger.Info("UpdateCredential succeeded",
			zap.String("credentialID", cred.ID),
		)
		return nil
	})
}

// DeleteCredential удаляет учётные данные по ID.
func (m *CredentialManager) DeleteCredential(ctx context.Context, id string) error {
	m.logger.Debug("DeleteCredential request started",
		zap.String("credentialID", id),
	)

	ctx = m.auth.ContextWithToken(ctx)
	return m.withClient(ctx, func(client pb.CredentialServiceClient) error {
		req := &pb.DeleteCredentialRequest{}
		req.SetId(id)

		_, err := client.DeleteCredential(ctx, req)
		if err != nil {
			m.logger.Error("DeleteCredential RPC failed", zap.Error(err))
			return fmt.Errorf("DeleteCredential RPC failed: %w", err)
		}

		m.logger.Info("DeleteCredential succeeded",
			zap.String("credentialID", id),
		)
		return nil
	})
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
