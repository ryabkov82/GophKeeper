package forms

import (
	"errors"
	"strings"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// CredentialAdapter обеспечивает преобразование model.Credential в поля формы и обратно.
type CredentialAdapter struct {
	*model.Credential
}

// FormFields возвращает описание полей формы для редактирования Credential.
func (a *CredentialAdapter) FormFields() []FormField {
	c := a.Credential
	return []FormField{
		{
			Label:       "Title",
			Value:       c.Title,
			MaxLength:   150,
			InputType:   "text",
			Placeholder: "Название (например, Gmail)",
		},
		{
			Label:       "Login",
			Value:       c.Login,
			MaxLength:   50,
			InputType:   "text",
			Placeholder: "Логин/Email",
		},
		{
			Label:       "Password",
			Value:       c.Password,
			MaxLength:   50,
			InputType:   "password",
			Placeholder: "Пароль",
		},
		{
			Label:       "Metadata",
			Value:       c.Metadata,
			InputType:   "multiline",
			Placeholder: "Дополнительные заметки",
		},
		{
			Label:       "UpdatedAt",
			Value:       c.UpdatedAt.String(),
			InputType:   "text",
			ReadOnly:    true,
			Placeholder: "Дата обновления",
		},
	}
}

// UpdateFromFields обновляет Credential по значениям из формы.
func (a *CredentialAdapter) UpdateFromFields(fields []FormField) error {
	if len(fields) != 5 {
		return errors.New("unexpected number of fields")
	}

	if strings.TrimSpace(fields[0].Value) == "" {
		return errors.New("title cannot be empty")
	}

	c := a.Credential
	c.Title = fields[0].Value
	c.Login = fields[1].Value
	c.Password = fields[2].Value
	c.Metadata = fields[3].Value
	c.UpdatedAt = time.Now()
	return nil
}
