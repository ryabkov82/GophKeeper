package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ryabkov82/gophkeeper/internal/server/storage/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewUserStorage(db)

	mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO users (login, password_hash, salt)
        VALUES ($1, $2, $3)
    `)).
		WithArgs("testuser", "hashedpass", "somesalt").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = storage.CreateUser(context.Background(), "testuser", "hashedpass", "somesalt")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_SQLFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewUserStorage(db)

	mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO users (login, password_hash, salt)
        VALUES ($1, $2, $3)
    `)).
		WithArgs("testuser", "hashedpass", "somesalt").
		WillReturnError(errors.New("insert error"))

	err = storage.CreateUser(context.Background(), "testuser", "hashedpass", "somesalt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByLogin_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewUserStorage(db)

	rows := sqlmock.NewRows([]string{"id", "login", "password_hash", "salt"}).
		AddRow("123", "testuser", "hashedpass", "somesalt")

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, login, password_hash, salt
        FROM users
        WHERE login = $1
    `)).
		WithArgs("testuser").
		WillReturnRows(rows)

	user, err := storage.GetUserByLogin(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "123", user.ID)
	assert.Equal(t, "testuser", user.Login)
	assert.Equal(t, "hashedpass", user.PasswordHash)
	assert.Equal(t, "somesalt", user.Salt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByLogin_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewUserStorage(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, login, password_hash, salt
        FROM users
        WHERE login = $1
    `)).
		WithArgs("unknownuser").
		WillReturnError(sql.ErrNoRows)

	user, err := storage.GetUserByLogin(context.Background(), "unknownuser")
	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByLogin_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	storage := postgres.NewUserStorage(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
        SELECT id, login, password_hash, salt
        FROM users
        WHERE login = $1
    `)).
		WithArgs("baduser").
		WillReturnError(errors.New("db error"))

	user, err := storage.GetUserByLogin(context.Background(), "baduser")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
