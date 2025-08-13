package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// PostgresStorage реализует интерфейс CredentialRepository для PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage создаёт новый экземпляр PostgresStorage
func NewCredentialStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

// Create сохраняет новую запись учётных данных
func (s *PostgresStorage) Create(ctx context.Context, cred *model.Credential) error {
	query := `
		INSERT INTO credentials (id, user_id, title, login, password, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err := s.db.ExecContext(ctx, query,
		cred.ID,
		cred.UserID,
		cred.Title,
		cred.Login,
		cred.Password,
		cred.Metadata,
	)
	return err
}

// GetByID возвращает запись по ID
func (s *PostgresStorage) GetByID(ctx context.Context, id string) (*model.Credential, error) {
	query := `
		SELECT id, user_id, title, login, password, metadata, created_at, updated_at
		FROM credentials WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)
	var cred model.Credential
	err := row.Scan(
		&cred.ID,
		&cred.UserID,
		&cred.Title,
		&cred.Login,
		&cred.Password,
		&cred.Metadata,
		&cred.CreatedAt,
		&cred.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("credential not found")
	}
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// GetByUserID возвращает все записи пользователя
func (s *PostgresStorage) GetByUserID(ctx context.Context, userID string) ([]model.Credential, error) {
	query := `
		SELECT id, user_id, title, login, password, metadata, created_at, updated_at
		FROM credentials WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []model.Credential
	for rows.Next() {
		var cred model.Credential
		if err := rows.Scan(
			&cred.ID,
			&cred.UserID,
			&cred.Title,
			&cred.Login,
			&cred.Password,
			&cred.Metadata,
			&cred.CreatedAt,
			&cred.UpdatedAt,
		); err != nil {
			return nil, err
		}
		creds = append(creds, cred)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return creds, nil
}

// Update изменяет существующую запись
func (s *PostgresStorage) Update(ctx context.Context, cred *model.Credential) error {
	query := `
		UPDATE credentials
		SET title = $1, login = $2, password = $3, metadata = $4, updated_at = NOW()
		WHERE id = $5
	`
	res, err := s.db.ExecContext(ctx, query,
		cred.Title,
		cred.Login,
		cred.Password,
		cred.Metadata,
		cred.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("credential not found")
	}
	return nil
}

// Delete удаляет запись по ID
func (s *PostgresStorage) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM credentials WHERE id = $1`
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("credential not found")
	}
	return nil
}
