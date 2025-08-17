package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
)

// textDataStorage — хранилище произвольных текстовых данных.
type textDataStorage struct {
	db *sqlx.DB
}

// NewTextDataStorage создаёт новое хранилище TextData.
// Принимает стандартный *sql.DB и оборачивает его в *sqlx.DB.
func NewTextDataStorage(db *sql.DB) repository.TextDataRepository {
	sqlxDB := sqlx.NewDb(db, "pgx")
	return &textDataStorage{db: sqlxDB}
}

// Create сохраняет новую запись TextData.
func (s *textDataStorage) Create(ctx context.Context, data *model.TextData) error {
	data.ID = uuid.NewString()
	query := `
		INSERT INTO text_data (
			id, user_id, title, content, metadata, created_at, updated_at
		) VALUES (
			:id, :user_id, :title, :content, :metadata, NOW(), NOW()
		)`
	_, err := s.db.NamedExecContext(ctx, query, data)
	return err
}

// GetByID возвращает полную запись TextData по id и userID.
func (s *textDataStorage) GetByID(ctx context.Context, userID, id string) (*model.TextData, error) {
	var data model.TextData
	query := `SELECT * FROM text_data WHERE id = $1 AND user_id = $2`
	err := s.db.GetContext(ctx, &data, query, id, userID)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// Update обновляет существующую запись TextData.
func (s *textDataStorage) Update(ctx context.Context, data *model.TextData) error {
	query := `
		UPDATE text_data
		SET title = :title,
		    content = :content,
		    metadata = :metadata,
		    updated_at = NOW()
		WHERE id = :id AND user_id = :user_id`
	res, err := s.db.NamedExecContext(ctx, query, data)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("text data with id %s not found", data.ID)
	}
	return nil
}

// Delete удаляет запись TextData по id и userID.
func (s *textDataStorage) Delete(ctx context.Context, userID, id string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM text_data WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("text data with id %s not found", id)
	}
	return nil
}

// ListTitles возвращает список всех записей пользователя с ID и Title.
func (s *textDataStorage) ListTitles(ctx context.Context, userID string) ([]*model.TextData, error) {
	query := `
		SELECT id, title
		FROM text_data
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.TextData
	for rows.Next() {
		var data model.TextData
		if err := rows.StructScan(&data); err != nil {
			return nil, err
		}
		list = append(list, &data)
	}
	return list, nil
}
