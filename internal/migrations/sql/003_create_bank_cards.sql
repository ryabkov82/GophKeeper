-- +goose Up
CREATE TABLE IF NOT EXISTS bank_cards (
    id UUID PRIMARY KEY,

    -- Владелец записи
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Краткое название карты (например, "Visa Сбербанк", "MasterCard Tinkoff")
    title TEXT CHECK (char_length(title) <= 255),

    -- Имя владельца карты (в зашифрованном виде)
    cardholder_name TEXT NOT NULL CHECK (char_length(cardholder_name) <= 1024),

    -- Номер карты (в зашифрованном виде)
    card_number TEXT NOT NULL CHECK (char_length(card_number) <= 1024),

    -- Срок действия (MM/YY, в зашифрованном виде)
    expiry_date TEXT NOT NULL CHECK (char_length(expiry_date) <= 1024),

    -- CVV (в зашифрованном виде)
    cvv TEXT NOT NULL CHECK (char_length(cvv) <= 1024),

    -- Произвольная дополнительная информация (в зашифрованном виде)
    metadata TEXT CHECK (char_length(metadata) <= 4096),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индекс для ускорения выборок по пользователю
CREATE INDEX IF NOT EXISTS idx_bank_cards_user_id ON bank_cards(user_id);

-- Индекс для быстрого поиска по названию внутри пользователя
CREATE INDEX IF NOT EXISTS idx_bank_cards_user_id_title ON bank_cards(user_id, title);

-- +goose Down
DROP INDEX IF EXISTS idx_bank_cards_user_id_title;
DROP INDEX IF EXISTS idx_bank_cards_user_id;
DROP TABLE IF EXISTS bank_cards;
