package handlers

import (
	"context"
	"fmt"
	"io"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtauth"
	"github.com/ryabkov82/gophkeeper/internal/pkg/mapper"
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

	data := mapper.BinaryDataFromPB(req.GetInfo())
	if data == nil {
		return status.Error(codes.InvalidArgument, "info is required")
	}
	data.UserID = userID

	h.logger.Debug("UploadBinaryData started",
		zap.String("userID", userID),
		zap.String("title", data.Title),
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

	if data.ID == "" {
		// Создаем запись в сервисе
		data, err = h.binarySvc.Create(stream.Context(), data, pr)
	} else {
		// обновляем запись в сервисе
		data, err = h.binarySvc.Update(stream.Context(), data, pr)
	}
	if err != nil {
		h.logger.Warn("UploadBinaryData failed", zap.String("userID", userID), zap.String("title", data.Title), zap.Error(err))
		return err
	}

	h.logger.Info("UploadBinaryData succeeded", zap.String("userID", userID), zap.String("binaryDataID", data.ID))

	resp := &pb.UploadBinaryDataResponse{}
	resp.SetId(data.ID)
	return stream.SendAndClose(resp)
}

// UpdateBinaryDataInfo обновляет только метаданные бинарных данных
func (h *BinaryDataHandler) UpdateBinaryDataInfo(ctx context.Context, req *pb.UpdateBinaryDataRequest) (*pb.UpdateBinaryDataResponse, error) {
	userID, err := jwtauth.FromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userID not found in context")
	}

	data := mapper.BinaryDataFromPB(req.GetInfo())
	if data == nil {
		return nil, status.Error(codes.InvalidArgument, "info is required")
	}
	data.UserID = userID

	data, err = h.binarySvc.UpdateInfo(ctx, data)
	if err != nil {
		h.logger.Warn("UpdateBinaryDataInfo failed",
			zap.String("userID", userID),
			zap.String("id", data.ID),
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
		d.Metadata = ""
		resp.SetItems(append(resp.GetItems(), mapper.BinaryDataToPB(d)))
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

	resp := &pb.GetBinaryDataInfoResponse{}
	resp.SetBinaryInfo(mapper.BinaryDataToPB(data))

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
	data := mapper.BinaryDataFromPB(info)
	if data == nil {
		return nil, status.Error(codes.InvalidArgument, "info is required")
	}
	data.UserID = userID

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
