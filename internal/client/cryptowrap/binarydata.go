package cryptowrap

import (
	"encoding/base64"

	"github.com/ryabkov82/gophkeeper/internal/client/crypto"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// BinaryDataCryptoWrapper — обёртка для модели BinaryData,
// предоставляющая методы шифрования и дешифрования только Metadata.
type BinaryDataCryptoWrapper struct {
	*model.BinaryData
}

// Encrypt шифрует только Metadata и кодирует в Base64.
func (b *BinaryDataCryptoWrapper) Encrypt(key []byte) error {

	encMetadata, err := crypto.EncryptAESGCM([]byte(b.Metadata), key)
	if err != nil {
		return err
	}

	encClientPath, err := crypto.EncryptAESGCM([]byte(b.ClientPath), key)
	if err != nil {
		return err
	}

	b.Metadata = base64.StdEncoding.EncodeToString(encMetadata)
	b.ClientPath = base64.StdEncoding.EncodeToString(encClientPath)
	return nil
}

// Decrypt расшифровывает Metadata из Base64.
func (b *BinaryDataCryptoWrapper) Decrypt(key []byte) error {
	encMetadataBytes, err := base64.StdEncoding.DecodeString(b.Metadata)
	if err != nil {
		return err
	}
	decMetadata, err := crypto.DecryptAESGCM(encMetadataBytes, key)
	if err != nil {
		return err
	}
	b.Metadata = string(decMetadata)

	encClientPath, err := base64.StdEncoding.DecodeString(b.ClientPath)
	if err != nil {
		return err
	}
	decClientPath, err := crypto.DecryptAESGCM(encClientPath, key)
	if err != nil {
		return err
	}
	b.ClientPath = string(decClientPath)

	return nil
}
