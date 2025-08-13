package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var embedMigrations embed.FS

// RunMigrations применяет миграции к базе данных с использованием библиотеки Goose.
//
// Функция выполняет следующие действия:
//   - Устанавливает виртуальную файловую систему для встроенных миграций (embedMigrations);
//   - Устанавливает диалект базы данных как "postgres";
//   - Запускает выполнение всех доступных миграций вверх (goose.Up);
//
// Параметры:
//   - db: открытое соединение к PostgreSQL (*sql.DB);
//
// Возвращает:
//   - error: если установка диалекта или выполнение миграций завершились с ошибкой.
//
// При успешном выполнении в стандартный лог выводится сообщение о применении миграций.
func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	err := goose.Up(db, "sql")
	if err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	log.Println("migrations applied successfully")
	return nil
}
