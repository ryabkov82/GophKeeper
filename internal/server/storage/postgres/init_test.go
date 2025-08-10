package postgres_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/server/storage/postgres"
	"github.com/stretchr/testify/assert"
)

// mockDB реализует интерфейс postgres.DB (только метод Ping)
type mockDB struct {
	pingErr error
}

func (m *mockDB) Ping() error {
	return m.pingErr
}

func TestInitWithDeps_Success(t *testing.T) {
	mockDB := &mockDB{pingErr: nil}

	db, err := postgres.InitWithDeps(
		"mock-dsn",
		func(driverName, dsn string) (postgres.DB, *sql.DB, error) {
			return mockDB, nil, nil
		},
		func(db postgres.DB) error {
			return nil
		},
	)

	assert.NoError(t, err)
	assert.Nil(t, db) // потому что мы возвращаем nil вместо *sql.DB
}

func TestInitWithDeps_OpenError(t *testing.T) {
	db, err := postgres.InitWithDeps(
		"mock-dsn",
		func(driverName, dsn string) (postgres.DB, *sql.DB, error) {
			return nil, nil, errors.New("open failed")
		},
		func(db postgres.DB) error {
			return nil
		},
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db open failed")
	assert.Nil(t, db)
}

func TestInitWithDeps_PingError(t *testing.T) {
	mockDB := &mockDB{pingErr: errors.New("ping failed")}

	db, err := postgres.InitWithDeps(
		"mock-dsn",
		func(driverName, dsn string) (postgres.DB, *sql.DB, error) {
			return mockDB, nil, nil
		},
		func(db postgres.DB) error {
			return nil
		},
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db ping failed")
	assert.Nil(t, db)
}

func TestInitWithDeps_MigrationError(t *testing.T) {
	mockDB := &mockDB{pingErr: nil}

	db, err := postgres.InitWithDeps(
		"mock-dsn",
		func(driverName, dsn string) (postgres.DB, *sql.DB, error) {
			return mockDB, nil, nil
		},
		func(db postgres.DB) error {
			return errors.New("migration failed")
		},
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "run migrations")
	assert.Nil(t, db)
}
