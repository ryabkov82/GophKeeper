package service

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
	"github.com/ryabkov82/gophkeeper/internal/domain/storage"
)

type BinaryDataService struct {
	repo    repository.BinaryDataRepository
	storage storage.BinaryDataStorage
}

func NewBinaryDataService(repo repository.BinaryDataRepository, storage storage.BinaryDataStorage) *BinaryDataService {
	return &BinaryDataService{repo: repo, storage: storage}
}

// Create сохраняет файл и метаданные
func (s *BinaryDataService) Create(ctx context.Context, userID string, title string, metadata string, r io.Reader) (*model.BinaryData, error) {
	// Сохраняем файл в хранилище
	storagePath, err := s.storage.Save(ctx, userID, r)
	if err != nil {
		return nil, err
	}

	data := &model.BinaryData{
		ID:          uuid.NewString(),
		UserID:      userID,
		Title:       title,
		StoragePath: storagePath,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Сохраняем метаданные в Postgres
	if err := s.repo.Save(ctx, data); err != nil {
		// Если запись в БД не удалась, удаляем файл из хранилища
		_ = s.storage.Delete(ctx, storagePath)
		return nil, err
	}

	return data, nil
}

// Update перезаписывает бинарные данные и/или метаданные существующей записи.
func (s *BinaryDataService) Update(ctx context.Context, userID, id, title, metadata string, r io.Reader) (*model.BinaryData, error) {
	// Получаем существующую запись
	data, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	var newStoragePath string
	// Если передан поток новых данных, сохраняем их в хранилище
	if r != nil {
		newStoragePath, err = s.storage.Save(ctx, userID, r)
		if err != nil {
			return nil, err
		}
		// Удаляем старый файл после успешной записи нового
		_ = s.storage.Delete(ctx, data.StoragePath)
		data.StoragePath = newStoragePath
	}

	// Обновляем метаданные, если они изменились
	if title != "" {
		data.Title = title
	}
	if metadata != "" {
		data.Metadata = metadata
	}
	data.UpdatedAt = time.Now()

	// Сохраняем изменения в репозитории
	if err := s.repo.Save(ctx, data); err != nil {
		// Если запись в БД не удалась, восстанавливаем старый файл при необходимости
		if newStoragePath != "" {
			_ = s.storage.Delete(ctx, newStoragePath)
		}
		return nil, err
	}

	return data, nil
}

// Get возвращает метаданные и открытый поток для чтения
func (s *BinaryDataService) Get(ctx context.Context, userID, id string) (*model.BinaryData, io.ReadCloser, error) {
	data, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, nil, err
	}

	reader, err := s.storage.Load(ctx, data.StoragePath)
	if err != nil {
		return nil, nil, err
	}

	return data, reader, nil
}

// List возвращает список метаданных
func (s *BinaryDataService) List(ctx context.Context, userID string) ([]*model.BinaryData, error) {
	return s.repo.ListByUser(ctx, userID)
}

// Delete удаляет файл и метаданные
func (s *BinaryDataService) Delete(ctx context.Context, userID, id string) error {
	data, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, userID, id); err != nil {
		return err
	}

	return s.storage.Delete(ctx, data.StoragePath)
}

func (s *BinaryDataService) Close() {
	s.storage.Close()
}
