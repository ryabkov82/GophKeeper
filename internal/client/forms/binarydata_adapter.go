package forms

import (
	"errors"
	"strings"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// BinaryDataAdapter обеспечивает преобразование model.BinaryData в поля формы и обратно.
type BinaryDataAdapter struct {
	*model.BinaryData
}

// FormFields возвращает описание полей формы для редактирования BinaryData.
func (a *BinaryDataAdapter) FormFields() []FormField {
	b := a.BinaryData
	return []FormField{
		{
			Label:       "Title",
			Value:       b.Title,
			MaxLength:   150,
			InputType:   "text",
			Placeholder: "Название файла",
		},
		{
			Label:       "Metadata",
			Value:       b.Metadata,
			InputType:   "multiline",
			Placeholder: "Дополнительная информация",
		},
		{
			Label:       "Client Path",
			Value:       b.ClientPath,
			InputType:   "text",
			ReadOnly:    true,
			Placeholder: "Путь к файлу на клиенте",
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

// UpdateFromFields обновляет BinaryData по значениям из формы.
func (a *BinaryDataAdapter) UpdateFromFields(fields []FormField) error {
	if len(fields) != 4 {
		return errors.New("unexpected number of fields")
	}

	if strings.TrimSpace(fields[0].Value) == "" {
		return errors.New("title cannot be empty")
	}

	b := a.BinaryData
	b.Title = fields[0].Value
	b.Metadata = fields[1].Value
	b.ClientPath = fields[2].Value
	b.UpdatedAt = time.Now()
	return nil
}
