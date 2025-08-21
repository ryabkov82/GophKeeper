package forms

import (
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialAdapter_FormFields(t *testing.T) {
	now := time.Now()
	cred := &model.Credential{
		ID:        "1",
		Title:     "Gmail",
		Login:     "user",
		Password:  "pass",
		Metadata:  "meta",
		CreatedAt: now,
		UpdatedAt: now,
	}
	adapter := &CredentialAdapter{Credential: cred}
	fields := adapter.FormFields()
	assert.Len(t, fields, 5)
	assert.Equal(t, "Title", fields[0].Label)
	assert.Equal(t, "Gmail", fields[0].Value)
}

func TestCredentialAdapter_UpdateFromFields(t *testing.T) {
	adapter := &CredentialAdapter{Credential: &model.Credential{}}
	fields := []FormField{
		{Value: "GitHub"},
		{Value: "login"},
		{Value: "password"},
		{Value: "meta"},
		{Value: ""},
	}
	require.NoError(t, adapter.UpdateFromFields(fields))
	assert.Equal(t, "GitHub", adapter.Title)
	assert.Equal(t, "login", adapter.Login)
	assert.Equal(t, "password", adapter.Password)
	assert.Equal(t, "meta", adapter.Metadata)
	assert.False(t, adapter.UpdatedAt.IsZero())
}

func TestCredentialAdapter_UpdateFromFields_Errors(t *testing.T) {
	adapter := &CredentialAdapter{Credential: &model.Credential{}}
	err := adapter.UpdateFromFields([]FormField{{Value: ""}})
	require.Error(t, err)
}
