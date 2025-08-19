package handlers_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtauth"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"github.com/ryabkov82/gophkeeper/internal/server/grpc/handlers"
)

// --- Мок BinaryDataService ---
type mockBinaryDataService struct {
	mock.Mock
	Received []byte // сюда запишем, что реально пришло
}

func (m *mockBinaryDataService) Create(ctx context.Context, data *model.BinaryData, r io.Reader) (*model.BinaryData, error) {
	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, r)   // читаем все данные
	m.Received = buf.Bytes() // сохраняем для проверки
	return &model.BinaryData{
		ID:          "123",
		UserID:      data.UserID,
		Title:       data.Title,
		StoragePath: "user123/123.bin",
		Metadata:    data.Metadata,
	}, nil
}

func (m *mockBinaryDataService) CreateInfo(ctx context.Context, data *model.BinaryData) (*model.BinaryData, error) {
	return &model.BinaryData{ID: "new-id", UserID: data.UserID, Title: data.Title, Metadata: data.Metadata, ClientPath: data.ClientPath}, nil
}

func (m *mockBinaryDataService) UpdateInfo(ctx context.Context, data *model.BinaryData) (*model.BinaryData, error) {
	return data, nil
}

func (m *mockBinaryDataService) Update(ctx context.Context, data *model.BinaryData, r io.Reader) (*model.BinaryData, error) {
	buf := new(bytes.Buffer)
	if r != nil {
		_, _ = io.Copy(buf, r)   // читаем все данные
		m.Received = buf.Bytes() // сохраняем для проверки
	}

	return &model.BinaryData{
		ID:          data.ID,
		UserID:      data.UserID,
		Title:       data.Title,
		StoragePath: "user123/" + data.ID + ".bin", // имитация нового пути
		Metadata:    data.Metadata,
	}, nil
}

func (m *mockBinaryDataService) Get(ctx context.Context, userID, id string) (*model.BinaryData, io.ReadCloser, error) {
	data := &model.BinaryData{
		ID:       id,
		UserID:   userID,
		Title:    "FileTitle",
		Metadata: "meta",
	}
	content := bytes.NewReader([]byte("hello world"))
	return data, io.NopCloser(content), nil
}

func (m *mockBinaryDataService) GetInfo(ctx context.Context, userID, id string) (*model.BinaryData, error) {
	args := m.Called(ctx, userID, id)
	if v := args.Get(0); v != nil {
		return v.(*model.BinaryData), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockBinaryDataService) List(ctx context.Context, userID string) ([]*model.BinaryData, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*model.BinaryData), args.Error(1)
}

func (m *mockBinaryDataService) Delete(ctx context.Context, userID, id string) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *mockBinaryDataService) Close() {}

// --- Вспомогательный контекст ---
func ctxWithUserID(userID string) context.Context {
	return jwtauth.WithUserID(context.Background(), userID)
}

// --- Тест UploadBinaryData ---
func TestBinaryDataHandler_UploadBinaryData(t *testing.T) {
	mockSvc := &mockBinaryDataService{}
	logger := zap.NewNop()
	handler := handlers.NewBinaryDataHandler(mockSvc, logger)

	userID := "user123"
	title := "MyFile"
	metadata := "meta"
	clientPath := "/tmp/file.txt"
	content := []byte("hello world")

	dataInfo := &pb.BinaryDataInfo{}
	dataInfo.SetTitle(title)
	dataInfo.SetMetadata(metadata)
	dataInfo.SetClientPath(clientPath)
	// Мок stream
	stream := &mockUploadStream{
		ctx: ctxWithUserID(userID),
		recvMsgs: []*pb.UploadBinaryDataRequest{
			func() *pb.UploadBinaryDataRequest {
				r := &pb.UploadBinaryDataRequest{}
				r.SetInfo(dataInfo)
				return r
			}(),
			func() *pb.UploadBinaryDataRequest {
				r := &pb.UploadBinaryDataRequest{}
				r.SetChunk(content)
				return r
			}(),
		},
	}

	err := handler.UploadBinaryData(stream)
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
	require.Equal(t, content, mockSvc.Received)
}

