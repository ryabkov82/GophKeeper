// internal/server/storage/postgres/user_storage.go
package postgres

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db: db}
}

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
