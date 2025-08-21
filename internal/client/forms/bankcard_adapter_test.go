package forms

import (
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBankCardAdapter_FormFields(t *testing.T) {
	now := time.Now()
	card := &model.BankCard{
		ID:             "test-id",
		UserID:         "user-id",
		Title:          "Test Card",
		CardholderName: "John Doe",
		CardNumber:     "4111111111111111",
		ExpiryDate:     "12/25",
		CVV:            "123",
		Metadata:       "test metadata",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	adapter := &BankCardAdapter{BankCard: card}
	fields := adapter.FormFields()

	assert.Len(t, fields, 7)
	assert.Equal(t, "Title", fields[0].Label)
	assert.Equal(t, "Test Card", fields[0].Value)
	assert.Equal(t, "Cardholder Name", fields[1].Label)
	assert.Equal(t, "John Doe", fields[1].Value)
}

func TestBankCardAdapter_UpdateFromFields(t *testing.T) {
	adapter := &BankCardAdapter{BankCard: &model.BankCard{}}
	fields := []FormField{
		{Value: "Updated Card"},
		{Value: "Jane Doe"},
		{Value: "5555555555554444"},
		{Value: "06/26"},
		{Value: "456"},
		{Value: "updated metadata"},
		{Value: "updated at"},
	}
	require.NoError(t, adapter.UpdateFromFields(fields))
	assert.Equal(t, "Updated Card", adapter.Title)
	assert.Equal(t, "Jane Doe", adapter.CardholderName)
	assert.Equal(t, "5555555555554444", adapter.CardNumber)
	assert.Equal(t, "06/26", adapter.ExpiryDate)
	assert.Equal(t, "456", adapter.CVV)
	assert.Equal(t, "updated metadata", adapter.Metadata)
	assert.False(t, adapter.UpdatedAt.IsZero())
}

func TestBankCardAdapter_UpdateFromFields_Errors(t *testing.T) {
	tests := []struct {
		name    string
		fields  []FormField
		wantErr bool
	}{
		{
			name:    "empty title",
			fields:  []FormField{{Value: ""}, {Value: "Jane"}, {Value: "5555555555554444"}, {Value: "06/26"}, {Value: "456"}, {Value: "m"}, {Value: ""}},
			wantErr: true,
		},
		{
			name:    "invalid card number",
			fields:  []FormField{{Value: "t"}, {Value: "Jane"}, {Value: "5555555555554445"}, {Value: "06/26"}, {Value: "456"}, {Value: "m"}, {Value: ""}},
			wantErr: true,
		},
		{
			name:    "invalid expiry",
			fields:  []FormField{{Value: "t"}, {Value: "Jane"}, {Value: "5555555555554444"}, {Value: "0626"}, {Value: "456"}, {Value: "m"}, {Value: ""}},
			wantErr: true,
		},
		{
			name:    "invalid cvv",
			fields:  []FormField{{Value: "t"}, {Value: "Jane"}, {Value: "5555555555554444"}, {Value: "06/26"}, {Value: "45"}, {Value: "m"}, {Value: ""}},
			wantErr: true,
		},
		{
			name:    "wrong fields count",
			fields:  []FormField{{Value: "t"}, {Value: "Jane"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &BankCardAdapter{BankCard: &model.BankCard{}}
			err := adapter.UpdateFromFields(tt.fields)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
