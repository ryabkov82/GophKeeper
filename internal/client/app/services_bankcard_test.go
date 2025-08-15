package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/app/cryptowrap"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateBankCard(t *testing.T) {
	ctx := context.Background()

	// Мок менеджера ключей
	mockKeyMgr := &mockCryptoKeyManager{
		loadKeyData: []byte("12345678901234567890123456789012"), // 32 байта для AES-256
	}

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:      &mockConnManager{connectErr: errors.New("connect failed")},
		CryptoKeyManager: mockKeyMgr,
		BankCardManager:  &mockBankCardManager{},
		Logger:           zap.NewNop(),
	}
	err := appSvc.CreateBankCard(ctx, &model.BankCard{})
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешный вызов
	bankCardMgr := &mockBankCardManager{}
	appSvc = &app.AppServices{
		ConnManager:      &mockConnManager{},
		CryptoKeyManager: mockKeyMgr,
		BankCardManager:  bankCardMgr,
		Logger:           zap.NewNop(),
	}
	card := &model.BankCard{
		CardholderName: "IVAN IVANOV",
		CardNumber:     "4111111111111111",
		ExpiryDate:     "12/25",
		CVV:            "123",
		Metadata:       "Primary card",
	}
	err = appSvc.CreateBankCard(ctx, card)
	require.NoError(t, err)
	require.True(t, bankCardMgr.setClientCalled)

	// Проверим, что поля зашифрованы
	require.NotEqual(t, "IVAN IVANOV", card.CardholderName)
	require.NotEqual(t, "4111111111111111", card.CardNumber)
	require.NotEqual(t, "12/25", card.ExpiryDate)
	require.NotEqual(t, "123", card.CVV)
	require.NotEqual(t, "Primary card", card.Metadata)
}

func TestGetBankCardByID(t *testing.T) {
	ctx := context.Background()

	// Создаём BankCard с зашифрованными полями
	plain := &model.BankCard{
		ID:             "123",
		CardholderName: "IVAN IVANOV",
		CardNumber:     "4111111111111111",
		ExpiryDate:     "12/25",
		CVV:            "123",
		Metadata:       "Primary card",
	}
	key := []byte("12345678901234567890123456789012")
	wrapper := &cryptowrap.BankcardCryptoWrapper{BankCard: plain}
	err := wrapper.Encrypt(key)
	require.NoError(t, err)

	// Мок менеджера ключей
	mockKeyMgr := &mockCryptoKeyManager{
		loadKeyData: key,
	}
	// Мок BankCardManager возвращает зашифрованные данные
	mockBankCardMgr := &mockBankCardManager{
		getByIDResult: plain,
	}

	appSvc := &app.AppServices{
		ConnManager:      &mockConnManager{},
		CryptoKeyManager: mockKeyMgr,
		BankCardManager:  mockBankCardMgr,
		Logger:           zap.NewNop(),
	}

	card, err := appSvc.GetBankCardByID(ctx, "123")
	require.NoError(t, err)
	require.Equal(t, "123", card.ID)
	require.Equal(t, "IVAN IVANOV", card.CardholderName) // Проверяем, что расшифровалось
	require.Equal(t, "4111111111111111", card.CardNumber)
	require.Equal(t, "12/25", card.ExpiryDate)
	require.Equal(t, "123", card.CVV)
	require.Equal(t, "Primary card", card.Metadata)

	// Ошибка подключения
	appSvc.ConnManager = &mockConnManager{connectErr: errors.New("connect error")}
	card, err = appSvc.GetBankCardByID(ctx, "123")
	require.Error(t, err)
	require.Nil(t, card)
	require.EqualError(t, err, "connect error")
}

