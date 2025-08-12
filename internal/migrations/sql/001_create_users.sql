-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Логин пользователя, уникальный, ограничен по длине
    login TEXT NOT NULL UNIQUE CHECK (char_length(login) <= 255),

    -- Хэш пароля (например, bcrypt, argon2) в base64/hex
    password_hash TEXT NOT NULL CHECK (char_length(password_hash) <= 1024),

    -- Соль для хэширования
    salt TEXT NOT NULL CHECK (char_length(salt) <= 255)
);

-- +goose Down
DROP TABLE IF EXISTS users;