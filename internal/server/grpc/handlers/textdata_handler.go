package handlers

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtauth"
	"github.com/ryabkov82/gophkeeper/internal/pkg/mapper"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TextDataHandler реализует gRPC сервер для TextDataService
type TextDataHandler struct {
	pb.UnimplementedTextDataServiceServer
	service service.TextDataService
	logger  *zap.Logger
}

// NewTextDataHandler создает новый TextDataHandler с внедрением сервиса и логгера.
func NewTextDataHandler(srv service.TextDataService, logger *zap.Logger) *TextDataHandler {
	return &TextDataHandler{
		service: srv,
		logger:  logger,
	}
}

// CreateTextData создает новую запись текстовых данных
func (h *TextDataHandler) CreateTextData(ctx context.Context, req *pb.CreateTextDataRequest) (*pb.CreateTextDataResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	text := &model.TextData{
		UserID:   userID,
		Title:    req.GetTextData().GetTitle(),
		Content:  req.GetTextData().GetContent(),
		Metadata: req.GetTextData().GetMetadata(),
	}

	h.logger.Debug("CreateTextData request received",
		zap.String("userID", userID),
		zap.String("title", text.Title),
	)

	if err := h.service.Create(ctx, text); err != nil {
		h.logger.Warn("CreateTextData failed",
			zap.String("userID", userID),
			zap.String("title", text.Title),
			zap.Error(err),
		)
		return nil, err
	}

	h.logger.Info("CreateTextData succeeded",
		zap.String("userID", userID),
		zap.String("textDataID", text.ID),
	)

	// Возвращаем только ID и Title
	resp := &pb.CreateTextDataResponse{}
	td := mapper.TextDataToPB(text)
	td.SetUserId("")
	td.SetContent(nil)
	td.SetMetadata("")
	td.SetCreatedAt(nil)
	td.SetUpdatedAt(nil)

	resp.SetTextData(td)

	return resp, nil
}

// GetTextDataByID возвращает текстовые данные по ID
func (h *TextDataHandler) GetTextDataByID(ctx context.Context, req *pb.GetTextDataByIDRequest) (*pb.GetTextDataByIDResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	text, err := h.service.GetByID(ctx, userID, req.GetId())
	if err != nil {
		return nil, err
	}
	if text == nil || text.UserID != userID {
		return nil, status.Error(codes.NotFound, "text data not found")
	}

	resp := &pb.GetTextDataByIDResponse{}
	resp.SetTextData(mapper.TextDataToPB(text))
	return resp, nil
}

// GetTextDataTitles возвращает список заголовков текстовых данных пользователя
func (h *TextDataHandler) GetTextDataTitles(ctx context.Context, req *pb.GetTextDataTitlesRequest) (*pb.GetTextDataTitlesResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	titles, err := h.service.ListTitles(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetTextDataTitlesResponse{}
	for _, t := range titles {
		td := mapper.TextDataToPB(t)
		td.SetUserId("")
		td.SetContent(nil)
		td.SetMetadata("")
		td.SetCreatedAt(nil)
		td.SetUpdatedAt(nil)
		resp.SetTextDataTitles(append(resp.GetTextDataTitles(), td))
	}
	return resp, nil
}

// UpdateTextData обновляет текстовые данные
func (h *TextDataHandler) UpdateTextData(ctx context.Context, req *pb.UpdateTextDataRequest) (*pb.UpdateTextDataResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	data := &model.TextData{
		ID:       req.GetTextData().GetId(),
		UserID:   userID,
		Title:    req.GetTextData().GetTitle(),
		Content:  req.GetTextData().GetContent(),
		Metadata: req.GetTextData().GetMetadata(),
	}

	err = h.service.Update(ctx, data)
	if err != nil {
		return nil, err
	}

	resp := &pb.UpdateTextDataResponse{}
	resp.SetSuccess(true)
	return resp, nil
}

// DeleteTextData удаляет текстовые данные по ID
func (h *TextDataHandler) DeleteTextData(ctx context.Context, req *pb.DeleteTextDataRequest) (*pb.DeleteTextDataResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	if err := h.service.Delete(ctx, userID, req.GetId()); err != nil {
		if err.Error() == fmt.Sprintf("text data with id %s not found", req.GetId()) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, err
	}

	resp := &pb.DeleteTextDataResponse{}
	resp.SetSuccess(true)
	return resp, nil
}
