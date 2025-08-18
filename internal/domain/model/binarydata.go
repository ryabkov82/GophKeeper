package model

import (
	"time"
)

// BinaryData представляет произвольные бинарные данные пользователя.
// Содержит путь к зашифрованному файлу в хранилище и дополнительную
// текстовую метаинформацию (также зашифрованную на клиенте).
type BinaryData struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	Title       string    `db:"title"`
	StoragePath string    `db:"storage_path"`
	Size        int64     `db:"size"`
	Metadata    string    `db:"metadata"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
