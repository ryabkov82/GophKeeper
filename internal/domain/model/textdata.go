package model

import (
	"errors"
	"strings"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/forms"
)

// TextData — модель для хранения произвольных текстовых данных.
// Все чувствительные поля (Content и Metadata) должны храниться в зашифрованном виде (например, base64 или raw bytes).
type TextData struct {
	ID        string    `db:"id"`         // Уникальный идентификатор записи (UUID)
	UserID    string    `db:"user_id"`    // Идентификатор пользователя-владельца записи
	Title     string    `db:"title"`      // Краткое название записи (например, "Рабочие заметки")
	Content   []byte    `db:"content"`    // Основной зашифрованный контент
	Metadata  string    `db:"metadata"`   // Дополнительные данные в формате JSON или свободный текст, зашифрованные
	CreatedAt time.Time `db:"created_at"` // Время создания записи
	UpdatedAt time.Time `db:"updated_at"` // Время последнего обновления записи
}

// Реализация интерфейса tui.FormEntity для TextData

// FormFields возвращает описание полей формы для редактирования TextData
func (t *TextData) FormFields() []forms.FormField {
	return []forms.FormField{
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

// UpdateFromFields обновляет TextData по значениям из формы
func (t *TextData) UpdateFromFields(fields []forms.FormField) error {
	if len(fields) != 4 {
		return errors.New("unexpected number of fields")
	}

	// Валидация: Title обязателен
	if strings.TrimSpace(fields[0].Value) == "" {
		return errors.New("title cannot be empty")
	}

	// Обновляем поля
	t.Title = fields[0].Value
	t.Content = []byte(fields[1].Value)
	t.Metadata = fields[2].Value
	t.UpdatedAt = time.Now()

	return nil
}

func (t *TextData) GetID() string   { return t.ID }
func (t *TextData) SetID(id string) { t.ID = id }
