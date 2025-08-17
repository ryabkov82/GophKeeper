package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
	"github.com/ryabkov82/gophkeeper/internal/server/storage/postgres"
	"github.com/stretchr/testify/assert"
)

func setupMockTextDataDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, repository.TextDataRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	repo := postgres.NewTextDataStorage(db)
	return db, mock, repo
}

func TestTextDataStorage_Create(t *testing.T) {
	db, mock, repo := setupMockTextDataDB(t)
	defer db.Close()

	data := &model.TextData{
		UserID:   uuid.NewString(),
		Title:    "Test Note",
		Content:  []byte("encrypted content"),
		Metadata: "{}",
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO text_data`)).
		WithArgs(sqlmock.AnyArg(), data.UserID, data.Title, data.Content, data.Metadata).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), data)
	assert.NoError(t, err)
}

func TestTextDataStorage_GetByID(t *testing.T) {
	db, mock, repo := setupMockTextDataDB(t)
	defer db.Close()

	id := uuid.NewString()
	data := &model.TextData{
		ID:     id,
		UserID: uuid.NewString(),
		Title:  "Test Note",
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "content", "metadata"}).
		AddRow(data.ID, data.UserID, data.Title, []byte("content"), "{}")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM text_data WHERE id = $1 AND user_id = $2")).
		WithArgs(id, data.UserID).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), data.UserID, id)
	assert.NoError(t, err)
	assert.Equal(t, data.ID, result.ID)
	assert.Equal(t, data.Title, result.Title)
}

func TestTextDataStorage_Update(t *testing.T) {
	db, mock, repo := setupMockTextDataDB(t)
	defer db.Close()

	data := &model.TextData{
		ID:       uuid.NewString(),
		UserID:   uuid.NewString(),
		Title:    "Updated",
		Content:  []byte("new content"),
		Metadata: "{}",
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE text_data`)).
		WithArgs(data.Title, data.Content, data.Metadata, data.ID, data.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), data)
	assert.NoError(t, err)
}

func TestTextDataStorage_Delete(t *testing.T) {
	db, mock, repo := setupMockTextDataDB(t)
	defer db.Close()

	id := uuid.NewString()
	userID := uuid.NewString()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM text_data WHERE id = $1 AND user_id = $2")).
		WithArgs(id, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), userID, id)
	assert.NoError(t, err)
}

func TestTextDataStorage_Delete_NotFound(t *testing.T) {
	db, mock, repo := setupMockTextDataDB(t)
	defer db.Close()

	id := uuid.NewString()
	userID := uuid.NewString()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM text_data WHERE id = $1 AND user_id = $2")).
		WithArgs(id, userID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	err := repo.Delete(context.Background(), userID, id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestTextDataStorage_ListTitles(t *testing.T) {
	db, mock, repo := setupMockTextDataDB(t)
	defer db.Close()

	userID := uuid.NewString()
	list := []*model.TextData{
		{ID: uuid.NewString(), Title: "Note 1"},
		{ID: uuid.NewString(), Title: "Note 2"},
	}

	rows := sqlmock.NewRows([]string{"id", "title"}).
		AddRow(list[0].ID, list[0].Title).
		AddRow(list[1].ID, list[1].Title)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title FROM text_data WHERE user_id = $1 ORDER BY created_at DESC")).
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := repo.ListTitles(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, list[0].Title, result[0].Title)
	assert.Equal(t, list[1].Title, result[1].Title)
}

func TestTextDataStorage_Create_Error(t *testing.T) {
	db, mock, repo := setupMockTextDataDB(t)
	defer db.Close()

	data := &model.TextData{
		UserID:  uuid.NewString(),
		Title:   "Test Note",
		Content: []byte("content"),
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO text_data`)).
		WithArgs(sqlmock.AnyArg(), data.UserID, data.Title, data.Content, data.Metadata).
		WillReturnError(errors.New("insert failed"))

	err := repo.Create(context.Background(), data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert failed")
}

func TestTextDataStorage_GetByID_NotFound(t *testing.T) {
	db, mock, repo := setupMockTextDataDB(t)
	defer db.Close()

	id := uuid.NewString()
	userID := uuid.NewString()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM text_data WHERE id = $1 AND user_id = $2")).
		WithArgs(id, userID).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByID(context.Background(), userID, id)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, result)
}

func TestTextDataStorage_Update_NotFound(t *testing.T) {
	db, mock, repo := setupMockTextDataDB(t)
	defer db.Close()

	data := &model.TextData{
		ID:      uuid.NewString(),
		UserID:  uuid.NewString(),
		Title:   "X",
		Content: []byte("content"),
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE text_data`)).
		WithArgs(data.Title, data.Content, data.Metadata, data.ID, data.UserID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	err := repo.Update(context.Background(), data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestTextDataStorage_Update_Error(t *testing.T) {
	db, mock, repo := setupMockTextDataDB(t)
	defer db.Close()

	data := &model.TextData{
		ID:      uuid.NewString(),
		UserID:  uuid.NewString(),
		Title:   "X",
		Content: []byte("content"),
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE text_data`)).
		WithArgs(data.Title, data.Content, data.Metadata, data.ID, data.UserID).
		WillReturnError(errors.New("update failed"))

	err := repo.Update(context.Background(), data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
}
