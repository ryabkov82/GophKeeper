package cryptowrap_test

import (
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/cryptowrap"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestBinaryDataCryptoWrapper_EncryptDecrypt(t *testing.T) {
	key := []byte("0123456789ABCDEF0123456789ABCDEF") // 32 байта для AES-256
	originalMetadata := "sensitive info"

	data := &model.BinaryData{
		ID:       "123",
		UserID:   "user1",
		Title:    "My File",
		Metadata: originalMetadata,
	}

	wrapper := &cryptowrap.BinaryDataCryptoWrapper{BinaryData: data}

	// Шифруем Metadata
	err := wrapper.Encrypt(key)
	assert.NoError(t, err)
	assert.NotEqual(t, originalMetadata, data.Metadata, "Metadata должно измениться после шифрования")

	// Дешифруем Metadata
	err = wrapper.Decrypt(key)
	assert.NoError(t, err)
	assert.Equal(t, originalMetadata, data.Metadata, "Metadata после расшифровки должна совпадать с исходной")
}
