package adapters

import (
	"context"
	"fmt"

	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/tui/contracts"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

type contextKey string

const (
	progressChanKey contextKey = "progressChan"
	filePathKey     contextKey = "filePath"
)

// BinaryDataAdapter — адаптер BinaryDataService
type BinaryDataAdapter struct {
	svc contracts.BinaryDataService
}

func NewBinaryDataAdapter(svc contracts.BinaryDataService) *BinaryDataAdapter {
	return &BinaryDataAdapter{svc: svc}
}

func (a *BinaryDataAdapter) List(ctx context.Context) ([]contracts.ListItem, error) {
	dataList, err := a.svc.ListBinaryData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list binary data: %w", err)
	}

	items := make([]contracts.ListItem, 0, len(dataList))
	for _, d := range dataList {
		items = append(items, contracts.ListItem{
			ID:    d.ID,
			Title: d.Metadata,
		})
	}
	return items, nil
}

func (a *BinaryDataAdapter) Get(ctx context.Context, id string) (interface{}, error) {
	data, err := a.svc.GetBinaryDataInfo(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get binary data: %w", err)
	}
	return data, nil
}

func (a *BinaryDataAdapter) Create(ctx context.Context, id string, v interface{}) error {
	data, ok := v.(*model.BinaryData)
	if !ok {
		return fmt.Errorf("invalid type for Create: expected *model.BinaryData, got %T", v)
	}

	progressCh, _ := ctx.Value(progressChanKey).(chan<- app.ProgressMsg)
	filePath, _ := ctx.Value(filePathKey).(string)
	if filePath == "" {
		return fmt.Errorf("filePath missing in context")
	}

	return a.svc.UploadBinaryData(ctx, data, filePath, progressCh)
}

func (a *BinaryDataAdapter) Update(ctx context.Context, id string, v interface{}) error {
	data, ok := v.(*model.BinaryData)
	if !ok {
		return fmt.Errorf("invalid type for Update: expected *model.BinaryData, got %T", v)
	}

	progressCh, _ := ctx.Value(progressChanKey).(chan<- app.ProgressMsg)
	filePath, _ := ctx.Value(filePathKey).(string)
	if filePath == "" {
		return fmt.Errorf("filePath missing in context")
	}

	if data.ID != id {
		data.ID = id
	}

	return a.svc.UpdateBinaryData(ctx, data, filePath, progressCh)
}

func (a *BinaryDataAdapter) Delete(ctx context.Context, id string) error {
	return a.svc.DeleteBinaryData(ctx, id)
}
