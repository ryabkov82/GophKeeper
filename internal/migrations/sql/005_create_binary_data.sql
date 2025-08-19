-- +goose Up
CREATE TABLE IF NOT EXISTS binary_data (
    id UUID PRIMARY KEY,

    -- Владелец записи
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Заголовок записи (например, "Фото паспорта", "Ключ для SSH")
    title TEXT CHECK (char_length(title) <= 255),

    -- Путь к исходному файлу на стороне клиента
    client_path TEXT NOT NULL CHECK (char_length(client_path) <= 1024),
    
    -- Путь к файлу в хранилище (локальная ФС, S3 и т.п.)
    storage_path TEXT NOT NULL CHECK (char_length(storage_path) <= 1024),

    -- Размер файла в байтах
    size BIGINT NOT NULL DEFAULT 0,

    -- Произвольная дополнительная информация (в зашифрованном виде)
    metadata TEXT CHECK (char_length(metadata) <= 4096),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индекс для ускорения выборок по пользователю
CREATE INDEX IF NOT EXISTS idx_binary_data_user_id ON binary_data(user_id);

-- Индекс для быстрого поиска по заголовку внутри пользователя
CREATE INDEX IF NOT EXISTS idx_binary_data_user_id_title ON binary_data(user_id, title);

-- +goose Down
DROP INDEX IF EXISTS idx_binary_data_user_id_title;
DROP INDEX IF EXISTS idx_binary_data_user_id;
DROP TABLE IF EXISTS binary_data;
