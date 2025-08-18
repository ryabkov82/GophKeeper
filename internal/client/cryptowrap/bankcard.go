package cryptowrap

import (
	"encoding/base64"

	"github.com/ryabkov82/gophkeeper/internal/client/crypto"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// BankcardCryptoWrapper — обёртка для модели BankCard,
// предоставляющая методы шифрования и дешифрования данных банковских карт.
type BankcardCryptoWrapper struct {
	*model.BankCard
}

// Encrypt шифрует чувствительные поля банковской карты:
// CardholderName, CardNumber, ExpiryDate, CVV и Metadata
// с использованием ключа key. Поля заменяются на base64-кодированные
// зашифрованные данные.
//
// Возвращает ошибку, если процесс шифрования завершился неудачей.
func (b *BankcardCryptoWrapper) Encrypt(key []byte) error {
	return EncryptBankCard(b.BankCard, key)
}

// Decrypt расшифровывает поля банковской карты,
// предполагая, что они содержат base64-кодированные зашифрованные данные.
// Расшифрованные значения записываются обратно в поля структуры.
//
// Возвращает ошибку, если процесс дешифрования завершился неудачей.
func (b *BankcardCryptoWrapper) Decrypt(key []byte) error {
	return DecryptBankCard(b.BankCard, key)
}

// EncryptBankCard шифрует чувствительные данные банковской карты,
// используя ключ key. Результат кодируется в base64 и записывается
// обратно в соответствующие поля.
//
// Возвращает ошибку при неудаче шифрования любого из полей.
func EncryptBankCard(card *model.BankCard, key []byte) error {
	encCardholder, err := crypto.EncryptAESGCM([]byte(card.CardholderName), key)
	if err != nil {
		return err
	}
	encCardNumber, err := crypto.EncryptAESGCM([]byte(card.CardNumber), key)
	if err != nil {
		return err
	}
	encExpiry, err := crypto.EncryptAESGCM([]byte(card.ExpiryDate), key)
	if err != nil {
		return err
	}
	encCVV, err := crypto.EncryptAESGCM([]byte(card.CVV), key)
	if err != nil {
		return err
	}
	encMetadata, err := crypto.EncryptAESGCM([]byte(card.Metadata), key)
	if err != nil {
		return err
	}

	card.CardholderName = base64.StdEncoding.EncodeToString(encCardholder)
	card.CardNumber = base64.StdEncoding.EncodeToString(encCardNumber)
	card.ExpiryDate = base64.StdEncoding.EncodeToString(encExpiry)
	card.CVV = base64.StdEncoding.EncodeToString(encCVV)
	card.Metadata = base64.StdEncoding.EncodeToString(encMetadata)

	return nil
}

// DecryptBankCard расшифровывает данные банковской карты,
// предполагая, что они содержат base64-кодированные зашифрованные данные.
// После расшифровки значения записываются обратно в поля структуры.
//
// Возвращает ошибку при неудаче декодирования base64 или дешифрования данных.
func DecryptBankCard(card *model.BankCard, key []byte) error {
	decCardholderBytes, err := base64.StdEncoding.DecodeString(card.CardholderName)
	if err != nil {
		return err
	}
	decCardholder, err := crypto.DecryptAESGCM(decCardholderBytes, key)
	if err != nil {
		return err
	}

	decCardNumberBytes, err := base64.StdEncoding.DecodeString(card.CardNumber)
	if err != nil {
		return err
	}
	decCardNumber, err := crypto.DecryptAESGCM(decCardNumberBytes, key)
	if err != nil {
		return err
	}

	decExpiryBytes, err := base64.StdEncoding.DecodeString(card.ExpiryDate)
	if err != nil {
		return err
	}
	decExpiry, err := crypto.DecryptAESGCM(decExpiryBytes, key)
	if err != nil {
		return err
	}

	decCVVBytes, err := base64.StdEncoding.DecodeString(card.CVV)
	if err != nil {
		return err
	}
	decCVV, err := crypto.DecryptAESGCM(decCVVBytes, key)
	if err != nil {
		return err
	}

	decMetadataBytes, err := base64.StdEncoding.DecodeString(card.Metadata)
	if err != nil {
		return err
	}
	decMetadata, err := crypto.DecryptAESGCM(decMetadataBytes, key)
	if err != nil {
		return err
	}

	card.CardholderName = string(decCardholder)
	card.CardNumber = string(decCardNumber)
	card.ExpiryDate = string(decExpiry)
	card.CVV = string(decCVV)
	card.Metadata = string(decMetadata)

	return nil
}
