package jwtauth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithUserIDAndFromContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user123")
	id, err := FromContext(ctx)
	require.NoError(t, err)
	require.Equal(t, "user123", id)
}

func TestFromContextMissing(t *testing.T) {
	_, err := FromContext(context.Background())
	require.Error(t, err)
}
