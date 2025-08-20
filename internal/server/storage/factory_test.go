package storage

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPostgresFactory(t *testing.T) {
	f := NewPostgresFactory(&sql.DB{})
	require.NotNil(t, f)
	require.NotNil(t, f.User())
	require.NotNil(t, f.Credential())
	require.NotNil(t, f.BankCard())
	require.NotNil(t, f.TextData())
}

func TestNewStorageFactory(t *testing.T) {
	f, err := NewStorageFactory("postgres", &sql.DB{})
	require.NoError(t, err)
	require.NotNil(t, f)

	f, err = NewStorageFactory("unknown", &sql.DB{})
	require.Error(t, err)
	require.Nil(t, f)
}
