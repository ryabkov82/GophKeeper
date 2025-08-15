package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
	"github.com/ryabkov82/gophkeeper/internal/server/storage/postgres"
	"github.com/stretchr/testify/assert"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, repository.BankCardRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	repo := postgres.NewBankCardStorage(db)
	return db, mock, repo
}

func TestBankCardStorage_Create(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	card := &model.BankCard{
		UserID:         uuid.NewString(),
		Title:          "Test Card",
		CardholderName: "John Doe",
		CardNumber:     "1234123412341234",
		ExpiryDate:     "12/30",
		CVV:            "123",
		Metadata:       "{}",
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO bank_cards`)).
		WithArgs(sqlmock.AnyArg(), card.UserID, card.Title, card.CardholderName, card.CardNumber, card.ExpiryDate, card.CVV, card.Metadata).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), card)
	assert.NoError(t, err)
}

func TestBankCardStorage_GetByID(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	id := uuid.NewString()
	card := &model.BankCard{
		ID:             id,
		UserID:         uuid.NewString(),
		Title:          "Test Card",
		CardholderName: "John Doe",
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "cardholder_name"}).
		AddRow(card.ID, card.UserID, card.Title, card.CardholderName)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM bank_cards WHERE id = $1")).
		WithArgs(id).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, card.ID, result.ID)
	assert.Equal(t, card.Title, result.Title)
}

func TestBankCardStorage_GetByUser(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.NewString()
	cards := []model.BankCard{
		{ID: uuid.NewString(), UserID: userID, Title: "Card 1"},
		{ID: uuid.NewString(), UserID: userID, Title: "Card 2"},
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "title"}).
		AddRow(cards[0].ID, cards[0].UserID, cards[0].Title).
		AddRow(cards[1].ID, cards[1].UserID, cards[1].Title)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM bank_cards WHERE user_id = $1 ORDER BY created_at DESC")).
		WithArgs(userID).
		WillReturnRows(rows)

	result, err := repo.GetByUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, cards[0].Title, result[0].Title)
}

func TestBankCardStorage_Update(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	card := &model.BankCard{
		ID:             uuid.NewString(),
		Title:          "Updated",
		CardholderName: "Jane Doe",
		CardNumber:     "4321432143214321",
		ExpiryDate:     "11/29",
		CVV:            "321",
		Metadata:       "{}",
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE bank_cards`)).
		WithArgs(card.Title, card.CardholderName, card.CardNumber, card.ExpiryDate, card.CVV, card.Metadata, card.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), card)
	assert.NoError(t, err)
}

func TestBankCardStorage_Delete(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	id := uuid.NewString()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM bank_cards WHERE id = $1")).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), id)
	assert.NoError(t, err)
}

func TestBankCardStorage_Delete_NotFound(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	id := uuid.NewString()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM bank_cards WHERE id = $1")).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	err := repo.Delete(context.Background(), id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBankCardStorage_Create_Error(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	card := &model.BankCard{
		UserID:         uuid.NewString(),
		Title:          "Test Card",
		CardholderName: "John Doe",
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO bank_cards`)).
		WithArgs(sqlmock.AnyArg(), card.UserID, card.Title, card.CardholderName, card.CardNumber, card.ExpiryDate, card.CVV, card.Metadata).
		WillReturnError(errors.New("insert failed"))

	err := repo.Create(context.Background(), card)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert failed")
}

func TestBankCardStorage_GetByID_NotFound(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	id := uuid.NewString()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM bank_cards WHERE id = $1")).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByID(context.Background(), id)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, result)
}

func TestBankCardStorage_GetByUser_Error(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	userID := uuid.NewString()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM bank_cards WHERE user_id = $1 ORDER BY created_at DESC")).
		WithArgs(userID).
		WillReturnError(errors.New("select failed"))

	result, err := repo.GetByUser(context.Background(), userID)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestBankCardStorage_Update_NotFound(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	card := &model.BankCard{ID: uuid.NewString(), Title: "X"}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE bank_cards`)).
		WithArgs(card.Title, card.CardholderName, card.CardNumber, card.ExpiryDate, card.CVV, card.Metadata, card.ID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	err := repo.Update(context.Background(), card)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBankCardStorage_Update_Error(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	card := &model.BankCard{ID: uuid.NewString(), Title: "X"}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE bank_cards`)).
		WithArgs(card.Title, card.CardholderName, card.CardNumber, card.ExpiryDate, card.CVV, card.Metadata, card.ID).
		WillReturnError(errors.New("update failed"))

	err := repo.Update(context.Background(), card)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
}
