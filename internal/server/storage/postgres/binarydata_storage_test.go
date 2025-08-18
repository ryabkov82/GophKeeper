package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
	"github.com/ryabkov82/gophkeeper/internal/server/storage/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockBinaryDataDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, repository.BinaryDataRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	repo := postgres.NewBinaryDataStorage(db)
	return db, mock, repo
}

func TestBinaryDataStorage_Save(t *testing.T) {
	db, mock, repo := setupMockBinaryDataDB(t)
	defer db.Close()

	data := &model.BinaryData{
		UserID:      uuid.NewString(),
		Title:       "Test File",
		Size:        1,
		StoragePath: "/tmp/testfile.bin",
		Metadata:    "{}",
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO binary_data`)).
		WithArgs(sqlmock.AnyArg(), data.UserID, data.Title, data.StoragePath, data.Size, data.Metadata).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Save(context.Background(), data)
	assert.NoError(t, err)
}

func TestBinaryDataStorage_GetByID(t *testing.T) {
	db, mock, repo := setupMockBinaryDataDB(t)
	defer db.Close()

	id := uuid.NewString()
	userID := uuid.NewString()
	data := &model.BinaryData{
		ID:          id,
		UserID:      userID,
		Title:       "Test File",
		StoragePath: "/tmp/testfile.bin",
		Metadata:    "{}",
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "storage_path", "metadata"}).
		AddRow(data.ID, data.UserID, data.Title, data.StoragePath, data.Metadata)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM binary_data WHERE id = $1 AND user_id = $2")).
		WithArgs(id, userID).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), userID, id)
	assert.NoError(t, err)
	assert.Equal(t, data.ID, result.ID)
	assert.Equal(t, data.StoragePath, result.StoragePath)
}

func TestBinaryDataStorage_ListByUser(t *testing.T) {
	db, mock, repo := setupMockBinaryDataDB(t)
	defer db.Close()

	userID := uuid.NewString()
	list := []*model.BinaryData{
		{ID: uuid.NewString(), Title: "File 1", StoragePath: "/tmp/1.bin"},
		{ID: uuid.NewString(), Title: "File 2", StoragePath: "/tmp/2.bin"},
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "storage_path", "metadata"}).
		AddRow(list[0].ID, userID, list[0].Title, list[0].StoragePath, "{}").
		AddRow(list[1].ID, userID, list[1].Title, list[1].StoragePath, "{}")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM binary_data WHERE user_id = $1 ORDER BY created_at DESC")).
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := repo.ListByUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestBinaryDataStorage_Delete(t *testing.T) {
	db, mock, repo := setupMockBinaryDataDB(t)
	defer db.Close()

	id := uuid.NewString()
	userID := uuid.NewString()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM binary_data WHERE id = $1 AND user_id = $2")).
		WithArgs(id, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), userID, id)
	assert.NoError(t, err)
}

func TestBinaryDataStorage_Delete_NotFound(t *testing.T) {
	db, mock, repo := setupMockBinaryDataDB(t)
	defer db.Close()

	id := uuid.NewString()
	userID := uuid.NewString()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM binary_data WHERE id = $1 AND user_id = $2")).
		WithArgs(id, userID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	err := repo.Delete(context.Background(), userID, id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBinaryDataStorage_Save_Error(t *testing.T) {
	db, mock, repo := setupMockBinaryDataDB(t)
	defer db.Close()

	data := &model.BinaryData{
		UserID:      uuid.NewString(),
		Title:       "Test File",
		Size:        1,
		StoragePath: "/tmp/testfile.bin",
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO binary_data`)).
		WithArgs(sqlmock.AnyArg(), data.UserID, data.Title, data.StoragePath, data.Size, data.Metadata).
		WillReturnError(errors.New("insert failed"))

	err := repo.Save(context.Background(), data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert failed")
}

func TestBinaryDataStorage_Update(t *testing.T) {

	db, mock, repo := setupMockBinaryDataDB(t)
	defer db.Close()

	data := &model.BinaryData{
		ID:          "123",
		UserID:      "456",
		Title:       "new title",
		StoragePath: "/tmp/file.bin",
		Size:        100,
		Metadata:    "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// --- 1. Успешное обновление ---
	mock.ExpectExec(`UPDATE binary_data`).
		WithArgs(
			data.Title,
			data.StoragePath,
			data.Size,
			data.Metadata,
			data.ID,
			data.UserID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 строка обновлена

	err := repo.Update(context.Background(), data)
	require.NoError(t, err)

	// --- 2. Нет обновлённых строк ---
	mock.ExpectExec(`UPDATE binary_data`).
		WithArgs(
			data.Title,
			data.StoragePath,
			data.Size,
			data.Metadata,
			data.ID,
			data.UserID,
		).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк обновлено

	err = repo.Update(context.Background(), data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "binary data with id")

	// --- 3. Ошибка БД ---
	mock.ExpectExec(`UPDATE binary_data`).
		WithArgs(
			data.Title,
			data.StoragePath,
			data.Size,
			data.Metadata,
			data.ID,
			data.UserID,
		).
		WillReturnError(sql.ErrConnDone)

	err = repo.Update(context.Background(), data)
	require.Error(t, err)
	require.Equal(t, sql.ErrConnDone, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
