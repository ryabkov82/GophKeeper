package app

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/client/cryptowrap"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto"
)

// ensureTextDataClient гарантирует создание gRPC клиента для TextData сервиса
func (s *AppServices) ensureTextDataClient(ctx context.Context) error {
	conn, err := s.getGRPCConn(ctx)
	if err != nil {
		return err
	}

	client := proto.NewTextDataServiceClient(conn)
	s.TextDataManager.SetClient(client)
	return nil
}

// CreateTextData создаёт новый текстовый объект на сервере с шифрованием содержимого
func (s *AppServices) CreateTextData(ctx context.Context, text *model.TextData) error {
	if err := s.ensureTextDataClient(ctx); err != nil {
		return err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return err
	}

	wrapper := &cryptowrap.TextDataCryptoWrapper{TextData: text}
	if err := wrapper.Encrypt(key); err != nil {
		return err
	}

	return s.TextDataManager.CreateTextData(ctx, text)
}

// GetTextDataByID получает текстовые данные по ID с расшифровкой содержимого
func (s *AppServices) GetTextDataByID(ctx context.Context, id string) (*model.TextData, error) {
	if err := s.ensureTextDataClient(ctx); err != nil {
		return nil, err
	}

	text, err := s.TextDataManager.GetTextDataByID(ctx, id)
	if err != nil {
		return nil, err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return nil, err
	}

	wrapper := &cryptowrap.TextDataCryptoWrapper{TextData: text}
	if err := wrapper.Decrypt(key); err != nil {
		return nil, err
	}

	return text, nil
}

// GetTextDataTitles получает только заголовки текстовых данных (без расшифровки контента)
func (s *AppServices) GetTextDataTitles(ctx context.Context) ([]*model.TextData, error) {
	if err := s.ensureTextDataClient(ctx); err != nil {
		return nil, err
	}
	return s.TextDataManager.GetTextDataTitles(ctx)
}

// UpdateTextData обновляет текстовые данные с шифрованием содержимого
func (s *AppServices) UpdateTextData(ctx context.Context, text *model.TextData) error {
	if err := s.ensureTextDataClient(ctx); err != nil {
		return err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return err
	}

	wrapper := &cryptowrap.TextDataCryptoWrapper{TextData: text}
	if err := wrapper.Encrypt(key); err != nil {
		return err
	}

	return s.TextDataManager.UpdateTextData(ctx, text)
}

// DeleteTextData удаляет текстовые данные по ID
func (s *AppServices) DeleteTextData(ctx context.Context, id string) error {
	if err := s.ensureTextDataClient(ctx); err != nil {
		return err
	}
	return s.TextDataManager.DeleteTextData(ctx, id)
}
