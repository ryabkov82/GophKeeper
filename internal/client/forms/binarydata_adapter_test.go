package forms

import (
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinaryDataAdapter_FormFields(t *testing.T) {
	now := time.Now()
	bd := &model.BinaryData{ID: "1", Title: "File", ClientPath: "/tmp/a", Metadata: "m", CreatedAt: now, UpdatedAt: now}
	adapter := &BinaryDataAdapter{BinaryData: bd}
	fields := adapter.FormFields()
	assert.Len(t, fields, 4)
	assert.Equal(t, "Title", fields[0].Label)
	assert.Equal(t, "File", fields[0].Value)
}

func TestBinaryDataAdapter_UpdateFromFields(t *testing.T) {
	adapter := &BinaryDataAdapter{BinaryData: &model.BinaryData{}}
	fields := []FormField{
		{Value: "Title"},
		{Value: "Meta"},
		{Value: "/tmp"},
		{Value: ""},
	}
	require.NoError(t, adapter.UpdateFromFields(fields))
	assert.Equal(t, "Title", adapter.Title)
	assert.Equal(t, "Meta", adapter.Metadata)
	assert.Equal(t, "/tmp", adapter.ClientPath)
	assert.False(t, adapter.UpdatedAt.IsZero())
}

func TestBinaryDataAdapter_UpdateFromFields_Errors(t *testing.T) {
	adapter := &BinaryDataAdapter{BinaryData: &model.BinaryData{}}
	err := adapter.UpdateFromFields([]FormField{{Value: ""}, {Value: ""}, {Value: ""}, {Value: ""}})
	require.Error(t, err)
}
