package model

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBankCard_ValidateCardNumber(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		wantErr bool
		errText string
	}{
		{
			name:    "valid visa",
			number:  "4111111111111111",
			wantErr: false,
		},
		{
			name:    "valid mastercard",
			number:  "5555555555554444",
			wantErr: false,
		},
		{
			name:    "valid mir",
			number:  "2200244108070333",
			wantErr: false,
		},
		{
			name:    "valid amex",
			number:  "378282246310005",
			wantErr: false,
		},
		{
			name:    "too short",
			number:  "41111111111",
			wantErr: true,
			errText: "card number must contain 12 to 19 digits",
		},
		{
			name:    "too long",
			number:  "41111111111111111111",
			wantErr: true,
			errText: "card number must contain 12 to 19 digits",
		},
		{
			name:    "invalid luhn",
			number:  "4111111111111112",
			wantErr: true,
			errText: "invalid card number (Luhn check failed)",
		},
		{
			name:    "unsupported card",
			number:  "36111111111111",
			wantErr: true,
			errText: "unsupported card type",
		},
		{
			name:    "valid with spaces",
			number:  "4111 1111 1111 1111",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := BankCard{CardNumber: tt.number}
			err := card.ValidateCardNumber()

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.errText, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLuhnCheck(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"valid visa", "4111111111111111", true},
		{"invalid", "4111111111111112", false},
		{"valid mastercard", "5555555555554444", true},
		{"non-digits", "4111abcd11111111", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, luhnCheck(tt.number))
		})
	}
}

func TestCheckPaymentSystem(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"visa", "4111111111111111", true},
		{"mastercard", "5555555555554444", true},
		{"mir", "2200123456789012", true},
		{"amex", "378282246310005", true},
		{"unsupported", "6011111111111111", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, checkPaymentSystem(tt.number))
		})
	}
}

func TestBankCard_IDMethods(t *testing.T) {
	card := BankCard{}
	assert.Empty(t, card.GetID())

	card.SetID("test-id")
	assert.Equal(t, "test-id", card.GetID())
}

func TestValidateExpiryDate(t *testing.T) {

	fixedNow := time.Date(2025, time.August, 1, 0, 0, 0, 0, time.UTC)
	currentYear := fixedNow.Year() % 100
	currentMonth := int(fixedNow.Month())
	nextMonth := currentMonth + 1
	if nextMonth > 12 {
		nextMonth = 1
	}

	tests := []struct {
		name     string
		input    string
		now      time.Time
		expected error
	}{
		// Валидные даты
		{
			name:     "valid current month",
			input:    formatExpiry(currentMonth, currentYear),
			now:      fixedNow,
			expected: nil,
		},
		{
			name:     "valid next month",
			input:    formatExpiry(nextMonth, currentYear),
			now:      fixedNow,
			expected: nil,
		},
		{
			name:     "valid future year",
			input:    formatExpiry(12, currentYear+1),
			now:      fixedNow,
			expected: nil,
		},

		// Неправильный формат
		{
			name:     "empty string",
			input:    "",
			now:      fixedNow,
			expected: errors.New("invalid expiry date format, use MM/YY"),
		},
		{
			name:     "missing slash",
			input:    "1225",
			now:      fixedNow,
			expected: errors.New("invalid expiry date format, use MM/YY"),
		},
		{
			name:     "wrong slash position",
			input:    "1/225",
			expected: errors.New("invalid expiry date format, use MM/YY"),
		},
		{
			name:     "too short",
			input:    "1/5",
			now:      fixedNow,
			expected: errors.New("invalid expiry date format, use MM/YY"),
		},
		{
			name:     "too long",
			input:    "12/2025",
			now:      fixedNow,
			expected: errors.New("invalid expiry date format, use MM/YY"),
		},

		// Неправильные числа
		{
			name:     "invalid month (zero)",
			input:    "00/25",
			now:      fixedNow,
			expected: errors.New("invalid month, must be between 01 and 12"),
		},
		{
			name:     "invalid month (13)",
			input:    "13/25",
			now:      fixedNow,
			expected: errors.New("invalid month, must be between 01 and 12"),
		},
		{
			name:     "non-numeric month",
			input:    "ab/25",
			now:      fixedNow,
			expected: errors.New("month and year must be numbers"),
		},
		{
			name:     "non-numeric year",
			input:    "12/ab",
			now:      fixedNow,
			expected: errors.New("month and year must be numbers"),
		},

		// Просроченные карты
		{
			name:     "expired last month",
			input:    formatExpiry(currentMonth-1, currentYear),
			now:      fixedNow,
			expected: errors.New("card has expired"),
		},
		{
			name:     "expired last year",
			input:    formatExpiry(12, currentYear-1),
			now:      fixedNow,
			expected: errors.New("invalid year, card expired or too far in future"),
		},

		// Слишком далекий год
		{
			name:     "too far in future",
			input:    formatExpiry(12, currentYear+6),
			now:      fixedNow,
			expected: errors.New("invalid year, card expired or too far in future"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Подменяем time.Now на время из теста
			oldNow := timeNow
			timeNow = func() time.Time { return tt.now }
			defer func() { timeNow = oldNow }()

			err := ValidateExpiryDate(tt.input)

			if tt.expected == nil {
				if err != nil {
					t.Errorf("expected nil, got error: %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("expected error %q, got nil", tt.expected.Error())
				return
			}

			if err.Error() != tt.expected.Error() {
				t.Errorf("expected error %q, got %q", tt.expected.Error(), err.Error())
			}
		})
	}
}

// Вспомогательная функция для форматирования даты
func formatExpiry(month, year int) string {
	return fmt.Sprintf("%02d/%02d", month, year)
}
