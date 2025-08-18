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

// binaryDataStorage — хранилище бинарных данных.
type binaryDataStorage struct {
	db *sqlx.DB
}

// NewBinaryDataStorage создаёт новое хранилище BinaryData.
func NewBinaryDataStorage(db *sql.DB) repository.BinaryDataRepository {
	sqlxDB := sqlx.NewDb(db, "pgx")
	return &binaryDataStorage{db: sqlxDB}
}

// Save сохраняет новую запись бинарных данных.
func (s *binaryDataStorage) Save(ctx context.Context, data *model.BinaryData) error {
	data.ID = uuid.NewString()
	query := `
		INSERT INTO binary_data (
			id, user_id, title, storage_path, metadata, created_at, updated_at
		) VALUES (
			:id, :user_id, :title, :storage_path, :size, :metadata, NOW(), NOW()
		)`
	_, err := s.db.NamedExecContext(ctx, query, data)
	return err
}

func (s *binaryDataStorage) Update(ctx context.Context, data *model.BinaryData) error {
	query := `
		UPDATE binary_data
		SET title = :title,
		    storage_path = :storage_path,
		    size = :size,
		    metadata = :metadata,
		    updated_at = NOW()
		WHERE id = :id AND user_id = :user_id`

	res, err := s.db.NamedExecContext(ctx, query, data)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("binary data with id %s not found", data.ID)
	}

	return nil
}

// GetByID возвращает запись бинарных данных по id и userID.
func (s *binaryDataStorage) GetByID(ctx context.Context, userID, id string) (*model.BinaryData, error) {
	var data model.BinaryData
	query := `SELECT * FROM binary_data WHERE id = $1 AND user_id = $2`
	err := s.db.GetContext(ctx, &data, query, id, userID)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// ListByUser возвращает все бинарные данные конкретного пользователя.
func (s *binaryDataStorage) ListByUser(ctx context.Context, userID string) ([]*model.BinaryData, error) {
	query := `SELECT * FROM binary_data WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := s.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.BinaryData
	for rows.Next() {
		var data model.BinaryData
		if err := rows.StructScan(&data); err != nil {
			return nil, err
		}
		list = append(list, &data)
	}
	return list, nil
}

// Delete удаляет запись по id и userID.
func (s *binaryDataStorage) Delete(ctx context.Context, userID, id string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM binary_data WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("binary data with id %s not found", id)
	}
	return nil
}