// --- Тест UpdateBinaryData ---
func TestBinaryDataHandler_UpdateBinaryData(t *testing.T) {
	mockSvc := &mockBinaryDataService{}
	logger := zap.NewNop()
	handler := handlers.NewBinaryDataHandler(mockSvc, logger)

	userID := "user123"
	dataID := "data123"
	newTitle := "UpdatedFile"
	newMeta := "newMeta"
	clientPath := "/tmp/newfile.txt"
	content := []byte("new content")

	dataInfo := &pb.BinaryDataInfo{}
	dataInfo.SetId(dataID)
	dataInfo.SetTitle(newTitle)
	dataInfo.SetMetadata(newMeta)
	dataInfo.SetClientPath(clientPath)

	// Мок stream
	stream := &mockUpdateStream{
		ctx: ctxWithUserID(userID),
		recvMsgs: []*pb.UpdateBinaryDataRequest{
			func() *pb.UpdateBinaryDataRequest {
				r := &pb.UpdateBinaryDataRequest{}
				r.SetInfo(dataInfo)
				return r
			}(),
			func() *pb.UpdateBinaryDataRequest {
				r := &pb.UpdateBinaryDataRequest{}
				r.SetChunk(content)
				return r
			}(),
		},
	}

	err := handler.UpdateBinaryData(stream)
	assert.NoError(t, err)
	require.Equal(t, content, mockSvc.Received)
	require.Equal(t, dataID, stream.sentResp.GetId())
}

// --- Тест UpdateBinaryDataInfo ---
func TestBinaryDataHandler_UpdateBinaryDataInfo(t *testing.T) {
	mockSvc := &mockBinaryDataService{}
	logger := zap.NewNop()
	handler := handlers.NewBinaryDataHandler(mockSvc, logger)

	userID := "user123"
	dataID := "data123"
	newTitle := "UpdatedFile"
	newMeta := "newMeta"

	req := &pb.UpdateBinaryDataRequest{}
	dataInfo := &pb.BinaryDataInfo{}
	dataInfo.SetId(dataID)
	dataInfo.SetTitle(newTitle)
	dataInfo.SetMetadata(newMeta)

	req.SetInfo(dataInfo)

	resp, err := handler.UpdateBinaryDataInfo(ctxWithUserID(userID), req)
	assert.NoError(t, err)
	assert.Equal(t, dataID, resp.GetId())
	mockSvc.AssertExpectations(t)
}

// --- Тест DownloadBinaryData ---
func TestBinaryDataHandler_DownloadBinaryData(t *testing.T) {
	logger := zap.NewNop()
	userID := "user123"
	dataID := "data123"

	mockSvc := &mockBinaryDataService{}
	handler := handlers.NewBinaryDataHandler(mockSvc, logger)

	stream := &mockDownloadStream{
		ctx: ctxWithUserID(userID),
		id:  dataID,
	}

	req := &pb.DownloadBinaryDataRequest{}
	req.SetId(dataID)

	err := handler.DownloadBinaryData(req, stream)
	assert.NoError(t, err)

	// Проверяем, что все данные получены
	allData := bytes.Join(stream.sentChunks, nil)
	assert.Equal(t, []byte("hello world"), allData)
}

// --- Тест ListBinaryData ---
func TestBinaryDataHandler_ListBinaryData(t *testing.T) {
	mockSvc := &mockBinaryDataService{}
	logger := zap.NewNop()
	handler := handlers.NewBinaryDataHandler(mockSvc, logger)

	userID := "user123"
	mockSvc.On("List", mock.Anything, userID).Return([]*model.BinaryData{
		{ID: "id1", Title: "t1", Metadata: "m1"},
		{ID: "id2", Title: "t2", Metadata: "m2"},
	}, nil)

	resp, err := handler.ListBinaryData(ctxWithUserID(userID), &pb.ListBinaryDataRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.GetItems(), 2)
	assert.Equal(t, "id1", resp.GetItems()[0].GetId())
	mockSvc.AssertExpectations(t)
}

