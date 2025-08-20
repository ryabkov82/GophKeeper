package model

import (
	"errors"
	"strings"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/forms"
)

// BinaryData представляет произвольные бинарные данные пользователя.
// Содержит путь к зашифрованному файлу в хранилище и дополнительную
// текстовую метаинформацию (также зашифрованную на клиенте).
type BinaryData struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	Title       string    `db:"title"`
	StoragePath string    `db:"storage_path"`
	ClientPath  string    `db:"client_path"`
	Size        int64     `db:"size"`
	Metadata    string    `db:"metadata"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// FormFields возвращает описание полей формы для редактирования BinaryData
func (b *BinaryData) FormFields() []forms.FormField {
	return []forms.FormField{
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

// UpdateFromFields обновляет BinaryData по значениям из формы
func (b *BinaryData) UpdateFromFields(fields []forms.FormField) error {
	if len(fields) != 4 {
		return errors.New("unexpected number of fields")
	}

	if strings.TrimSpace(fields[0].Value) == "" {
		return errors.New("title cannot be empty")
	}

	b.Title = fields[0].Value
	b.Metadata = fields[1].Value
	b.ClientPath = fields[2].Value
	b.UpdatedAt = time.Now()

	return nil
}

// Реализация интерфейса forms.Identifiable
func (b *BinaryData) GetID() string   { return b.ID }
func (b *BinaryData) SetID(id string) { b.ID = id }
