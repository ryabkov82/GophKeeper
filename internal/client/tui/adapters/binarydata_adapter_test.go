package adapters_test

import (
	"context"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/tui/adapters"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBinaryDataService — мок для BinaryDataService
type MockBinaryDataService struct {
	mock.Mock
}

func (m *MockBinaryDataService) UploadBinaryData(ctx context.Context, data *model.BinaryData, filePath string, progressChan chan<- int64) error {
	return nil
}

func (m *MockBinaryDataService) UpdateBinaryData(ctx context.Context, data *model.BinaryData, filePath string, progressChan chan<- int64) error {
	return nil
}

func (m *MockBinaryDataService) DownloadBinaryData(ctx context.Context, dataID, destPath string, progressCh chan<- int64) error {
	return nil
}

func (m *MockBinaryDataService) GetBinaryDataInfo(ctx context.Context, id string) (*model.BinaryData, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.BinaryData), args.Error(1)
}

func (m *MockBinaryDataService) ListBinaryData(ctx context.Context) ([]model.BinaryData, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.BinaryData), args.Error(1)
}

func (m *MockBinaryDataService) DeleteBinaryData(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBinaryDataService) CreateBinaryDataInfo(ctx context.Context, data *model.BinaryData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockBinaryDataService) UpdateBinaryDataInfo(ctx context.Context, data *model.BinaryData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func TestBinaryDataAdapter_List(t *testing.T) {
	mockSvc := new(MockBinaryDataService)
	adapter := adapters.NewBinaryDataAdapter(mockSvc)

	dataList := []model.BinaryData{
		{ID: "1", Title: "Data 1"},
		{ID: "2", Title: "Data 2"},
	}

	mockSvc.On("ListBinaryData", mock.Anything).Return(dataList, nil)

	items, err := adapter.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Data 1", items[0].Title)
	mockSvc.AssertExpectations(t)
}

func TestBinaryDataAdapter_Get(t *testing.T) {
	mockSvc := new(MockBinaryDataService)
	adapter := adapters.NewBinaryDataAdapter(mockSvc)

	data := &model.BinaryData{ID: "1", Metadata: "Meta"}
	mockSvc.On("GetBinaryDataInfo", mock.Anything, "1").Return(data, nil)

	got, err := adapter.Get(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, data, got)
	mockSvc.AssertExpectations(t)
}

func TestBinaryDataAdapter_Create(t *testing.T) {
	mockSvc := new(MockBinaryDataService)
	adapter := adapters.NewBinaryDataAdapter(mockSvc)

	data := &model.BinaryData{ID: "1", Metadata: "New"}
	mockSvc.On("CreateBinaryDataInfo", mock.Anything, data).Return(nil)

	err := adapter.Create(context.Background(), data)
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestBinaryDataAdapter_Update(t *testing.T) {
	mockSvc := new(MockBinaryDataService)
	adapter := adapters.NewBinaryDataAdapter(mockSvc)

	data := &model.BinaryData{ID: "1", Metadata: "Updated"}
	mockSvc.On("UpdateBinaryDataInfo", mock.Anything, data).Return(nil)

	err := adapter.Update(context.Background(), "1", data)
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestBinaryDataAdapter_Delete(t *testing.T) {
	mockSvc := new(MockBinaryDataService)
	adapter := adapters.NewBinaryDataAdapter(mockSvc)

	mockSvc.On("DeleteBinaryData", mock.Anything, "1").Return(nil)

	err := adapter.Delete(context.Background(), "1")
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestBinaryDataAdapter_Create_InvalidType(t *testing.T) {
	mockSvc := new(MockBinaryDataService)
	adapter := adapters.NewBinaryDataAdapter(mockSvc)

	err := adapter.Create(context.Background(), "not binary data")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected *model.BinaryData")
}

func TestBinaryDataAdapter_Update_InvalidType(t *testing.T) {
	mockSvc := new(MockBinaryDataService)
	adapter := adapters.NewBinaryDataAdapter(mockSvc)

	err := adapter.Update(context.Background(), "1", "not binary data")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected *model.BinaryData")
}

func TestBinaryDataAdapter_Update_IDMismatch(t *testing.T) {
	mockSvc := new(MockBinaryDataService)
	adapter := adapters.NewBinaryDataAdapter(mockSvc)

	data := &model.BinaryData{ID: "2", Metadata: "Some"}
	mockSvc.On("UpdateBinaryDataInfo", mock.Anything, mock.MatchedBy(func(d *model.BinaryData) bool {
		return d.ID == "1"
	})).Return(nil)

	err := adapter.Update(context.Background(), "1", data)
	assert.NoError(t, err)
	assert.Equal(t, "1", data.ID)
	mockSvc.AssertExpectations(t)
}
