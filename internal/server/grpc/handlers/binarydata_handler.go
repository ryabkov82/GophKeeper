package handlers

import (
	"context"
	"fmt"
	"io"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtauth"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BinaryDataHandler реализует gRPC сервер для BinaryDataService
type BinaryDataHandler struct {
	pb.UnimplementedBinaryDataServiceServer
	binarySvc service.BinaryDataService
	logger    *zap.Logger
}

// NewBinaryDataHandler создает новый BinaryDataHandler с внедрением сервиса и логгера.
func NewBinaryDataHandler(srv service.BinaryDataService, logger *zap.Logger) *BinaryDataHandler {
	return &BinaryDataHandler{
		binarySvc: srv,
		logger:    logger,
	}
}

// UploadBinaryData загружает бинарные данные через поток
func (h *BinaryDataHandler) UploadBinaryData(stream pb.BinaryDataService_UploadBinaryDataServer) error {
	// Получаем userID из контекста JWT
	userID, err := jwtauth.FromContext(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "userID not found in context")
	}

	// Читаем первый пакет, который должен содержать метаданные
	req, err := stream.Recv()
	if err != nil {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("failed to receive initial message: %v", err))
	}

	title := req.GetInfo().GetTitle()
	metadata := req.GetInfo().GetMetadata()
	clientPath := req.GetInfo().GetClientPath()

	data := &model.BinaryData{
		UserID:     userID,
		Title:      title,
		Metadata:   metadata,
		ClientPath: clientPath,
	}

	h.logger.Debug("UploadBinaryData started",
		zap.String("userID", userID),
		zap.String("title", title),
	)

	pr, pw := io.Pipe()

	// Запускаем горутину для асинхронного получения чанков
	go func() {
		defer pw.Close()
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				_ = pw.CloseWithError(err)
				return
			}
			chunk := req.GetChunk()
			if len(chunk) > 0 {
				if _, err := pw.Write(chunk); err != nil {
					_ = pw.CloseWithError(err)
					return
				}
			}
		}
	}()

	// Создаем запись в сервисе
	data, err = h.binarySvc.Create(stream.Context(), data, pr)
	if err != nil {
		h.logger.Warn("UploadBinaryData failed", zap.String("userID", userID), zap.String("title", title), zap.Error(err))
		return err
	}

	h.logger.Info("UploadBinaryData succeeded", zap.String("userID", userID), zap.String("binaryDataID", data.ID))

	resp := &pb.UploadBinaryDataResponse{}
	resp.SetId(data.ID)
	return stream.SendAndClose(resp)
}

// UpdateBinaryData обновляет существующую запись бинарных данных через поток
func (h *BinaryDataHandler) UpdateBinaryData(stream pb.BinaryDataService_UpdateBinaryDataServer) error {
	// Получаем userID из контекста JWT
	userID, err := jwtauth.FromContext(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "userID not found in context")
	}

	// Читаем первый пакет, который должен содержать ID записи и новые метаданные
	req, err := stream.Recv()
	if err != nil {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("failed to receive initial message: %v", err))
	}

	id := req.GetInfo().GetId()
	title := req.GetInfo().GetTitle()
	metadata := req.GetInfo().GetMetadata()
	clientPath := req.GetInfo().GetClientPath()

	data := &model.BinaryData{
		ID:         id,
		UserID:     userID,
		Title:      title,
		Metadata:   metadata,
		ClientPath: clientPath,
	}

	h.logger.Debug("UpdateBinaryData started",
		zap.String("userID", userID),
		zap.String("title", title),
	)

	pr, pw := io.Pipe()

	// Горрутина для асинхронного чтения чанков
	go func() {
		defer pw.Close()
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				_ = pw.CloseWithError(err)
				return
			}
			chunk := req.GetChunk()
			if len(chunk) > 0 {
				if _, err := pw.Write(chunk); err != nil {
					_ = pw.CloseWithError(err)
					return
				}
			}
		}
	}()

	// Вызываем сервис для обновления записи
	data, err = h.binarySvc.Update(stream.Context(), data, pr)
	if err != nil {
		h.logger.Warn("UpdateBinaryData failed",
			zap.String("userID", userID),
			zap.String("title", title),
			zap.Error(err),
		)
		return err
	}

	h.logger.Info("UpdateBinaryData succeeded",
		zap.String("userID", userID),
		zap.String("binaryDataID", data.ID),
	)

	resp := &pb.UpdateBinaryDataResponse{}
	resp.SetId(data.ID)
	return stream.SendAndClose(resp)
}

// UpdateBinaryDataInfo обновляет только метаданные бинарных данных
func (h *BinaryDataHandler) UpdateBinaryDataInfo(ctx context.Context, req *pb.UpdateBinaryDataRequest) (*pb.UpdateBinaryDataResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	data := &model.BinaryData{
		ID:       req.GetInfo().GetId(),
		UserID:   userID,
		Title:    req.GetInfo().GetTitle(),
		Metadata: req.GetInfo().GetMetadata(),
	}

	data, err = h.binarySvc.UpdateInfo(ctx, data)
	if err != nil {
		h.logger.Warn("UpdateBinaryDataInfo failed",
			zap.String("userID", userID),
			zap.String("id", req.GetInfo().GetId()),
			zap.Error(err),
		)
		return nil, err
	}

	resp := &pb.UpdateBinaryDataResponse{}
	resp.SetId(data.ID)

	h.logger.Info("UpdateBinaryDataInfo succeeded",
		zap.String("userID", userID),
		zap.String("binaryDataID", data.ID),
	)

	return resp, nil
}

