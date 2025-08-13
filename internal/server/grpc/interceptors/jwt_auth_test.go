package interceptors_test

import (
	"context"
	"testing"
	"time"

	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtauth"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtutils"
	"github.com/ryabkov82/gophkeeper/internal/server/grpc/interceptors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryAuthInterceptor(t *testing.T) {
	secret := "testsecret"
	tm := jwtutils.New(secret, 10*time.Minute)

	userID := "user123"
	login := "login"

	tokenStr, err := tm.GenerateToken(userID, login)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	interceptor := interceptors.UnaryAuthInterceptor(tm, zap.NewNop())

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		uid, err := jwtauth.FromContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, userID, uid)
		return "ok", nil
	}

	t.Run("success", func(t *testing.T) {
		md := metadata.Pairs("authorization", "Bearer "+tokenStr)
		ctx := metadata.NewIncomingContext(context.Background(), md)

		resp, err := interceptor(ctx, "req", &grpc.UnaryServerInfo{
			FullMethod: "/test.Method",
		}, handler)

		assert.NoError(t, err)
		assert.Equal(t, "ok", resp)
		assert.True(t, handlerCalled)
	})

	t.Run("missing metadata", func(t *testing.T) {
		ctx := context.Background()

		_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("missing authorization header", func(t *testing.T) {
		md := metadata.Pairs("some-header", "value")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("invalid authorization header format", func(t *testing.T) {
		md := metadata.Pairs("authorization", "InvalidFormat")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("invalid token", func(t *testing.T) {
		md := metadata.Pairs("authorization", "Bearer invalidtoken")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("missing userID in token claims", func(t *testing.T) {
		token := jwtutils.New(secret, 0)
		tokenStr, err := token.GenerateToken("", login)
		assert.NoError(t, err)

		md := metadata.Pairs("authorization", "Bearer "+tokenStr)
		ctx := metadata.NewIncomingContext(context.Background(), md)

		_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
		assert.Error(t, err)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	t.Run("public method bypasses auth", func(t *testing.T) {
		handlerCalled = false

		// Предполагаем, что метод /public.Method считается публичным в isPublicMethod
		resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
			FullMethod: "/gophkeeper.proto.AuthService/Register",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			handlerCalled = true
			return "public_ok", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "public_ok", resp)
		assert.True(t, handlerCalled)
	})
}
