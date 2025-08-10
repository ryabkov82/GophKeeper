-- +goose Up
-- Хранение пар логин/пароль с метаинформацией

CREATE TABLE IF NOT EXISTS credentials (
    id UUID PRIMARY KEY,

    -- Владелец записи
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Краткое название (например, "Gmail", "GitHub")
    title TEXT CHECK (char_length(title) <= 255),

    -- Логин пользователя
    login TEXT NOT NULL CHECK (char_length(login) <= 255),

    -- Пароль (в зашифрованном виде, например base64)
    password TEXT NOT NULL CHECK (char_length(password) <= 1024),

    -- Произвольная дополнительная информация
    metadata TEXT CHECK (char_length(metadata) <= 4096),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индекс для ускорения выборок по пользователю
CREATE INDEX IF NOT EXISTS idx_credentials_user_id ON credentials(user_id);

-- Индекс для быстрого поиска по названию внутри пользователя
CREATE INDEX IF NOT EXISTS idx_credentials_user_id_title ON credentials(user_id, title);

-- +goose Down
DROP INDEX IF EXISTS idx_credentials_user_id_title;
DROP INDEX IF EXISTS idx_credentials_user_id;
DROP TABLE IF EXISTS credentials;

