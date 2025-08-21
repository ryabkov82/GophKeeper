package forms

import (
	"errors"
	"strings"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// TextDataAdapter обеспечивает преобразование model.TextData в поля формы и обратно.
type TextDataAdapter struct {
	*model.TextData
}

// FormFields возвращает описание полей формы для редактирования TextData.
func (a *TextDataAdapter) FormFields() []FormField {
	t := a.TextData
	return []FormField{
		{
			Label:       "Title",
			Value:       t.Title,
			MaxLength:   150,
			InputType:   "text",
			Placeholder: "Название заметки",
		},
		{
			Label:       "Content",
			Value:       string(t.Content),
			Fullscreen:  true,
			InputType:   "multiline",
			Placeholder: "Введите текст заметки",
		},
		{
			Label:       "Metadata",
			Value:       t.Metadata,
			InputType:   "multiline",
			Placeholder: "Дополнительная информация",
		},
		{
			Label:       "UpdatedAt",
			Value:       t.UpdatedAt.String(),
			InputType:   "text",
			ReadOnly:    true,
			Placeholder: "Дата обновления",
		},
	}
}

// UpdateFromFields обновляет TextData по значениям из формы.
func (a *TextDataAdapter) UpdateFromFields(fields []FormField) error {
	if len(fields) != 4 {
		return errors.New("unexpected number of fields")
	}

	if strings.TrimSpace(fields[0].Value) == "" {
		return errors.New("title cannot be empty")
	}

	t := a.TextData
	t.Title = fields[0].Value
	t.Content = []byte(fields[1].Value)
	t.Metadata = fields[2].Value
	t.UpdatedAt = time.Now()
	return nil
}
