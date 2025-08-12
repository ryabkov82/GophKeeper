package app_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLoginUser_Success(t *testing.T) {
	authMgr := &mockAuthManager{saltToReturn: []byte("salt")}
	cryptoMgr := &mockCryptoKeyManager{}
	connMgr := &mockConnManager{}

	appSvc := &app.AppServices{
		AuthManager:      authMgr,
		CryptoKeyManager: cryptoMgr,
		ConnManager:      connMgr,
		Logger:           zap.NewNop(),
	}

	err := appSvc.LoginUser(context.Background(), "user", "pass")
	require.NoError(t, err)
	require.True(t, authMgr.setClientCalled)
	require.True(t, authMgr.loginCalled)
	require.True(t, cryptoMgr.generateCalled)
	require.True(t, connMgr.connectCalled)
}

func TestLoginUser_FailEmptySalt(t *testing.T) {
	authMgr := &mockAuthManager{saltToReturn: []byte{}}
	cryptoMgr := &mockCryptoKeyManager{}
	connMgr := &mockConnManager{}

	appSvc := &app.AppServices{
		AuthManager:      authMgr,
		CryptoKeyManager: cryptoMgr,
		ConnManager:      connMgr,
		Logger:           zap.NewNop(),
	}

	err := appSvc.LoginUser(context.Background(), "user", "pass")
	require.ErrorContains(t, err, "no salt received")
}

func TestLoginUser_GenerateKeyError(t *testing.T) {
	authMgr := &mockAuthManager{saltToReturn: []byte("salt")}
	cryptoMgr := &mockCryptoKeyManager{generateErr: fmt.Errorf("fail generate")}
	connMgr := &mockConnManager{}

	appSvc := &app.AppServices{
		AuthManager:      authMgr,
		CryptoKeyManager: cryptoMgr,
		ConnManager:      connMgr,
		Logger:           zap.NewNop(),
	}

	err := appSvc.LoginUser(context.Background(), "user", "pass")
	require.ErrorContains(t, err, "failed to generate encryption key")
}

func TestRegisterUser_Success(t *testing.T) {
	authMgr := &mockAuthManager{}
	cryptoMgr := &mockCryptoKeyManager{}
	connMgr := &mockConnManager{}

	appSvc := &app.AppServices{
		AuthManager:      authMgr,
		CryptoKeyManager: cryptoMgr,
		ConnManager:      connMgr,
		Logger:           zap.NewNop(),
	}

	// Мокаем Register: без ошибок
	authMgr.registerErr = nil
	// LoginUser внутри RegisterUser — используем тот же mockAuthManager, salt возвращаем нормальный
	authMgr.saltToReturn = []byte("salt")

	err := appSvc.RegisterUser(context.Background(), "user", "pass")
	require.NoError(t, err)
	require.True(t, authMgr.setClientCalled)
	require.True(t, authMgr.registerCalled)
	require.True(t, authMgr.loginCalled)
	require.True(t, cryptoMgr.generateCalled)
}

func TestRegisterUser_RegisterFail(t *testing.T) {
	authMgr := &mockAuthManager{registerErr: fmt.Errorf("register failed")}
	connMgr := &mockConnManager{}
	appSvc := &app.AppServices{
		AuthManager: authMgr,
		ConnManager: connMgr,
		Logger:      zap.NewNop(),
	}

	err := appSvc.RegisterUser(context.Background(), "user", "pass")
	require.ErrorContains(t, err, "register failed")
}