func TestGetBankCards(t *testing.T) {
	ctx := context.Background()

	key := []byte("12345678901234567890123456789012")

	// Создаём список с зашифрованными BankCard
	plainList := []model.BankCard{
		{
			ID:             "1",
			CardholderName: "USER 1",
			CardNumber:     "1111222233334444",
			ExpiryDate:     "01/23",
			CVV:            "111",
			Metadata:       "Card 1",
		},
		{
			ID:             "2",
			CardholderName: "USER 2",
			CardNumber:     "5555666677778888",
			ExpiryDate:     "02/24",
			CVV:            "222",
			Metadata:       "Card 2",
		},
	}

	for i := range plainList {
		wrapper := &cryptowrap.BankcardCryptoWrapper{BankCard: &plainList[i]}
		err := wrapper.Encrypt(key)
		require.NoError(t, err)
	}

	mockKeyMgr := &mockCryptoKeyManager{loadKeyData: key}
	mockBankCardMgr := &mockBankCardManager{getAllResult: plainList}

	appSvc := &app.AppServices{
		ConnManager:      &mockConnManager{},
		CryptoKeyManager: mockKeyMgr,
		BankCardManager:  mockBankCardMgr,
		Logger:           zap.NewNop(),
	}

	cards, err := appSvc.GetBankCards(ctx)
	require.NoError(t, err)
	require.Len(t, cards, 2)

	// Проверяем, что данные расшифровались
	require.Equal(t, "USER 1", cards[0].CardholderName)
	require.Equal(t, "1111222233334444", cards[0].CardNumber)
	require.Equal(t, "01/23", cards[0].ExpiryDate)
	require.Equal(t, "111", cards[0].CVV)
	require.Equal(t, "Card 1", cards[0].Metadata)

	require.Equal(t, "USER 2", cards[1].CardholderName)
	require.Equal(t, "5555666677778888", cards[1].CardNumber)
	require.Equal(t, "02/24", cards[1].ExpiryDate)
	require.Equal(t, "222", cards[1].CVV)
	require.Equal(t, "Card 2", cards[1].Metadata)

	// Ошибка подключения
	appSvc.ConnManager = &mockConnManager{connectErr: errors.New("connect error")}
	cards, err = appSvc.GetBankCards(ctx)
	require.Error(t, err)
	require.Nil(t, cards)
	require.EqualError(t, err, "connect error")
}

func TestUpdateBankCard(t *testing.T) {
	ctx := context.Background()
	mockKeyMgr := &mockCryptoKeyManager{
		loadKeyData: []byte("12345678901234567890123456789012"),
	}

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:      &mockConnManager{connectErr: errors.New("connect failed")},
		CryptoKeyManager: mockKeyMgr,
		BankCardManager:  &mockBankCardManager{},
		Logger:           zap.NewNop(),
	}
	err := appSvc.UpdateBankCard(ctx, &model.BankCard{})
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешное обновление
	bankCardMgr := &mockBankCardManager{}
	appSvc = &app.AppServices{
		ConnManager:      &mockConnManager{},
		CryptoKeyManager: mockKeyMgr,
		BankCardManager:  bankCardMgr,
		Logger:           zap.NewNop(),
	}

	card := &model.BankCard{
		CardholderName: "IVAN IVANOV",
		CardNumber:     "4111111111111111",
		ExpiryDate:     "12/25",
		CVV:            "123",
		Metadata:       "Primary card",
	}
	err = appSvc.UpdateBankCard(ctx, card)
	require.NoError(t, err)
	require.True(t, bankCardMgr.setClientCalled)
	require.NotEqual(t, "IVAN IVANOV", card.CardholderName)
	require.NotEqual(t, "4111111111111111", card.CardNumber)
	require.NotEqual(t, "12/25", card.ExpiryDate)
	require.NotEqual(t, "123", card.CVV)
	require.NotEqual(t, "Primary card", card.Metadata)
}

func TestDeleteBankCard(t *testing.T) {
	ctx := context.Background()

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:     &mockConnManager{connectErr: errors.New("connect failed")},
		BankCardManager: &mockBankCardManager{},
		Logger:          zap.NewNop(),
	}
	err := appSvc.DeleteBankCard(ctx, "id")
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешное удаление
	bankCardMgr := &mockBankCardManager{}
	appSvc = &app.AppServices{
		ConnManager:     &mockConnManager{},
		BankCardManager: bankCardMgr,
		Logger:          zap.NewNop(),
	}
	err = appSvc.DeleteBankCard(ctx, "id")
	require.NoError(t, err)
	require.True(t, bankCardMgr.setClientCalled)
}
