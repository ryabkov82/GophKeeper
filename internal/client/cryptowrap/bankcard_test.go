package cryptowrap_test

import (
	"encoding/base64"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/cryptowrap"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBankcardCryptoWrapper(t *testing.T) {
	// Генерируем тестовый ключ шифрования
	key := []byte("12345678901234567890123456789012")

	// Подготовка тестовых данных
	originalCard := &model.BankCard{
		CardholderName: "IVAN IVANOV",
		CardNumber:     "4111111111111111",
		ExpiryDate:     "12/25",
		CVV:            "123",
		Metadata:       "Primary card",
	}

	t.Run("Encrypt/Decrypt with wrapper", func(t *testing.T) {
		card := &model.BankCard{
			CardholderName: originalCard.CardholderName,
			CardNumber:     originalCard.CardNumber,
			ExpiryDate:     originalCard.ExpiryDate,
			CVV:            originalCard.CVV,
			Metadata:       originalCard.Metadata,
		}

		wrapper := cryptowrap.BankcardCryptoWrapper{card}

		// Шифруем данные
		err := wrapper.Encrypt(key)
		require.NoError(t, err)

		// Проверяем, что данные зашифрованы и закодированы в base64
		assert.NotEqual(t, originalCard.CardholderName, card.CardholderName)
		assert.NotEqual(t, originalCard.CardNumber, card.CardNumber)
		assert.NotEqual(t, originalCard.ExpiryDate, card.ExpiryDate)
		assert.NotEqual(t, originalCard.CVV, card.CVV)
		assert.NotEqual(t, originalCard.Metadata, card.Metadata)

		_, err = base64.StdEncoding.DecodeString(card.CardholderName)
		assert.NoError(t, err)
		_, err = base64.StdEncoding.DecodeString(card.CardNumber)
		assert.NoError(t, err)

		// Дешифруем данные
		err = wrapper.Decrypt(key)
		require.NoError(t, err)

		// Проверяем, что данные вернулись к исходным значениям
		assert.Equal(t, originalCard.CardholderName, card.CardholderName)
		assert.Equal(t, originalCard.CardNumber, card.CardNumber)
		assert.Equal(t, originalCard.ExpiryDate, card.ExpiryDate)
		assert.Equal(t, originalCard.CVV, card.CVV)
		assert.Equal(t, originalCard.Metadata, card.Metadata)
	})

	t.Run("EncryptBankCard/DecryptBankCard standalone", func(t *testing.T) {
		card := &model.BankCard{
			CardholderName: originalCard.CardholderName,
			CardNumber:     originalCard.CardNumber,
			ExpiryDate:     originalCard.ExpiryDate,
			CVV:            originalCard.CVV,
			Metadata:       originalCard.Metadata,
		}

		// Шифруем данные
		err := cryptowrap.EncryptBankCard(card, key)
		require.NoError(t, err)

		// Проверяем зашифрованные данные
		assertEncryptedField(t, originalCard.CardholderName, card.CardholderName)
		assertEncryptedField(t, originalCard.CardNumber, card.CardNumber)
		assertEncryptedField(t, originalCard.ExpiryDate, card.ExpiryDate)
		assertEncryptedField(t, originalCard.CVV, card.CVV)
		assertEncryptedField(t, originalCard.Metadata, card.Metadata)

		// Дешифруем данные
		err = cryptowrap.DecryptBankCard(card, key)
		require.NoError(t, err)

		// Проверяем исходные данные
		assert.Equal(t, originalCard.CardholderName, card.CardholderName)
		assert.Equal(t, originalCard.CardNumber, card.CardNumber)
		assert.Equal(t, originalCard.ExpiryDate, card.ExpiryDate)
		assert.Equal(t, originalCard.CVV, card.CVV)
		assert.Equal(t, originalCard.Metadata, card.Metadata)
	})

	t.Run("Decrypt with invalid base64", func(t *testing.T) {
		card := &model.BankCard{
			CardholderName: "invalid-base64",
			CardNumber:     "invalid-base64",
		}

		err := cryptowrap.DecryptBankCard(card, key)
		assert.Error(t, err, "Ожидалась ошибка при расшифровке невалидных данных")
	})

	t.Run("Decrypt with invalid ciphertext", func(t *testing.T) {
		invalidCipher := base64.StdEncoding.EncodeToString([]byte("invalid-cipher"))

		card := &model.BankCard{
			CardholderName: invalidCipher,
			CardNumber:     invalidCipher,
		}

		err := cryptowrap.DecryptBankCard(card, key)
		assert.Error(t, err, "Ожидалась ошибка при расшифровке невалидных данных")
	})

	t.Run("Encrypt with wrong key size", func(t *testing.T) {
		invalidKey := []byte("short-key")
		card := &model.BankCard{
			CardNumber: "4111111111111111",
		}

		err := cryptowrap.EncryptBankCard(card, invalidKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid AES key size")
	})
}

// Вспомогательная функция для проверки зашифрованных полей
func assertEncryptedField(t *testing.T, original, encrypted string) {
	t.Helper()

	if original == "" {
		assert.Empty(t, encrypted)
		return
	}

	assert.NotEqual(t, original, encrypted)

	// Проверяем, что зашифрованные данные являются валидным base64
	decoded, err := base64.StdEncoding.DecodeString(encrypted)
	assert.NoError(t, err)
	assert.NotEmpty(t, decoded)
}
