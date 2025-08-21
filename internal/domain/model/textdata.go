package model

import (
	"time"
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

// GetID возвращает идентификатор текстовых данных.
func (t *TextData) GetID() string { return t.ID }

// SetID устанавливает идентификатор текстовых данных.
func (t *TextData) SetID(id string) { t.ID = id }
