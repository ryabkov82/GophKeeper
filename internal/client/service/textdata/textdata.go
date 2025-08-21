package textdata

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"go.uber.org/zap"
)

// MaxContentSize определяет максимальный размер содержимого текстовых данных в байтах (10 MB).
const MaxContentSize = 10 * 1024 * 1024 // 10 MB

// TextDataManagerIface описывает интерфейс клиента для работы с текстовыми данными.
type TextDataManagerIface interface {
	// CreateTextData создает новую текстовую запись на сервере.
	CreateTextData(ctx context.Context, data *model.TextData) error

	// GetTextDataByID получает текстовую запись по её идентификатору.
	GetTextDataByID(ctx context.Context, id string) (*model.TextData, error)

	// GetTextDataTitles возвращает список всех заголовков текстовых данных пользователя.
	GetTextDataTitles(ctx context.Context) ([]*model.TextData, error)

	// UpdateTextData обновляет текстовую запись на сервере.
	UpdateTextData(ctx context.Context, data *model.TextData) error

	// DeleteTextData удаляет текстовую запись по её идентификатору.
	DeleteTextData(ctx context.Context, id string) error

	// SetClient позволяет установить кастомный gRPC-клиент (например, мок для тестов).
	SetClient(client pb.TextDataServiceClient)
}

// TextDataManager управляет CRUD-операциями с текстовыми данными через gRPC.
type TextDataManager struct {
	logger *zap.Logger
	client pb.TextDataServiceClient
}

// NewTextDataManager создает новый экземпляр TextDataManager с указанным логгером.
func NewTextDataManager(logger *zap.Logger) *TextDataManager {
	return &TextDataManager{logger: logger}
}

// SetClient задает gRPC-клиент для взаимодействия с сервером.
func (m *TextDataManager) SetClient(client pb.TextDataServiceClient) {
	m.client = client
}

// CreateTextData создает новую текстовую запись на сервере.
// Проверяет, что размер Content не превышает MaxContentSize.
func (m *TextDataManager) CreateTextData(ctx context.Context, data *model.TextData) error {
	if len(data.Content) > MaxContentSize {
		return fmt.Errorf("content too large: %d bytes, max %d bytes", len(data.Content), MaxContentSize)
	}

	req := &pb.CreateTextDataRequest{}
	req.SetTextData(toProtoTextData(data))

	_, err := m.client.CreateTextData(ctx, req)
	if err != nil {
		m.logger.Error("CreateTextData RPC failed", zap.Error(err))
		return fmt.Errorf("CreateTextData RPC failed: %w", err)
	}

	m.logger.Info("CreateTextData succeeded", zap.String("textDataID", data.ID))
	return nil
}

// GetTextDataByID получает текстовую запись по её ID.
func (m *TextDataManager) GetTextDataByID(ctx context.Context, id string) (*model.TextData, error) {
	req := &pb.GetTextDataByIDRequest{}
	req.SetId(id)

	resp, err := m.client.GetTextDataByID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetTextDataByID RPC failed: %w", err)
	}

	return fromProtoTextData(resp.GetTextData()), nil
}

// GetTextDataTitles возвращает список всех текстовых данных пользователя с их ID и Title.
// Content и Metadata не возвращаются для экономии трафика.
func (m *TextDataManager) GetTextDataTitles(ctx context.Context) ([]*model.TextData, error) {
	resp, err := m.client.GetTextDataTitles(ctx, &pb.GetTextDataTitlesRequest{})
	if err != nil {
		m.logger.Error("GetTextDataTitles RPC failed", zap.Error(err))
		return nil, fmt.Errorf("GetTextDataTitles RPC failed: %w", err)
	}

	result := make([]*model.TextData, 0, len(resp.GetTextDataTitles()))
	for _, td := range resp.GetTextDataTitles() {
		result = append(result, &model.TextData{
			ID:     td.GetId(),
			UserID: td.GetUserId(),
			Title:  td.GetTitle(),
		})
	}

	m.logger.Info("GetTextDataTitles succeeded",
		zap.Int("count", len(result)),
	)
	return result, nil
}

// UpdateTextData обновляет текстовую запись на сервере.
// Проверяет, что размер Content не превышает MaxContentSize.
func (m *TextDataManager) UpdateTextData(ctx context.Context, data *model.TextData) error {
	if len(data.Content) > MaxContentSize {
		return fmt.Errorf("content too large: %d bytes, max %d bytes", len(data.Content), MaxContentSize)
	}

	req := &pb.UpdateTextDataRequest{}
	req.SetTextData(toProtoTextData(data))

	_, err := m.client.UpdateTextData(ctx, req)
	if err != nil {
		return fmt.Errorf("UpdateTextData RPC failed: %w", err)
	}

	return nil
}

// DeleteTextData удаляет текстовую запись по её ID.
func (m *TextDataManager) DeleteTextData(ctx context.Context, id string) error {
	req := &pb.DeleteTextDataRequest{}
	req.SetId(id)

	_, err := m.client.DeleteTextData(ctx, req)
	if err != nil {
		return fmt.Errorf("DeleteTextData RPC failed: %w", err)
	}

	return nil
}

// Преобразование model.TextData -> pb.TextData
func toProtoTextData(td *model.TextData) *pb.TextData {
	pbtd := &pb.TextData{}
	pbtd.SetId(td.ID)
	pbtd.SetUserId(td.UserID)
	pbtd.SetTitle(td.Title)
	pbtd.SetContent(td.Content)
	pbtd.SetMetadata(td.Metadata)
	return pbtd
}

// Преобразование pb.TextData -> model.TextData
func fromProtoTextData(pbtd *pb.TextData) *model.TextData {
	return &model.TextData{
		ID:       pbtd.GetId(),
		UserID:   pbtd.GetUserId(),
		Title:    pbtd.GetTitle(),
		Content:  pbtd.GetContent(),
		Metadata: pbtd.GetMetadata(),
	}
}
