package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
)

type bankCardStorage struct {
	db *sqlx.DB
}

// NewBankCardStorage создаёт новое хранилище банковских карт.
// Принимает стандартный *sql.DB и оборачивает его в *sqlx.DB.
func NewBankCardStorage(db *sql.DB) repository.BankCardRepository {
	sqlxDB := sqlx.NewDb(db, "pgx")
	return &bankCardStorage{db: sqlxDB}
}

func (s *bankCardStorage) Create(ctx context.Context, card *model.BankCard) error {
	card.ID = uuid.NewString()
	query := `
		INSERT INTO bank_cards (
			id, user_id, title, cardholder_name, card_number, expiry_date, cvv, metadata, created_at, updated_at
		) VALUES (
			:id, :user_id, :title, :cardholder_name, :card_number, :expiry_date, :cvv, :metadata, NOW(), NOW()
		)`
	_, err := s.db.NamedExecContext(ctx, query, card)
	return err
}

func (s *bankCardStorage) GetByID(ctx context.Context, id string) (*model.BankCard, error) {
	var card model.BankCard
	err := s.db.GetContext(ctx, &card, "SELECT * FROM bank_cards WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (s *bankCardStorage) GetByUser(ctx context.Context, userID string) ([]model.BankCard, error) {
	var cards []model.BankCard
	err := s.db.SelectContext(ctx, &cards, "SELECT * FROM bank_cards WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	return cards, nil
}

func (s *bankCardStorage) Update(ctx context.Context, card *model.BankCard) error {
	query := `
		UPDATE bank_cards
		SET title = :title,
		    cardholder_name = :cardholder_name,
		    card_number = :card_number,
		    expiry_date = :expiry_date,
		    cvv = :cvv,
		    metadata = :metadata,
		    updated_at = NOW()
		WHERE id = :id`
	res, err := s.db.NamedExecContext(ctx, query, card)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("bank card with id %s not found", card.ID)
	}
	return nil
}

func (s *bankCardStorage) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM bank_cards WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("bank card with id %s not found", id)
	}
	return nil
}
