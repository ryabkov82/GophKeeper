package binarydata

import (
	"context"
	"fmt"
	"io"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/mapper"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
)

// BinaryDataManagerIface описывает интерфейс управления бинарными данными.
type BinaryDataManagerIface interface {
	Upload(ctx context.Context, data *model.BinaryData, r io.Reader) error
	Download(ctx context.Context, id string) (io.ReadCloser, error)
	List(ctx context.Context) ([]model.BinaryData, error)
	GetInfo(ctx context.Context, id string) (*model.BinaryData, error)
	CreateInfo(ctx context.Context, data *model.BinaryData) error
	UpdateInfo(ctx context.Context, data *model.BinaryData) error
	Delete(ctx context.Context, id string) error
	SetClient(client pb.BinaryDataServiceClient)
}

// BinaryDataManager управляет CRUD операциями с бинарными данными через gRPC.
type BinaryDataManager struct {
	logger *zap.Logger
	client pb.BinaryDataServiceClient
}

// NewBinaryDataManager создаёт новый BinaryDataManager.
func NewBinaryDataManager(logger *zap.Logger) *BinaryDataManager {
	return &BinaryDataManager{
		logger: logger,
	}
}

// SetClient позволяет установить кастомный (например, моковый) gRPC-клиент.
func (m *BinaryDataManager) SetClient(client pb.BinaryDataServiceClient) {
	m.client = client
}

// Upload загружает бинарные данные на сервер через поток.
func (m *BinaryDataManager) Upload(ctx context.Context, data *model.BinaryData, r io.Reader) error {
	m.logger.Debug("Upload started", zap.String("userID", data.UserID), zap.String("title", data.Title))

	stream, err := m.client.UploadBinaryData(ctx)
	if err != nil {
		return fmt.Errorf("failed to create upload stream: %w", err)
	}

	// Отправляем метаданные первым сообщением
	metaReq := &pb.UploadBinaryDataRequest{}
	metaReq.SetInfo(mapper.BinaryDataToPB(data))
	if err := stream.Send(metaReq); err != nil {
		// сервер мог сразу закрыть поток — заберём статус
		if _, recvErr := stream.CloseAndRecv(); recvErr != nil {
			return fmt.Errorf("server rejected metadata: %w", recvErr)
		}
		return fmt.Errorf("server closed stream after metadata: %w", err)
	}

	buf := make([]byte, 32*1024) // 32 KB чанки
	for {
		n, rerr := r.Read(buf)

		if n > 0 {
			chunkReq := &pb.UploadBinaryDataRequest{}
			chunkReq.SetChunk(buf[:n])
			if serr := stream.Send(chunkReq); serr != nil {
				// ВАЖНО: получаем причину от сервера
				if _, recvErr := stream.CloseAndRecv(); recvErr != nil {
					// здесь будет codes.InvalidArgument / PermissionDenied / ResourceExhausted и т.п.
					return fmt.Errorf("upload aborted by server: %w", recvErr)
				}
				return fmt.Errorf("send failed: %w", serr)
			}
		}
		if rerr == io.EOF {
			break // последний кусок отправили — выходим на CloseAndRecv
		}
		if rerr != nil {
			return fmt.Errorf("failed to read input: %w", rerr)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to receive upload response: %w", err)
	}

	data.ID = resp.GetId()
	m.logger.Info("Upload succeeded", zap.String("binaryDataID", data.ID))
	return nil
}

// CreateInfo сохраняет только метаданные бинарных данных без содержимого.
func (m *BinaryDataManager) CreateInfo(ctx context.Context, data *model.BinaryData) error {
	m.logger.Debug("CreateInfo started", zap.String("title", data.Title))

	req := &pb.SaveBinaryDataInfoRequest{}
	req.SetInfo(mapper.BinaryDataToPB(data))

	resp, err := m.client.SaveBinaryDataInfo(ctx, req)
	if err != nil {
		return fmt.Errorf("SaveBinaryDataInfo RPC failed: %w", err)
	}

	data.ID = resp.GetId()
	m.logger.Info("CreateInfo succeeded", zap.String("binaryDataID", data.ID))
	return nil
}

// UpdateInfo обновляет только метаданные бинарных данных без загрузки содержимого.
func (m *BinaryDataManager) UpdateInfo(ctx context.Context, data *model.BinaryData) error {
	m.logger.Debug("UpdateInfo started", zap.String("binaryDataID", data.ID))

	req := &pb.UpdateBinaryDataRequest{}
	req.SetInfo(mapper.BinaryDataToPB(data))

	resp, err := m.client.UpdateBinaryDataInfo(ctx, req)
	if err != nil {
		return fmt.Errorf("UpdateBinaryDataInfo RPC failed: %w", err)
	}

	data.ID = resp.GetId()
	m.logger.Info("UpdateInfo succeeded", zap.String("binaryDataID", data.ID))
	return nil
}

// Download возвращает поток бинарных данных с сервера.
func (m *BinaryDataManager) Download(ctx context.Context, id string) (io.ReadCloser, error) {
	m.logger.Debug("Download started", zap.String("binaryDataID", id))

	req := &pb.DownloadBinaryDataRequest{}
	req.SetId(id)

	stream, err := m.client.DownloadBinaryData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start download: %w", err)
	}

	// Используем io.Pipe для потоковой передачи данных клиенту
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("failed to receive chunk: %w", err))
				return
			}
			if len(resp.GetChunk()) > 0 {
				if _, err := pw.Write(resp.GetChunk()); err != nil {
					_ = pw.CloseWithError(fmt.Errorf("failed to write to pipe: %w", err))
					return
				}
			}
		}
	}()

	return pr, nil
}

// List возвращает список бинарных данных пользователя.
func (m *BinaryDataManager) List(ctx context.Context) ([]model.BinaryData, error) {
	m.logger.Debug("List started")

	resp, err := m.client.ListBinaryData(ctx, &pb.ListBinaryDataRequest{})
	if err != nil {
		return nil, fmt.Errorf("ListBinaryData RPC failed: %w", err)
	}

	result := make([]model.BinaryData, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		if bd := mapper.BinaryDataFromPB(item); bd != nil {
			result = append(result, *bd)
		}
	}

	m.logger.Info("List succeeded", zap.Int("count", len(result)))
	return result, nil
}

// Delete удаляет бинарные данные по ID.
func (m *BinaryDataManager) Delete(ctx context.Context, id string) error {
	m.logger.Debug("Delete started", zap.String("binaryDataID", id))

	req := &pb.DeleteBinaryDataRequest{}
	req.SetId(id)

	_, err := m.client.DeleteBinaryData(ctx, req)
	if err != nil {
		return fmt.Errorf("DeleteBinaryData RPC failed: %w", err)
	}

	m.logger.Info("Delete succeeded", zap.String("binaryDataID", id))
	return nil
}

// GetInfo возвращает информацию о конкретных бинарных данных по их ID.
func (m *BinaryDataManager) GetInfo(ctx context.Context, id string) (*model.BinaryData, error) {
	m.logger.Debug("GetInfo started", zap.String("binaryDataID", id))

	req := &pb.GetBinaryDataInfoRequest{}
	req.SetId(id)

	resp, err := m.client.GetBinaryDataInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetBinaryDataInfo RPC failed: %w", err)
	}

	info := resp.GetBinaryInfo()
	if info == nil {
		return nil, fmt.Errorf("binary data info is nil")
	}
	data := mapper.BinaryDataFromPB(info)

	m.logger.Info("GetInfo succeeded", zap.String("binaryDataID", data.ID))
	return data, nil
}
