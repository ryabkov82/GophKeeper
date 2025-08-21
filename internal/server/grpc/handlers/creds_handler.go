package handlers

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtauth"
	"github.com/ryabkov82/gophkeeper/internal/pkg/mapper"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CredentialHandler реализует gRPC сервер для CredentialService
type CredentialHandler struct {
	pb.UnimplementedCredentialServiceServer
	service service.CredentialService
	logger  *zap.Logger
}

// NewCredentialHandler создает новый CredentialHandler с внедрением сервиса и логгера.
func NewCredentialHandler(srv service.CredentialService, logger *zap.Logger) *CredentialHandler {
	return &CredentialHandler{
		service: srv,
		logger:  logger,
	}
}

// CreateCredential создает новую запись учётных данных.
// Получает userID из контекста, логирует запрос и результат, возвращает созданные данные в protobuf формате.
func (h *CredentialHandler) CreateCredential(ctx context.Context, req *pb.CreateCredentialRequest) (*pb.CreateCredentialResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	h.logger.Debug("CreateCredential request received",
		zap.String("userID", userID),
		zap.String("title", req.GetCredential().GetTitle()),
		zap.String("login", req.GetCredential().GetLogin()),
	)

	cred := &model.Credential{
		UserID:   userID,
		Title:    req.GetCredential().GetTitle(),
		Login:    req.GetCredential().GetLogin(),
		Password: req.GetCredential().GetPassword(),
		Metadata: req.GetCredential().GetMetadata(),
	}

	err = h.service.Create(ctx, cred)
	if err != nil {
		h.logger.Warn("CreateCredential failed",
			zap.String("userID", userID),
			zap.String("title", req.GetCredential().GetTitle()),
			zap.Error(err),
		)
		return nil, err
	}

	h.logger.Info("CreateCredential succeeded",
		zap.String("userID", userID),
		zap.String("credentialID", cred.ID),
	)

	resp := &pb.CreateCredentialResponse{}
	resp.SetCredential(mapper.CredentialToPB(cred))
	return resp, nil
}

// GetCredentialByID возвращает учётные данные по их уникальному идентификатору.
// Получает userID из контекста, проверяет принадлежность данных пользователю, логирует операции.
func (h *CredentialHandler) GetCredentialByID(ctx context.Context, req *pb.GetCredentialByIDRequest) (*pb.GetCredentialByIDResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	h.logger.Debug("GetCredentialByID request received",
		zap.String("userID", userID),
		zap.String("credentialID", req.GetId()),
	)

	cred, err := h.service.GetByID(ctx, req.GetId())
	if err != nil {
		h.logger.Warn("GetCredentialByID failed",
			zap.String("userID", userID),
			zap.String("credentialID", req.GetId()),
			zap.Error(err),
		)
		return nil, err
	}
	if cred == nil || cred.UserID != userID {
		h.logger.Info("Credential not found",
			zap.String("userID", userID),
			zap.String("credentialID", req.GetId()),
		)
		return nil, status.Error(codes.NotFound, "credential not found")
	}

	h.logger.Info("GetCredentialByID succeeded",
		zap.String("userID", userID),
		zap.String("credentialID", req.GetId()),
	)

	resp := &pb.GetCredentialByIDResponse{}
	resp.SetCredential(mapper.CredentialToPB(cred))
	return resp, nil
}

// GetCredentialsByUserID возвращает все учётные данные пользователя.
// Получает userID из контекста, логирует количество возвращаемых записей.
func (h *CredentialHandler) GetCredentials(ctx context.Context, _ *emptypb.Empty) (*pb.GetCredentialsResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	h.logger.Debug("GetCredentialsByUserID request received",
		zap.String("userID", userID),
	)

	creds, err := h.service.GetByUserID(ctx, userID)
	if err != nil {
		h.logger.Warn("GetCredentialsByUserID failed",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return nil, err
	}

	h.logger.Info("GetCredentialsByUserID succeeded",
		zap.String("userID", userID),
		zap.Int("count", len(creds)),
	)

	resp := &pb.GetCredentialsResponse{}
	for i := range creds {
		resp.SetCredentials(append(resp.GetCredentials(), mapper.CredentialToPB(&creds[i])))
	}

	return resp, nil
}

// UpdateCredential обновляет существующую запись учётных данных.
// Проверяет, что запись принадлежит текущему пользователю, логирует операции.
func (h *CredentialHandler) UpdateCredential(ctx context.Context, req *pb.UpdateCredentialRequest) (*pb.UpdateCredentialResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	h.logger.Debug("UpdateCredential request received",
		zap.String("userID", userID),
		zap.String("credentialID", req.GetCredential().GetId()),
	)

	credProto := req.GetCredential()
	cred := &model.Credential{
		ID:       credProto.GetId(),
		UserID:   userID,
		Title:    credProto.GetTitle(),
		Login:    credProto.GetLogin(),
		Password: credProto.GetPassword(),
		Metadata: credProto.GetMetadata(),
	}

	existing, err := h.service.GetByID(ctx, cred.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil || existing.UserID != userID {
		return nil, status.Error(codes.NotFound, "credential not found")
	}

	err = h.service.Update(ctx, cred)
	if err != nil {
		h.logger.Warn("UpdateCredential failed",
			zap.String("userID", userID),
			zap.String("credentialID", credProto.GetId()),
			zap.Error(err),
		)
		return nil, err
	}

	h.logger.Info("UpdateCredential succeeded",
		zap.String("userID", userID),
		zap.String("credentialID", credProto.GetId()),
	)

	return &pb.UpdateCredentialResponse{}, nil
}

// DeleteCredential удаляет запись учётных данных по идентификатору.
// Проверяет принадлежность записи пользователю, логирует операции.
func (h *CredentialHandler) DeleteCredential(ctx context.Context, req *pb.DeleteCredentialRequest) (*pb.DeleteCredentialResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	h.logger.Debug("DeleteCredential request received",
		zap.String("userID", userID),
		zap.String("credentialID", req.GetId()),
	)

	existing, err := h.service.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if existing == nil || existing.UserID != userID {
		return nil, status.Error(codes.NotFound, "credential not found")
	}

	err = h.service.Delete(ctx, req.GetId())
	if err != nil {
		h.logger.Warn("DeleteCredential failed",
			zap.String("userID", userID),
			zap.String("credentialID", req.GetId()),
			zap.Error(err),
		)
		return nil, err
	}

	h.logger.Info("DeleteCredential succeeded",
		zap.String("userID", userID),
		zap.String("credentialID", req.GetId()),
	)

	return &pb.DeleteCredentialResponse{}, nil
}
