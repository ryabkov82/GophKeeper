package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ryabkov82/gophkeeper/internal/migrations"
)

// DB — интерфейс, содержащий только метод Ping, необходимый для инициализации.
type DB interface {
	Ping() error
}

// DBOpener — функция, открывающая соединение с БД и возвращающая интерфейс DB и оригинальный *sql.DB.
type DBOpener func(driverName, dsn string) (DB, *sql.DB, error)

// MigrationRunner — функция, применяющая миграции к базе данных через интерфейс DB.
type MigrationRunner func(DB) error

// InitWithDeps выполняет инициализацию БД с использованием переданных зависимостей (опенер и раннер миграций).
func InitWithDeps(dsn string, opener DBOpener, migrator MigrationRunner) (*sql.DB, error) {
	dbIface, dbRaw, err := opener("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("db open failed: %w", err)
	}

	if err := dbIface.Ping(); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	if err := migrator(dbIface); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return dbRaw, nil
}

// Init — основной публичный метод инициализации, использующий реальные зависимости.
func Init(dsn string) (*sql.DB, error) {
	return InitWithDeps(
		dsn,
		func(driverName, dsn string) (DB, *sql.DB, error) {
			db, err := sql.Open(driverName, dsn)
			if err != nil {
				return nil, nil, err
			}
			return db, db, nil
		},
		func(db DB) error {
			if realDB, ok := db.(*sql.DB); ok {
				return migrations.RunMigrations(realDB)
			}
			return fmt.Errorf("invalid db type, expected *sql.DB")
		},
	)
}
