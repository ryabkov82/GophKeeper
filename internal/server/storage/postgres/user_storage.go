package postgres

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// UserStorage реализует хранилище пользователей с использованием PostgreSQL.
//
// Содержит ссылку на открытое подключение к базе данных через *sql.DB.
type UserStorage struct {
	db *sql.DB
}

// NewUserStorage создаёт новый экземпляр UserStorage.
//
// Принимает подключение к базе данных и возвращает структуру,
// реализующую методы доступа к таблице пользователей.
func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db: db}
}

// CreateUser сохраняет нового пользователя в базе данных.
//
// Хеш пароля и соль должны быть заранее вычислены (например, через bcrypt).
//
// Параметры:
//   - ctx: контекст выполнения (может содержать таймаут или отмену);
//   - login: логин пользователя (уникальный);
//   - hash: хеш пароля;
//   - salt: соль, использованная при хешировании.
//
// Возвращает ошибку, если пользователь не может быть добавлен
// (например, логин уже существует или возникает ошибка SQL).
func (s *UserStorage) CreateUser(ctx context.Context, login, hash, salt string) error {
	query := `
    INSERT INTO users (login, password_hash, salt)
    VALUES ($1, $2, $3)
  `
	_, err := s.db.ExecContext(ctx, query, login, hash, salt)
	if err != nil {
		return err
	}
	return nil
}

// GetUserByLogin находит пользователя по логину.
//
// Выполняет запрос к таблице пользователей и возвращает структуру model.User,
// содержащую ID, логин, хеш пароля и соль.
//
// Параметры:
//   - ctx: контекст выполнения (может содержать таймаут или отмену);
//   - login: логин пользователя для поиска.
//
// Возвращает:
//   - *model.User: если пользователь найден;
//   - nil: если пользователь не существует;
//   - ошибку: при возникновении SQL-ошибок, кроме sql.ErrNoRows.
func (s *UserStorage) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	query := `
    SELECT id, login, password_hash, salt
    FROM users
    WHERE login = $1
  `
	row := s.db.QueryRowContext(ctx, query, login)

	var user model.User
	err := row.Scan(&user.ID, &user.Login, &user.PasswordHash, &user.Salt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