func TestBinaryDataHandler_GetBinaryDataInfo_Success(t *testing.T) {
	mockSvc := &mockBinaryDataService{}
	logger := zap.NewNop()
	handler := handlers.NewBinaryDataHandler(mockSvc, logger)

	userID := "user123"
	dataID := "id123"
	bd := &model.BinaryData{
		ID:       dataID,
		Title:    "My File",
		Metadata: "some meta",
		Size:     1024,
	}

	// Мокаем GetInfo
	mockSvc.On("GetInfo", mock.Anything, userID, dataID).Return(bd, nil).Once()

	req := &pb.GetBinaryDataInfoRequest{}
	req.SetId(dataID)
	resp, err := handler.GetBinaryDataInfo(ctxWithUserID(userID), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, dataID, resp.GetBinaryInfo().GetId())
	assert.Equal(t, "My File", resp.GetBinaryInfo().GetTitle())
	assert.Equal(t, "some meta", resp.GetBinaryInfo().GetMetadata())
	assert.EqualValues(t, 1024, resp.GetBinaryInfo().GetSize())

	mockSvc.AssertExpectations(t)
}

func TestBinaryDataHandler_SaveBinaryDataInfo_Create(t *testing.T) {
	mockSvc := &mockBinaryDataService{}
	logger := zap.NewNop()
	handler := handlers.NewBinaryDataHandler(mockSvc, logger)

	userID := "user123"

	info := &pb.BinaryDataInfo{}
	info.SetTitle("InfoTitle")
	info.SetMetadata("meta")
	info.SetClientPath("/tmp/info.txt")

	req := &pb.SaveBinaryDataInfoRequest{}
	req.SetInfo(info)

	resp, err := handler.SaveBinaryDataInfo(ctxWithUserID(userID), req)
	assert.NoError(t, err)
	assert.Equal(t, "new-id", resp.GetId())
}

// --- Тест DeleteBinaryData ---
func TestBinaryDataHandler_DeleteBinaryData(t *testing.T) {
	mockSvc := &mockBinaryDataService{}
	logger := zap.NewNop()
	handler := handlers.NewBinaryDataHandler(mockSvc, logger)

	userID := "user123"
	mockSvc.On("Delete", mock.Anything, userID, "file123").Return(nil)

	req := &pb.DeleteBinaryDataRequest{}
	req.SetId("file123")
	resp, err := handler.DeleteBinaryData(ctxWithUserID(userID), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockSvc.AssertExpectations(t)
}

// --- мок стрима для UploadBinaryData ---
type mockUploadStream struct {
	pb.BinaryDataService_UploadBinaryDataServer
	recvMsgs []*pb.UploadBinaryDataRequest
	recvIdx  int
	sentResp *pb.UploadBinaryDataResponse
	ctx      context.Context
}

func (m *mockUploadStream) Recv() (*pb.UploadBinaryDataRequest, error) {
	if m.recvIdx >= len(m.recvMsgs) {
		return nil, io.EOF
	}
	msg := m.recvMsgs[m.recvIdx]
	m.recvIdx++
	return msg, nil
}

func (m *mockUploadStream) SendAndClose(resp *pb.UploadBinaryDataResponse) error {
	m.sentResp = resp
	return nil
}

func (m *mockUploadStream) Context() context.Context {
	return m.ctx
}

// --- мок стрима для UpdateBinaryData ---
type mockUpdateStream struct {
	pb.BinaryDataService_UpdateBinaryDataServer
	recvMsgs []*pb.UpdateBinaryDataRequest
	recvIdx  int
	sentResp *pb.UpdateBinaryDataResponse
	ctx      context.Context
}

func (m *mockUpdateStream) Recv() (*pb.UpdateBinaryDataRequest, error) {
	if m.recvIdx >= len(m.recvMsgs) {
		return nil, io.EOF
	}
	msg := m.recvMsgs[m.recvIdx]
	m.recvIdx++
	return msg, nil
}

func (m *mockUpdateStream) SendAndClose(resp *pb.UpdateBinaryDataResponse) error {
	m.sentResp = resp
	return nil
}

func (m *mockUpdateStream) Context() context.Context {
	return m.ctx
}

// --- мок стрима для DownloadBinaryData ---
type mockDownloadStream struct {
	pb.BinaryDataService_DownloadBinaryDataServer
	sentChunks [][]byte
	ctx        context.Context
	id         string
}

func (m *mockDownloadStream) Send(resp *pb.DownloadBinaryDataResponse) error {
	m.sentChunks = append(m.sentChunks, resp.GetChunk())
	return nil
}

func (m *mockDownloadStream) Context() context.Context {
	return m.ctx
}
