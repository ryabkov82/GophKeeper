package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/ryabkov82/gophkeeper/internal/server/grpc/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// Мок TextDataService
type mockTextDataService struct {
	mock.Mock
}

func (m *mockTextDataService) Create(ctx context.Context, td *model.TextData) error {
	return m.Called(ctx, td).Error(0)
}

func (m *mockTextDataService) GetByID(ctx context.Context, userID, id string) (*model.TextData, error) {
	args := m.Called(ctx, userID, id)
	td := args.Get(0)
	if td == nil {
		return nil, args.Error(1)
	}
	return td.(*model.TextData), args.Error(1)
}

func (m *mockTextDataService) Update(ctx context.Context, td *model.TextData) error {
	return m.Called(ctx, td).Error(0)
}

func (m *mockTextDataService) Delete(ctx context.Context, userID, id string) error {
	return m.Called(ctx, userID, id).Error(0)
}

func (m *mockTextDataService) ListTitles(ctx context.Context, userID string) ([]*model.TextData, error) {
	args := m.Called(ctx, userID)
	data := args.Get(0)
	if data == nil {
		return nil, args.Error(1)
	}
	return data.([]*model.TextData), args.Error(1)
}

func TestCreateTextData_Success(t *testing.T) {
	mockSvc := new(mockTextDataService)
	logger := zap.NewNop()
	h := handlers.NewTextDataHandler(mockSvc, logger)

	userID := uuid.NewString()
	ctx := mockJWTContext(userID)

	req := &pb.CreateTextDataRequest{}
	tdProto := &pb.TextData{}
	tdProto.SetTitle("My Note")
	tdProto.SetContent([]byte("Secret content"))
	tdProto.SetMetadata("{}")
	req.SetTextData(tdProto)

	mockSvc.On("Create", mock.Anything, mock.Anything).Return(nil)

	resp, err := h.CreateTextData(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "My Note", resp.GetTextData().GetTitle())
	assert.Empty(t, resp.GetTextData().GetContent()) // Content не возвращаем
}

func TestCreateTextData_Unauthenticated(t *testing.T) {
	mockSvc := new(mockTextDataService)
	logger := zap.NewNop()
	h := handlers.NewTextDataHandler(mockSvc, logger)

	ctx := context.Background() // нет userID
	req := &pb.CreateTextDataRequest{}
	_, err := h.CreateTextData(ctx, req)
	assert.Error(t, err)
}

func TestGetTextDataByID_Success(t *testing.T) {
	mockSvc := new(mockTextDataService)
	logger := zap.NewNop()
	h := handlers.NewTextDataHandler(mockSvc, logger)

	userID := uuid.NewString()
	tdID := uuid.NewString()
	ctx := mockJWTContext(userID)

	td := &model.TextData{ID: tdID, UserID: userID, Title: "Note", Content: []byte("Secret"), CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mockSvc.On("GetByID", mock.Anything, userID, tdID).Return(td, nil)

	req := &pb.GetTextDataByIDRequest{}
	req.SetId(tdID)

	resp, err := h.GetTextDataByID(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, tdID, resp.GetTextData().GetId())
	assert.Equal(t, "Note", resp.GetTextData().GetTitle())
	assert.Equal(t, "Secret", string(resp.GetTextData().GetContent()))
}

func TestGetTextDataByID_NotFound(t *testing.T) {
	mockSvc := new(mockTextDataService)
	logger := zap.NewNop()
	h := handlers.NewTextDataHandler(mockSvc, logger)

	userID := uuid.NewString()
	tdID := uuid.NewString()
	ctx := mockJWTContext(userID)

	mockSvc.On("GetByID", mock.Anything, userID, tdID).Return(nil, errors.New("text data with id "+tdID+" not found"))

	req := &pb.GetTextDataByIDRequest{}
	req.SetId(tdID)

	_, err := h.GetTextDataByID(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetTextDataTitles_Success(t *testing.T) {
	mockSvc := new(mockTextDataService)
	logger := zap.NewNop()
	h := handlers.NewTextDataHandler(mockSvc, logger)

	userID := uuid.NewString()
	ctx := mockJWTContext(userID)

	// Мокируем возвращаемые данные сервиса
	tds := []*model.TextData{
		{ID: uuid.NewString(), UserID: userID, Title: "Note1"},
		{ID: uuid.NewString(), UserID: userID, Title: "Note2"},
	}

	mockSvc.On("ListTitles", mock.Anything, userID).Return(tds, nil)

	// Вызов хендлера
	resp, err := h.GetTextDataTitles(ctx, &pb.GetTextDataTitlesRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.GetTextDataTitles(), 2)
	assert.Equal(t, "Note1", resp.GetTextDataTitles()[0].GetTitle())
	assert.Equal(t, "Note2", resp.GetTextDataTitles()[1].GetTitle())
}

func TestGetTextDataTitles_ServiceError(t *testing.T) {
	mockSvc := new(mockTextDataService)
	logger := zap.NewNop()
	h := handlers.NewTextDataHandler(mockSvc, logger)

	userID := uuid.NewString()
	ctx := mockJWTContext(userID)

	mockSvc.On("ListTitles", mock.Anything, userID).Return(nil, errors.New("service error"))

	_, err := h.GetTextDataTitles(ctx, &pb.GetTextDataTitlesRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service error")
}

func TestUpdateTextData_Success(t *testing.T) {
	mockSvc := new(mockTextDataService)
	logger := zap.NewNop()
	h := handlers.NewTextDataHandler(mockSvc, logger)

	userID := uuid.NewString()
	tdID := uuid.NewString()
	ctx := mockJWTContext(userID)

	mockSvc.On("Update", mock.Anything, mock.Anything).Return(nil)

	req := &pb.UpdateTextDataRequest{}
	tdProto := &pb.TextData{}
	tdProto.SetId(tdID)
	tdProto.SetTitle("Updated Note")
	req.SetTextData(tdProto)

	_, err := h.UpdateTextData(ctx, req)
	assert.NoError(t, err)

}

func TestDeleteTextData_Success(t *testing.T) {
	mockSvc := new(mockTextDataService)
	logger := zap.NewNop()
	h := handlers.NewTextDataHandler(mockSvc, logger)

	userID := uuid.NewString()
	tdID := uuid.NewString()
	ctx := mockJWTContext(userID)

	mockSvc.On("Delete", mock.Anything, userID, tdID).Return(nil)

	req := &pb.DeleteTextDataRequest{}
	req.SetId(tdID)

	resp, err := h.DeleteTextData(ctx, req)
	assert.NoError(t, err)
	assert.True(t, resp.GetSuccess())
}

func TestDeleteTextData_NotFound(t *testing.T) {
	mockSvc := new(mockTextDataService)
	logger := zap.NewNop()
	h := handlers.NewTextDataHandler(mockSvc, logger)

	userID := uuid.NewString()
	tdID := uuid.NewString()
	ctx := mockJWTContext(userID)

	mockSvc.On("Delete", mock.Anything, userID, tdID).Return(errors.New("text data with id " + tdID + " not found"))

	req := &pb.DeleteTextDataRequest{}
	req.SetId(tdID)

	_, err := h.DeleteTextData(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
