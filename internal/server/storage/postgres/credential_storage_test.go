package postgres_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/server/storage/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCreateCredential_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewCredentialStorage(db)

	cred := &model.Credential{
		ID:       "uuid-1234",
		UserID:   "user-uuid",
		Title:    "GitHub",
		Login:    "login123",
		Password: "encryptedpass",
		Metadata: "some meta",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO credentials (id, user_id, title, login, password, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`)).
		WithArgs(cred.ID, cred.UserID, cred.Title, cred.Login, cred.Password, cred.Metadata).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.Create(context.Background(), cred)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCredentialByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewCredentialStorage(db)

	createdAt := time.Now()
	updatedAt := createdAt

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "title", "login", "password", "metadata", "created_at", "updated_at",
	}).AddRow("uuid-1234", "user-uuid", "GitHub", "login123", "encryptedpass", "some meta", createdAt, updatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, title, login, password, metadata, created_at, updated_at
		FROM credentials WHERE id = $1
	`)).
		WithArgs("uuid-1234").
		WillReturnRows(rows)

	cred, err := storage.GetByID(context.Background(), "uuid-1234")
	assert.NoError(t, err)
	assert.NotNil(t, cred)
	assert.Equal(t, "uuid-1234", cred.ID)
	assert.Equal(t, "user-uuid", cred.UserID)
	assert.Equal(t, "GitHub", cred.Title)
	assert.Equal(t, "login123", cred.Login)
	assert.Equal(t, "encryptedpass", cred.Password)
	assert.Equal(t, "some meta", cred.Metadata)
	assert.WithinDuration(t, createdAt, cred.CreatedAt, time.Second)
	assert.WithinDuration(t, updatedAt, cred.UpdatedAt, time.Second)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCredentialByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewCredentialStorage(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, title, login, password, metadata, created_at, updated_at
		FROM credentials WHERE id = $1
	`)).
		WithArgs("non-existent-id").
		WillReturnError(sql.ErrNoRows)

	cred, err := storage.GetByID(context.Background(), "non-existent-id")
	assert.Error(t, err)
	assert.Nil(t, cred)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCredential_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewCredentialStorage(db)

	cred := &model.Credential{
		ID:       "uuid-1234",
		Title:    "Updated Title",
		Login:    "updated-login",
		Password: "updated-password",
		Metadata: "updated meta",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE credentials
		SET title = $1, login = $2, password = $3, metadata = $4, updated_at = NOW()
		WHERE id = $5
	`)).
		WithArgs(cred.Title, cred.Login, cred.Password, cred.Metadata, cred.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.Update(context.Background(), cred)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCredential_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewCredentialStorage(db)

	cred := &model.Credential{
		ID:       "uuid-1234",
		Title:    "Updated Title",
		Login:    "updated-login",
		Password: "updated-password",
		Metadata: "updated meta",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE credentials
		SET title = $1, login = $2, password = $3, metadata = $4, updated_at = NOW()
		WHERE id = $5
	`)).
		WithArgs(cred.Title, cred.Login, cred.Password, cred.Metadata, cred.ID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	err = storage.Update(context.Background(), cred)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCredential_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewCredentialStorage(db)

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM credentials WHERE id = $1`)).
		WithArgs("uuid-1234").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.Delete(context.Background(), "uuid-1234")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCredential_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewCredentialStorage(db)

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM credentials WHERE id = $1`)).
		WithArgs("uuid-1234").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = storage.Delete(context.Background(), "uuid-1234")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_GetByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	s := postgres.NewCredentialStorage(db)

	userID := "user123"
	createdAt := time.Now().Add(-time.Hour)
	updatedAt := time.Now()

	// Создаем ожидаемые строки результата
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "title", "login", "password", "metadata", "created_at", "updated_at",
	}).AddRow(
		"cred1", userID, "Title1", "login1", "pass1", "meta1", createdAt, updatedAt,
	).AddRow(
		"cred2", userID, "Title2", "login2", "pass2", "meta2", createdAt, updatedAt,
	)

	// Ожидаемый SQL запрос
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, title, login, password, metadata, created_at, updated_at
		FROM credentials WHERE user_id = $1 ORDER BY created_at DESC
	`)).WithArgs(userID).WillReturnRows(rows)

	ctx := context.Background()
	creds, err := s.GetByUserID(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, creds, 2)

	assert.Equal(t, "cred1", creds[0].ID)
	assert.Equal(t, userID, creds[0].UserID)
	assert.Equal(t, "Title1", creds[0].Title)

	assert.Equal(t, "cred2", creds[1].ID)
	assert.Equal(t, userID, creds[1].UserID)
	assert.Equal(t, "Title2", creds[1].Title)

	// Проверяем, что все ожидания моков были выполнены
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
