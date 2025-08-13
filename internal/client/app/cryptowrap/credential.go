package cryptowrap

import (
	"encoding/base64"

	"github.com/ryabkov82/gophkeeper/internal/client/crypto"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// CredentialCryptoWrapper — обёртка для модели Credential,
// предоставляющая методы шифрования и дешифрования данных учётных записей.
type CredentialCryptoWrapper struct {
	*model.Credential
}

// Encrypt шифрует поля Login, Password и Metadata структуры Credential
// с использованием ключа key. Поля заменяются на base64-кодированные
// зашифрованные данные.
//
// Возвращает ошибку, если процесс шифрования завершился неудачей.
func (c *CredentialCryptoWrapper) Encrypt(key []byte) error {
	return EncryptCredential(c.Credential, key)
}

// Decrypt расшифровывает поля Login, Password и Metadata структуры Credential,
// предполагая, что они содержат base64-кодированные зашифрованные данные.
// Расшифрованные значения записываются обратно в поля структуры.
//
// Возвращает ошибку, если процесс дешифрования завершился неудачей.
func (c *CredentialCryptoWrapper) Decrypt(key []byte) error {
	return DecryptCredential(c.Credential, key)
}

// EncryptCredential шифрует данные полей Login, Password и Metadata переданной
// Credential, используя ключ key. Результат кодируется в base64 и записывается
// обратно в соответствующие поля.
//
// Возвращает ошибку при неудаче шифрования любого из полей.
func EncryptCredential(c *model.Credential, key []byte) error {
	encLogin, err := crypto.EncryptAESGCM([]byte(c.Login), key)
	if err != nil {
		return err
	}
	encPassword, err := crypto.EncryptAESGCM([]byte(c.Password), key)
	if err != nil {
		return err
	}
	encMetadata, err := crypto.EncryptAESGCM([]byte(c.Metadata), key)
	if err != nil {
		return err
	}

	c.Login = base64.StdEncoding.EncodeToString(encLogin)
	c.Password = base64.StdEncoding.EncodeToString(encPassword)
	c.Metadata = base64.StdEncoding.EncodeToString(encMetadata)

	return nil
}

// DecryptCredential расшифровывает данные полей Login, Password и Metadata
// переданной Credential, предполагая, что они содержат base64-кодированные
// зашифрованные данные. После расшифровки значения записываются обратно.
//
// Возвращает ошибку при неудаче декодирования base64 или дешифрования данных.
func DecryptCredential(c *model.Credential, key []byte) error {
	decLoginBytes, err := base64.StdEncoding.DecodeString(c.Login)
	if err != nil {
		return err
	}
	decLogin, err := crypto.DecryptAESGCM(decLoginBytes, key)
	if err != nil {
		return err
	}

	decPasswordBytes, err := base64.StdEncoding.DecodeString(c.Password)
	if err != nil {
		return err
	}
	decPassword, err := crypto.DecryptAESGCM(decPasswordBytes, key)
	if err != nil {
		return err
	}

	decMetadataBytes, err := base64.StdEncoding.DecodeString(c.Metadata)
	if err != nil {
		return err
	}
	decMetadata, err := crypto.DecryptAESGCM(decMetadataBytes, key)
	if err != nil {
		return err
	}

	c.Login = string(decLogin)
	c.Password = string(decPassword)
	c.Metadata = string(decMetadata)

	return nil
}
