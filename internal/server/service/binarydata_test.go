package service_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/server/service"
)

// --- моки ---

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Save(ctx context.Context, data *model.BinaryData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *mockRepo) Update(ctx context.Context, data *model.BinaryData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *mockRepo) GetByID(ctx context.Context, userID, id string) (*model.BinaryData, error) {
	args := m.Called(ctx, userID, id)
	if v := args.Get(0); v != nil {
		return v.(*model.BinaryData), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepo) ListByUser(ctx context.Context, userID string) ([]*model.BinaryData, error) {
	args := m.Called(ctx, userID)
	if v := args.Get(0); v != nil {
		return v.([]*model.BinaryData), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepo) Delete(ctx context.Context, userID, id string) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

// ----

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Save(ctx context.Context, userID string, r io.Reader) (string, int64, error) {
	args := m.Called(ctx, userID, r)
	return args.String(0), args.Get(1).(int64), args.Error(2)
}

func (m *mockStorage) Load(ctx context.Context, path string) (io.ReadCloser, error) {
	args := m.Called(ctx, path)
	if v := args.Get(0); v != nil {
		return v.(io.ReadCloser), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockStorage) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *mockStorage) Close() {
	m.Called()
}

// --- тесты ---

func TestBinaryDataService_Create_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	storage := new(mockStorage)
	svc := service.NewBinaryDataService(repo, storage)

	userID := "user1"
	content := []byte("hello world")
	reader := bytes.NewReader(content)
	storagePath := "path/to/file"

	// моки
	storage.On("Save", ctx, userID, mock.Anything).Return(storagePath, int64(1), nil).Once()
	repo.On("Save", ctx, mock.AnythingOfType("*model.BinaryData")).Return(nil).Once()

	data, err := svc.Create(ctx, userID, "title", "meta", reader)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, userID, data.UserID)
	assert.Equal(t, "title", data.Title)
	assert.Equal(t, "meta", data.Metadata)
	assert.Equal(t, storagePath, data.StoragePath)
	assert.NotEmpty(t, data.ID)

	storage.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestBinaryDataService_Create_RepoError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	storage := new(mockStorage)
	svc := service.NewBinaryDataService(repo, storage)

	userID := "user1"
	reader := bytes.NewReader([]byte("fail"))
	storagePath := "path/to/file"

	storage.On("Save", ctx, userID, mock.Anything).Return(storagePath, int64(1), nil).Once()
	repo.On("Save", ctx, mock.AnythingOfType("*model.BinaryData")).Return(errors.New("db fail")).Once()
	storage.On("Delete", ctx, storagePath).Return(nil).Once()

	data, err := svc.Create(ctx, userID, "title", "meta", reader)
	assert.Nil(t, data)
	assert.EqualError(t, err, "db fail")

	storage.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestBinaryDataService_Get_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	storage := new(mockStorage)
	svc := service.NewBinaryDataService(repo, storage)

	userID := "user1"
	id := uuid.NewString()
	bd := &model.BinaryData{ID: id, UserID: userID, StoragePath: "path/to/file"}
	reader := io.NopCloser(bytes.NewReader([]byte("data")))

	repo.On("GetByID", ctx, userID, id).Return(bd, nil).Once()
	storage.On("Load", ctx, bd.StoragePath).Return(reader, nil).Once()

	got, r, err := svc.Get(ctx, userID, id)
	assert.NoError(t, err)
	assert.Equal(t, bd, got)
	assert.Equal(t, reader, r)

	storage.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestBinaryDataService_GetInfo_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	storage := new(mockStorage) // нужен для конструктора, но в тесте не используется
	svc := service.NewBinaryDataService(repo, storage)

	userID := "user1"
	id := uuid.NewString()
	bd := &model.BinaryData{ID: id, UserID: userID, Title: "test"}

	// Ожидаем вызов только репозитория
	repo.On("GetByID", ctx, userID, id).Return(bd, nil).Once()

	got, err := svc.GetInfo(ctx, userID, id)
	assert.NoError(t, err)
	assert.Equal(t, bd, got)

	repo.AssertExpectations(t)
	storage.AssertExpectations(t) // не должно быть вызовов
}

func TestBinaryDataService_List(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	storage := new(mockStorage)
	svc := service.NewBinaryDataService(repo, storage)

	userID := "user1"
	expected := []*model.BinaryData{{ID: "1"}, {ID: "2"}}

	repo.On("ListByUser", ctx, userID).Return(expected, nil).Once()

	got, err := svc.List(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)

	repo.AssertExpectations(t)
}

func TestBinaryDataService_Delete_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	storage := new(mockStorage)
	svc := service.NewBinaryDataService(repo, storage)

	userID := "user1"
	id := "id1"
	bd := &model.BinaryData{ID: id, StoragePath: "path/to/file"}

	repo.On("GetByID", ctx, userID, id).Return(bd, nil).Once()
	repo.On("Delete", ctx, userID, id).Return(nil).Once()
	storage.On("Delete", ctx, bd.StoragePath).Return(nil).Once()

	err := svc.Delete(ctx, userID, id)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
}

func TestBinaryDataService_Close(t *testing.T) {
	repo := new(mockRepo)
	storage := new(mockStorage)
	svc := service.NewBinaryDataService(repo, storage)

	storage.On("Close").Once()
	svc.Close()
	storage.AssertExpectations(t)
}

func TestBinaryDataService_Update_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockRepo)
	storage := new(mockStorage)
	svc := service.NewBinaryDataService(repo, storage)

	userID := "user1"
	id := "file123"
	oldPath := "user1/old.bin"
	newPath := "user1/new.bin"
	newTitle := "NewTitle"
	newMetadata := "newMeta"
	newContent := []byte("new content")

	// Существующая запись
	existing := &model.BinaryData{
		ID:          id,
		UserID:      userID,
		Title:       "OldTitle",
		StoragePath: oldPath,
		Metadata:    "oldMeta",
	}

	// Моки репозитория
	repo.On("GetByID", ctx, userID, id).Return(existing, nil).Once()
	repo.On("Update", ctx, mock.Anything).Return(nil).Once()

	// Моки хранилища
	storage.On("Save", ctx, userID, mock.Anything).Return(newPath, int64(1), nil).Once()
	storage.On("Delete", ctx, oldPath).Return(nil).Once()

	// Вызываем метод
	r := bytes.NewReader(newContent)
	updated, err := svc.Update(ctx, userID, id, newTitle, newMetadata, r)
	assert.NoError(t, err)
	assert.Equal(t, newTitle, updated.Title)
	assert.Equal(t, newMetadata, updated.Metadata)
	assert.Equal(t, newPath, updated.StoragePath)

	repo.AssertExpectations(t)
	storage.AssertExpectations(t)
}
