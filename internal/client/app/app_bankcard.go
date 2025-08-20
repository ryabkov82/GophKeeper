package app

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/client/cryptowrap"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto"
)

// ensureBankCardClient гарантирует создание gRPC клиента для BankCard сервиса
func (s *AppServices) ensureBankCardClient(ctx context.Context) error {
	conn, err := s.getGRPCConn(ctx)
	if err != nil {
		return err
	}

	client := proto.NewBankCardServiceClient(conn)
	s.BankCardManager.SetClient(client)
	return nil
}

// CreateBankCard создаёт новую банковскую карту на сервере с шифрованием данных
func (s *AppServices) CreateBankCard(ctx context.Context, card *model.BankCard) error {
	err := s.ensureBankCardClient(ctx)
	if err != nil {
		return err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return err
	}

	wrapper := &cryptowrap.BankcardCryptoWrapper{BankCard: card}
	if err := wrapper.Encrypt(key); err != nil {
		return err
	}

	return s.BankCardManager.CreateBankCard(ctx, card)
}

// GetBankCardByID получает банковскую карту по ID с расшифровкой данных
func (s *AppServices) GetBankCardByID(ctx context.Context, id string) (*model.BankCard, error) {
	err := s.ensureBankCardClient(ctx)
	if err != nil {
		return nil, err
	}

	card, err := s.BankCardManager.GetBankCardByID(ctx, id)
	if err != nil {
		return nil, err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return nil, err
	}

	wrapper := &cryptowrap.BankcardCryptoWrapper{BankCard: card}
	if err := wrapper.Decrypt(key); err != nil {
		return nil, err
	}

	return card, nil
}

// GetBankCards возвращает список банковских карт с расшифровкой данных
func (s *AppServices) GetBankCards(ctx context.Context) ([]model.BankCard, error) {
	err := s.ensureBankCardClient(ctx)
	if err != nil {
		return nil, err
	}

	cards, err := s.BankCardManager.GetBankCards(ctx)
	if err != nil {
		return nil, err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return nil, err
	}

	for i := range cards {
		wrapper := &cryptowrap.BankcardCryptoWrapper{BankCard: &cards[i]}
		if err := wrapper.Decrypt(key); err != nil {
			return nil, err
		}
	}

	return cards, nil
}

// UpdateBankCard обновляет данные банковской карты с шифрованием
func (s *AppServices) UpdateBankCard(ctx context.Context, card *model.BankCard) error {
	err := s.ensureBankCardClient(ctx)
	if err != nil {
		return err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return err
	}

	wrapper := &cryptowrap.BankcardCryptoWrapper{BankCard: card}
	if err := wrapper.Encrypt(key); err != nil {
		return err
	}

	return s.BankCardManager.UpdateBankCard(ctx, card)
}

// DeleteBankCard удаляет банковскую карту по ID
func (s *AppServices) DeleteBankCard(ctx context.Context, id string) error {
	err := s.ensureBankCardClient(ctx)
	if err != nil {
		return err
	}
	return s.BankCardManager.DeleteBankCard(ctx, id)
}
