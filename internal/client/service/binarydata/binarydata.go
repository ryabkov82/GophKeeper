package binarydata

import (
	"context"
	"fmt"
	"io"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
)

// BinaryDataManagerIface описывает интерфейс управления бинарными данными.
type BinaryDataManagerIface interface {
	Upload(ctx context.Context, data *model.BinaryData, r io.Reader) error
	Download(ctx context.Context, id string) (io.ReadCloser, error)
	List(ctx context.Context) ([]model.BinaryData, error)
	GetInfo(ctx context.Context, id string) (*model.BinaryData, error)
	Update(ctx context.Context, data *model.BinaryData, r io.Reader) error
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
	metaReq.SetTitle(data.Title)
	metaReq.SetMetadata(data.Metadata)
	if err := stream.Send(metaReq); err != nil {
		return fmt.Errorf("failed to send metadata: %w", err)
	}

	buf := make([]byte, 32*1024) // 32 KB чанки
	for {
		n, err := r.Read(buf)
		if n > 0 {
			chunkReq := &pb.UploadBinaryDataRequest{}
			chunkReq.SetChunk(buf[:n])
			if err := stream.Send(chunkReq); err != nil {
				return fmt.Errorf("failed to send chunk: %w", err)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
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
		result = append(result, model.BinaryData{
			ID:       item.GetId(),
			Title:    item.GetTitle(),
			Metadata: item.GetMetadata(),
		})
	}

	m.logger.Info("List succeeded", zap.Int("count", len(result)))
	return result, nil
}

// Update обновляет бинарные данные на сервере.
func (m *BinaryDataManager) Update(ctx context.Context, data *model.BinaryData, r io.Reader) error {
	m.logger.Debug("Update started", zap.String("binaryDataID", data.ID))

	stream, err := m.client.UpdateBinaryData(ctx)
	if err != nil {
		return fmt.Errorf("failed to create update stream: %w", err)
	}

	// Отправляем метаданные первым сообщением
	metaReq := &pb.UpdateBinaryDataRequest{}
	metaReq.SetId(data.ID)
	metaReq.SetTitle(data.Title)
	metaReq.SetMetadata(data.Metadata)
	if err := stream.Send(metaReq); err != nil {
		return fmt.Errorf("failed to send metadata: %w", err)
	}

	if r != nil {
		buf := make([]byte, 32*1024)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				chunkReq := &pb.UpdateBinaryDataRequest{}
				chunkReq.SetChunk(buf[:n])
				if err := stream.Send(chunkReq); err != nil {
					return fmt.Errorf("failed to send chunk: %w", err)
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to receive update response: %w", err)
	}

	data.ID = resp.GetId()
	m.logger.Info("Update succeeded", zap.String("binaryDataID", data.ID))
	return nil
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

	data := &model.BinaryData{
		ID:       info.GetId(),
		Title:    info.GetTitle(),
		Metadata: info.GetMetadata(),
		Size:     info.GetSize(),
	}

	m.logger.Info("GetInfo succeeded", zap.String("binaryDataID", data.ID))
	return data, nil
}
