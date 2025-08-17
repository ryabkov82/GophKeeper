package cryptowrap

import (
	"encoding/base64"

	"github.com/ryabkov82/gophkeeper/internal/client/crypto"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// TextDataCryptoWrapper — обёртка для модели TextData,
// предоставляющая методы шифрования и дешифрования данных.
type TextDataCryptoWrapper struct {
	*model.TextData
}

// Encrypt шифрует поля TextData:
// Content ([]byte) и Metadata (string).
// Content шифруется напрямую, Metadata шифруется и кодируется в Base64.
func (t *TextDataCryptoWrapper) Encrypt(key []byte) error {
	return EncryptTextData(t.TextData, key)
}

// Decrypt расшифровывает поля TextData:
// Content ([]byte) и Metadata (string с Base64).
func (t *TextDataCryptoWrapper) Decrypt(key []byte) error {
	return DecryptTextData(t.TextData, key)
}

// EncryptTextData шифрует Content и Metadata.
// Content хранится как []byte, Metadata — Base64.
func EncryptTextData(td *model.TextData, key []byte) error {
	encContent, err := crypto.EncryptAESGCM(td.Content, key)
	if err != nil {
		return err
	}
	td.Content = encContent

	encMetadata, err := crypto.EncryptAESGCM([]byte(td.Metadata), key)
	if err != nil {
		return err
	}
	td.Metadata = base64.StdEncoding.EncodeToString(encMetadata)

	return nil
}

// DecryptTextData расшифровывает Content и Metadata.
// Content хранится как []byte, Metadata декодируется из Base64.
func DecryptTextData(td *model.TextData, key []byte) error {
	decContent, err := crypto.DecryptAESGCM(td.Content, key)
	if err != nil {
		return err
	}
	td.Content = decContent

	encMetadataBytes, err := base64.StdEncoding.DecodeString(td.Metadata)
	if err != nil {
		return err
	}
	decMetadata, err := crypto.DecryptAESGCM(encMetadataBytes, key)
	if err != nil {
		return err
	}
	td.Metadata = string(decMetadata)

	return nil
}
