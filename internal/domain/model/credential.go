package model

import (
	"time"
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

// GetID возвращает идентификатор учётных данных.
func (c *Credential) GetID() string { return c.ID }

// SetID устанавливает идентификатор учётных данных.
func (c *Credential) SetID(id string) { c.ID = id }
