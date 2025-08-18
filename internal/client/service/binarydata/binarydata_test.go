package binarydata_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/ryabkov82/gophkeeper/internal/client/service/binarydata"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
)

// --- Моковый клиент ---
type mockBinaryDataClient struct {
	pb.BinaryDataServiceClient

	ListErr      error
	UploadErr    error
	UpdateErr    error
	DeleteErr    error
	DownloadData [][]byte
}

func (m *mockBinaryDataClient) ListBinaryData(ctx context.Context, req *pb.ListBinaryDataRequest, opts ...grpc.CallOption) (*pb.ListBinaryDataResponse, error) {
	items := []*pb.BinaryDataInfo{}
	for _, data := range [][]string{{"1", "file1", "meta1"}, {"2", "file2", "meta2"}} {
		item := &pb.BinaryDataInfo{}
		item.SetId(data[0])
		item.SetTitle(data[1])
		item.SetMetadata(data[2])
		items = append(items, item)
	}
	resp := &pb.ListBinaryDataResponse{}
	resp.SetItems(items)
	return resp, m.ListErr
}

func (m *mockBinaryDataClient) GetBinaryDataInfo(ctx context.Context, req *pb.GetBinaryDataInfoRequest, opts ...grpc.CallOption) (*pb.GetBinaryDataInfoResponse, error) {
	// Здесь можно вернуть фиксированный ответ для теста
	info := &pb.BinaryDataInfo{}
	info.SetId(req.GetId())
	info.SetTitle("FileTitle")   // тестовое значение
	info.SetMetadata("FileMeta") // тестовое значение
	info.SetSize(1024)           // тестовое значение

	resp := &pb.GetBinaryDataInfoResponse{}
	resp.SetBinaryInfo(info)

	return resp, nil
}

func (m *mockBinaryDataClient) UploadBinaryData(ctx context.Context, opts ...grpc.CallOption) (pb.BinaryDataService_UploadBinaryDataClient, error) {
	return &mockUploadStream{client: m}, m.UploadErr
}

func (m *mockBinaryDataClient) UpdateBinaryData(ctx context.Context, opts ...grpc.CallOption) (pb.BinaryDataService_UpdateBinaryDataClient, error) {
	return &mockUpdateStream{client: m}, m.UpdateErr
}

func (m *mockBinaryDataClient) DeleteBinaryData(ctx context.Context, req *pb.DeleteBinaryDataRequest, opts ...grpc.CallOption) (*pb.DeleteBinaryDataResponse, error) {
	return &pb.DeleteBinaryDataResponse{}, m.DeleteErr
}

func (m *mockBinaryDataClient) DownloadBinaryData(ctx context.Context, req *pb.DownloadBinaryDataRequest, opts ...grpc.CallOption) (pb.BinaryDataService_DownloadBinaryDataClient, error) {
	return &mockDownloadStream{dataChunks: m.DownloadData}, nil
}

// --- Моки стримов ---
type mockUploadStream struct {
	pb.BinaryDataService_UploadBinaryDataClient
	client *mockBinaryDataClient
}

func (m *mockUploadStream) Send(req *pb.UploadBinaryDataRequest) error { return nil }
func (m *mockUploadStream) CloseAndRecv() (*pb.UploadBinaryDataResponse, error) {
	resp := &pb.UploadBinaryDataResponse{}
	resp.SetId("123")
	return resp, nil
}

type mockUpdateStream struct {
	pb.BinaryDataService_UpdateBinaryDataClient
	client *mockBinaryDataClient
}

func (m *mockUpdateStream) Send(req *pb.UpdateBinaryDataRequest) error { return nil }

func (m *mockUpdateStream) CloseAndRecv() (*pb.UpdateBinaryDataResponse, error) {
	resp := &pb.UpdateBinaryDataResponse{}
	resp.SetId("123")
	return resp, nil
}

type mockDownloadStream struct {
	pb.BinaryDataService_DownloadBinaryDataClient
	dataChunks [][]byte
	idx        int
}

func (m *mockDownloadStream) Recv() (*pb.DownloadBinaryDataResponse, error) {
	if m.idx >= len(m.dataChunks) {
		return nil, io.EOF
	}
	chunk := m.dataChunks[m.idx]
	m.idx++
	resp := &pb.DownloadBinaryDataResponse{}
	resp.SetChunk(chunk)
	return resp, nil
}

// --- Тесты ---

func TestBinaryDataManager_List(t *testing.T) {
	logger := zap.NewNop()
	manager := binarydata.NewBinaryDataManager(logger)
	client := &mockBinaryDataClient{}
	manager.SetClient(client)

	list, err := manager.List(context.Background())
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "1", list[0].ID)
}

func TestBinaryDataManager_Upload(t *testing.T) {
	logger := zap.NewNop()
	manager := binarydata.NewBinaryDataManager(logger)
	client := &mockBinaryDataClient{}
	manager.SetClient(client)

	data := &model.BinaryData{UserID: "user1", Title: "file1", Metadata: "meta1"}
	content := bytes.NewReader([]byte("hello world"))

	err := manager.Upload(context.Background(), data, content)
	assert.NoError(t, err)
	assert.Equal(t, "123", data.ID)
}

func TestBinaryDataManager_Update(t *testing.T) {
	logger := zap.NewNop()
	manager := binarydata.NewBinaryDataManager(logger)
	client := &mockBinaryDataClient{}
	manager.SetClient(client)

	data := &model.BinaryData{ID: "123", UserID: "user1", Title: "file1", Metadata: "meta1"}
	content := bytes.NewReader([]byte("updated content"))

	err := manager.Update(context.Background(), data, content)
	assert.NoError(t, err)
	assert.Equal(t, "123", data.ID)
}

func TestBinaryDataManager_Delete(t *testing.T) {
	logger := zap.NewNop()
	manager := binarydata.NewBinaryDataManager(logger)
	client := &mockBinaryDataClient{}
	manager.SetClient(client)

	err := manager.Delete(context.Background(), "123")
	assert.NoError(t, err)
}

func TestBinaryDataManager_Download(t *testing.T) {
	logger := zap.NewNop()
	manager := binarydata.NewBinaryDataManager(logger)
	client := &mockBinaryDataClient{
		DownloadData: [][]byte{
			[]byte("chunk1"),
			[]byte("chunk2"),
		},
	}
	manager.SetClient(client)

	r, err := manager.Download(context.Background(), "123")
	assert.NoError(t, err)
	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, r)
	assert.Equal(t, "chunk1chunk2", buf.String())
}

func TestBinaryDataManager_GetInfo(t *testing.T) {
	logger := zap.NewNop()
	manager := binarydata.NewBinaryDataManager(logger)
	client := &mockBinaryDataClient{}
	manager.SetClient(client)

	ctx := context.Background()
	dataID := "123"

	info, err := manager.GetInfo(ctx, dataID)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, dataID, info.ID)
	assert.Equal(t, "FileTitle", info.Title)
	assert.Equal(t, "FileMeta", info.Metadata)
	assert.EqualValues(t, 1024, info.Size)
}
