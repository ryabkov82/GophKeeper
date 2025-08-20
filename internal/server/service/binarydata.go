package service

import (
	"context"
	"errors"
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
func (s *BinaryDataService) Create(ctx context.Context, data *model.BinaryData, r io.Reader) (*model.BinaryData, error) {
	// Сохраняем файл в хранилище
	storagePath, size, err := s.storage.Save(ctx, data.UserID, r)
	if err != nil {
		return nil, err
	}

	data.ID = uuid.NewString()
	data.StoragePath = storagePath
	data.Size = size
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	// Сохраняем метаданные в Postgres
	if err := s.repo.Save(ctx, data); err != nil {
		// Если запись в БД не удалась, удаляем файл из хранилища
		_ = s.storage.Delete(ctx, storagePath)
		return nil, err
	}

	return data, nil
}

// CreateInfo сохраняет только метаданные без бинарного содержимого.
func (s *BinaryDataService) CreateInfo(ctx context.Context, data *model.BinaryData) (*model.BinaryData, error) {

	data.ID = uuid.NewString()
	data.StoragePath = ""
	data.Size = 0
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, data); err != nil {
		return nil, err
	}

	return data, nil
}

// Update перезаписывает бинарные данные и/или метаданные существующей записи.
func (s *BinaryDataService) Update(ctx context.Context, data *model.BinaryData, r io.Reader) (*model.BinaryData, error) {
	// Получаем существующую запись
	stored, err := s.repo.GetByID(ctx, data.UserID, data.ID)
	if err != nil {
		return nil, err
	}

	if stored == nil {
		return nil, errors.New("binary data not found")
	}

	var newStoragePath, oldStoragePath string
	var newSize int64
	// Если передан поток новых данных, сохраняем их в хранилище
	if r != nil {
		newStoragePath, newSize, err = s.storage.Save(ctx, data.UserID, r)
		if err != nil {
			return nil, err
		}
		oldStoragePath = stored.StoragePath
		stored.StoragePath = newStoragePath
	}

	// Обновляем метаданные, если они изменились
	stored.Title = data.Title
	stored.Metadata = data.Metadata
	stored.ClientPath = data.ClientPath
	stored.Size = newSize
	stored.UpdatedAt = time.Now()

	// Сохраняем изменения в репозитории
	if err := s.repo.Update(ctx, stored); err != nil {
		// Если запись в БД не удалась, восстанавливаем старый файл при необходимости
		if newStoragePath != "" {
			_ = s.storage.Delete(ctx, newStoragePath)
		}
		return nil, err
	}
	if newStoragePath != "" {
		// Удаляем старый файл после успешной записи нового
		_ = s.storage.Delete(ctx, oldStoragePath)
	}

	return stored, nil
}

// UpdateInfo изменяет только метаданные файла без перезаписи его содержимого.
func (s *BinaryDataService) UpdateInfo(ctx context.Context, data *model.BinaryData) (*model.BinaryData, error) {

	stored, err := s.repo.GetByID(ctx, data.UserID, data.ID)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return nil, errors.New("binary data not found")
	}

	stored.Title = data.Title
	stored.Metadata = data.Metadata
	stored.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, stored); err != nil {
		return nil, err
	}
	return stored, nil
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

// GetInfo возвращает только метаданные без чтения бинарного содержимого.
func (s *BinaryDataService) GetInfo(ctx context.Context, userID, id string) (*model.BinaryData, error) {
	data, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errors.New("binary data not found")
	}
	return data, nil
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
