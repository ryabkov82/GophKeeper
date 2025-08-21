package forms

import (
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextDataAdapter_FormFields(t *testing.T) {
	now := time.Now()
	td := &model.TextData{ID: "1", Title: "Note", Content: []byte("hello"), Metadata: "m", CreatedAt: now, UpdatedAt: now}
	adapter := &TextDataAdapter{TextData: td}
	fields := adapter.FormFields()
	assert.Len(t, fields, 4)
	assert.Equal(t, "Title", fields[0].Label)
	assert.Equal(t, "Note", fields[0].Value)
}

func TestTextDataAdapter_UpdateFromFields(t *testing.T) {
	adapter := &TextDataAdapter{TextData: &model.TextData{}}
	fields := []FormField{
		{Value: "Title"},
		{Value: "Content"},
		{Value: "Meta"},
		{Value: ""},
	}
	require.NoError(t, adapter.UpdateFromFields(fields))
	assert.Equal(t, "Title", adapter.Title)
	assert.Equal(t, "Content", string(adapter.Content))
	assert.Equal(t, "Meta", adapter.Metadata)
	assert.False(t, adapter.UpdatedAt.IsZero())
}

func TestTextDataAdapter_UpdateFromFields_Errors(t *testing.T) {
	adapter := &TextDataAdapter{TextData: &model.TextData{}}
	err := adapter.UpdateFromFields([]FormField{{Value: ""}, {Value: ""}, {Value: ""}, {Value: ""}})
	require.Error(t, err)
}