// DownloadBinaryData возвращает бинарные данные пользователю
func (h *BinaryDataHandler) DownloadBinaryData(
	req *pb.DownloadBinaryDataRequest,
	stream pb.BinaryDataService_DownloadBinaryDataServer,
) error {
	userID, err := jwtauth.FromContext(stream.Context())
	if err != nil {
		return status.Error(codes.Unauthenticated, "userID not found in context")
	}

	data, reader, err := h.binarySvc.Get(stream.Context(), userID, req.GetId())
	if err != nil {
		return err
	}
	defer reader.Close()

	h.logger.Debug("DownloadBinaryData started", zap.String("userID", userID), zap.String("id", data.ID))

	buf := make([]byte, 32*1024) // 32KB
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return status.Errorf(codes.Internal, "failed to read file: %v", err)
		}
		if n == 0 {
			break
		}

		resp := &pb.DownloadBinaryDataResponse{}
		resp.SetChunk(buf[:n])

		if err := stream.Send(resp); err != nil {
			return status.Errorf(codes.Internal, "failed to send chunk: %v", err)
		}

		if err == io.EOF {
			break
		}
	}

	h.logger.Info("DownloadBinaryData completed", zap.String("userID", userID), zap.String("id", data.ID))
	return nil
}

// ListBinaryData возвращает список бинарных данных пользователя
func (h *BinaryDataHandler) ListBinaryData(ctx context.Context, req *pb.ListBinaryDataRequest) (*pb.ListBinaryDataResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	list, err := h.binarySvc.List(ctx, userID)
	if err != nil {
		h.logger.Warn("ListBinaryData failed", zap.String("userID", userID), zap.Error(err))
		return nil, err
	}

	resp := &pb.ListBinaryDataResponse{}
	for _, d := range list {
		item := &pb.BinaryDataInfo{}
		item.SetId(d.ID)
		item.SetTitle(d.Title)
		//item.SetMetadata(d.Metadata)
		item.SetClientPath(d.ClientPath)
		resp.SetItems(append(resp.GetItems(), item))
	}

	h.logger.Info("ListBinaryData succeeded", zap.String("userID", userID), zap.Int("items", len(list)))
	return resp, nil
}

// GetBinaryDataInfo возвращает информацию о конкретных бинарных данных
func (h *BinaryDataHandler) GetBinaryDataInfo(ctx context.Context, req *pb.GetBinaryDataInfoRequest) (*pb.GetBinaryDataInfoResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	data, err := h.binarySvc.GetInfo(ctx, userID, req.GetId())
	if err != nil {
		h.logger.Warn("GetBinaryDataInfo failed",
			zap.String("userID", userID),
			zap.String("binaryDataID", req.GetId()),
			zap.Error(err),
		)
		return nil, err
	}

	info := &pb.BinaryDataInfo{}
	info.SetId(data.ID)
	info.SetTitle(data.Title)
	info.SetMetadata(data.Metadata)
	info.SetSize(data.Size)
	info.SetClientPath(data.ClientPath)

	resp := &pb.GetBinaryDataInfoResponse{}
	resp.SetBinaryInfo(info)

	h.logger.Info("GetBinaryDataInfo succeeded",
		zap.String("userID", userID),
		zap.String("binaryDataID", data.ID),
	)

	return resp, nil
}

// SaveBinaryDataInfo создает метаданные бинарных данных без их содержимого
func (h *BinaryDataHandler) SaveBinaryDataInfo(ctx context.Context, req *pb.SaveBinaryDataInfoRequest) (*pb.SaveBinaryDataInfoResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	info := req.GetInfo()
	if info == nil {
		return nil, status.Error(codes.InvalidArgument, "info is required")
	}

	data := &model.BinaryData{
		UserID:     userID,
		Title:      info.GetTitle(),
		Metadata:   info.GetMetadata(),
		ClientPath: info.GetClientPath(),
	}

	var res *model.BinaryData
	res, err = h.binarySvc.CreateInfo(ctx, data)

	if err != nil {
		h.logger.Warn("SaveBinaryDataInfo failed",
			zap.String("userID", userID),
			zap.String("binaryDataID", data.ID),
			zap.Error(err),
		)
		return nil, err
	}

	h.logger.Info("SaveBinaryDataInfo succeeded",
		zap.String("userID", userID),
		zap.String("binaryDataID", res.ID),
	)

	resp := &pb.SaveBinaryDataInfoResponse{}
	resp.SetId(res.ID)
	return resp, nil
}

// DeleteBinaryData удаляет запись бинарных данных
func (h *BinaryDataHandler) DeleteBinaryData(ctx context.Context, req *pb.DeleteBinaryDataRequest) (*pb.DeleteBinaryDataResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	if err := h.binarySvc.Delete(ctx, userID, req.GetId()); err != nil {
		h.logger.Warn("DeleteBinaryData failed", zap.String("userID", userID), zap.String("binaryDataID", req.GetId()), zap.Error(err))
		return nil, err
	}

	h.logger.Info("DeleteBinaryData succeeded", zap.String("userID", userID), zap.String("binaryDataID", req.GetId()))
	resp := &pb.DeleteBinaryDataResponse{}
	return resp, nil
}
