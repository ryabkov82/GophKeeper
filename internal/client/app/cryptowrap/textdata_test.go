package cryptowrap_test

import (
	"encoding/base64"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/app/cryptowrap"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextDataCryptoWrapper(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 байта AES-256

	original := &model.TextData{
		Title:    "Test Note",
		Content:  []byte("Secret content"),
		Metadata: "Important note",
	}

	t.Run("Encrypt/Decrypt with wrapper", func(t *testing.T) {
		td := &model.TextData{
			Title:    original.Title,
			Content:  original.Content,
			Metadata: original.Metadata,
		}

		wrapper := cryptowrap.TextDataCryptoWrapper{td}

		err := wrapper.Encrypt(key)
		require.NoError(t, err)

		// Content изменился
		assert.NotEqual(t, original.Content, td.Content)
		// Metadata изменился и закодирован в Base64
		assert.NotEqual(t, original.Metadata, td.Metadata)
		_, err = base64.StdEncoding.DecodeString(td.Metadata)
		assert.NoError(t, err)

		// Дешифруем
		err = wrapper.Decrypt(key)
		require.NoError(t, err)

		assert.Equal(t, original.Title, td.Title)
		assert.Equal(t, original.Content, td.Content)
		assert.Equal(t, original.Metadata, td.Metadata)
	})

	t.Run("EncryptTextData/DecryptTextData standalone", func(t *testing.T) {
		td := &model.TextData{
			Title:    original.Title,
			Content:  original.Content,
			Metadata: original.Metadata,
		}

		err := cryptowrap.EncryptTextData(td, key)
		require.NoError(t, err)

		assertEncryptedBytes(t, original.Content, td.Content)
		assertEncryptedField(t, original.Metadata, td.Metadata)

		err = cryptowrap.DecryptTextData(td, key)
		require.NoError(t, err)

		assert.Equal(t, original.Content, td.Content)
		assert.Equal(t, original.Metadata, td.Metadata)
	})

	t.Run("Decrypt with invalid base64 metadata", func(t *testing.T) {
		td := &model.TextData{
			Metadata: "invalid-base64",
			Content:  original.Content,
		}

		err := cryptowrap.DecryptTextData(td, key)
		assert.Error(t, err)
	})

	t.Run("Decrypt with invalid ciphertext", func(t *testing.T) {
		td := &model.TextData{
			Metadata: base64.StdEncoding.EncodeToString([]byte("invalid-cipher")),
			Content:  []byte("invalid-cipher"),
		}

		err := cryptowrap.DecryptTextData(td, key)
		assert.Error(t, err)
	})

	t.Run("Encrypt with wrong key size", func(t *testing.T) {
		td := &model.TextData{
			Content: []byte("some content"),
		}
		invalidKey := []byte("short-key")
		err := cryptowrap.EncryptTextData(td, invalidKey)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid AES key size")
	})
}

// Вспомогательная проверка зашифрованных []byte
func assertEncryptedBytes(t *testing.T, original, encrypted []byte) {
	t.Helper()
	assert.NotEqual(t, original, encrypted)
	assert.NotEmpty(t, encrypted)
}
