-- +goose Up
CREATE TABLE IF NOT EXISTS text_data (
    id UUID PRIMARY KEY,

    -- Владелец записи
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Заголовок записи (например, "Рабочие заметки", "Пароли от сайта")
    title TEXT CHECK (char_length(title) <= 255),

    -- Основной контент (в зашифрованном виде)
    content BYTEA NOT NULL,

    -- Произвольная дополнительная информация (в зашифрованном виде)
    metadata TEXT CHECK (char_length(metadata) <= 4096),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индекс для ускорения выборок по пользователю
CREATE INDEX IF NOT EXISTS idx_text_data_user_id ON text_data(user_id);

-- Индекс для быстрого поиска по заголовку внутри пользователя
CREATE INDEX IF NOT EXISTS idx_text_data_user_id_title ON text_data(user_id, title);

-- +goose Down
DROP INDEX IF EXISTS idx_text_data_user_id_title;
DROP INDEX IF EXISTS idx_text_data_user_id;
DROP TABLE IF EXISTS text_data;
