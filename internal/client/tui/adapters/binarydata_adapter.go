package adapters

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// BinaryDataAdapter — адаптер BinaryDataService
type BinaryDataAdapter struct {
	svc contracts.BinaryDataService
}

// NewBinaryDataAdapter создаёт адаптер для работы с бинарными данными через слой сервисов.
func NewBinaryDataAdapter(svc contracts.BinaryDataService) *BinaryDataAdapter {
	return &BinaryDataAdapter{svc: svc}
}

// List возвращает список всех бинарных объектов пользователя в виде списка элементов.
func (a *BinaryDataAdapter) List(ctx context.Context) ([]contracts.ListItem, error) {
	dataList, err := a.svc.ListBinaryData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list binary data: %w", err)
	}

	items := make([]contracts.ListItem, 0, len(dataList))
	for _, d := range dataList {
		items = append(items, contracts.ListItem{
			ID:    d.ID,
			Title: d.Title,
		})
	}
	return items, nil
}

// Get получает метаданные бинарных данных по их идентификатору.
func (a *BinaryDataAdapter) Get(ctx context.Context, id string) (interface{}, error) {
	data, err := a.svc.GetBinaryDataInfo(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get binary data: %w", err)
	}
	return data, nil
}

// Create создаёт запись метаданных бинарных данных.
func (a *BinaryDataAdapter) Create(ctx context.Context, v interface{}) error {
	data, ok := v.(*model.BinaryData)
	if !ok {
		return fmt.Errorf("invalid type for Create: expected *model.BinaryData, got %T", v)
	}

	return a.svc.CreateBinaryDataInfo(ctx, data)
}

// Update обновляет метаданные бинарных данных по их идентификатору.
func (a *BinaryDataAdapter) Update(ctx context.Context, id string, v interface{}) error {
	data, ok := v.(*model.BinaryData)
	if !ok {
		return fmt.Errorf("invalid type for Update: expected *model.BinaryData, got %T", v)
	}

	if data.ID != id {
		data.ID = id
	}

	return a.svc.UpdateBinaryDataInfo(ctx, data)
}

// Delete удаляет бинарные данные по идентификатору.
func (a *BinaryDataAdapter) Delete(ctx context.Context, id string) error {
	return a.svc.DeleteBinaryData(ctx, id)
}

// UploadBinaryData загружает бинарный файл на сервер с передачей прогресса.
func (a *BinaryDataAdapter) UploadBinaryData(ctx context.Context, data *model.BinaryData, filePath string, progress chan<- int64) error {
	return a.svc.UploadBinaryData(ctx, data, filePath, progress)
}

// DownloadBinaryData скачивает бинарный файл с сервера в указанное место.
func (a *BinaryDataAdapter) DownloadBinaryData(ctx context.Context, dataID, destPath string, progress chan<- int64) error {
	return a.svc.DownloadBinaryData(ctx, dataID, destPath, progress)
}
