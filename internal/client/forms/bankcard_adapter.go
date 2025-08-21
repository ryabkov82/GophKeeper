package forms

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// BankCardAdapter обеспечивает преобразование model.BankCard в поля формы и обратно.
type BankCardAdapter struct {
	*model.BankCard
}

// FormFields возвращает описание полей формы для редактирования BankCard.
func (a *BankCardAdapter) FormFields() []FormField {
	b := a.BankCard
	return []FormField{
		{
			Label:       "Title",
			Value:       b.Title,
			MaxLength:   150,
			InputType:   "text",
			Placeholder: "Название карты (например, Основная карта)",
		},
		{
			Label:       "Cardholder Name",
			Value:       b.CardholderName,
			MaxLength:   150,
			InputType:   "text",
			Placeholder: "Имя держателя карты",
		},
		{
			Label:       "Card Number",
			Value:       b.CardNumber,
			InputType:   "text",
			Placeholder: "Номер карты (XXXX XXXX XXXX XXXX)",
			Mask:        "#### #### #### ####",
		},
		{
			Label:       "Expiry Date",
			Value:       b.ExpiryDate,
			InputType:   "text",
			Placeholder: "MM/YY",
			Mask:        "##/##",
		},
		{
			Label:       "CVV",
			Value:       b.CVV,
			InputType:   "password",
			Placeholder: "XXX",
			Mask:        "###",
		},
		{
			Label:       "Metadata",
			Value:       b.Metadata,
			InputType:   "multiline",
			Fullscreen:  false,
			Placeholder: "Дополнительная информация о карте",
		},
		{
			Label:       "UpdatedAt",
			Value:       b.UpdatedAt.String(),
			InputType:   "text",
			ReadOnly:    true,
			Placeholder: "Дата обновления",
		},
	}
}

// UpdateFromFields обновляет BankCard по значениям из формы.
func (a *BankCardAdapter) UpdateFromFields(fields []FormField) error {
	if len(fields) != 7 {
		return errors.New("unexpected number of fields")
	}

	if strings.TrimSpace(fields[0].Value) == "" {
		return errors.New("title cannot be empty")
	}

	cardNumber := strings.ReplaceAll(fields[2].Value, " ", "")
	tmp := &model.BankCard{CardNumber: cardNumber}
	if err := tmp.ValidateCardNumber(); err != nil {
		return err
	}

	expiryDate := fields[3].Value
	if err := model.ValidateExpiryDate(expiryDate); err != nil {
		return err
	}

	cvv := fields[4].Value
	if _, err := strconv.Atoi(cvv); err != nil || len(cvv) != 3 {
		return errors.New("CVV must be 3 digits")
	}

	b := a.BankCard
	b.Title = fields[0].Value
	b.CardholderName = fields[1].Value
	b.CardNumber = cardNumber
	b.ExpiryDate = expiryDate
	b.CVV = fields[4].Value
	b.Metadata = fields[5].Value
	b.UpdatedAt = time.Now()
	return nil
}
