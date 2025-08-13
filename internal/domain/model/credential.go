package model

import (
	"errors"
	"strings"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/client/forms"
)

// Credential представляет собой пару логин/пароль с дополнительной метаинформацией.
// Используется для хранения учётных данных пользователя в зашифрованном виде.
//
// Каждая запись принадлежит конкретному пользователю (UserID) и может содержать:
//   - название или описание (Title), например "Gmail", "GitHub";
//   - логин (Login);
//   - пароль (Password) в зашифрованном виде;
//   - произвольную текстовую метаинформацию (Metadata), например ссылки, заметки,
//     одноразовые коды и т. п.
//
// Поля CreatedAt и UpdatedAt фиксируют время создания и последнего обновления записи.
type Credential struct {
	ID        string // UUID
	UserID    string // Владелец
	Title     string // Метаинформация (например, "Gmail", "GitHub")
	Login     string
	Password  string // Храним в зашифрованном виде
	Metadata  string // Произвольный текст
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Реализация интерфейса tui.FormEntity

// FormFields возвращает описание полей формы для редактирования Credential
func (c *Credential) FormFields() []forms.FormField {
	return []forms.FormField{
		{
			Label:       "Title",
			Value:       c.Title,
			InputType:   "text",
			Placeholder: "Название (например, Gmail)",
		},
		{
			Label:       "Login",
			Value:       c.Login,
			InputType:   "text",
			Placeholder: "Логин/Email",
		},
		{
			Label:       "Password",
			Value:       c.Password,
			InputType:   "password",
			Placeholder: "Пароль",
		},
		{
			Label:       "Metadata",
			Value:       c.Metadata,
			InputType:   "multiline",
			Placeholder: "Дополнительные заметки",
		},
	}
}

// UpdateFromFields обновляет Credential по значениям из формы
func (c *Credential) UpdateFromFields(fields []forms.FormField) error {
	if len(fields) != 4 {
		return errors.New("unexpected number of fields")
	}

	// Можно добавить валидацию по необходимости, например:
	if strings.TrimSpace(fields[0].Value) == "" {
		return errors.New("title cannot be empty")
	}

	c.Title = fields[0].Value
	c.Login = fields[1].Value
	c.Password = fields[2].Value
	c.Metadata = fields[3].Value
	c.UpdatedAt = time.Now()

	return nil
}

func (c *Credential) GetID() string   { return c.ID }
func (c *Credential) SetID(id string) { c.ID = id }
