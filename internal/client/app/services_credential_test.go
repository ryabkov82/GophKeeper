package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/app/cryptowrap"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateCredential(t *testing.T) {
	ctx := context.Background()

	// Мок менеджера ключей — возвращает фиктивный ключ
	mockKeyMgr := &mockCryptoKeyManager{
		loadKeyData: []byte("12345678901234567890123456789012"), // 32 байта для AES-256
	}

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:       &mockConnManager{connectErr: errors.New("connect failed")},
		CryptoKeyManager:  mockKeyMgr,
		CredentialManager: &mockCredentialManager{},
		Logger:            zap.NewNop(),
	}
	err := appSvc.CreateCredential(ctx, &model.Credential{})
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешный вызов
	credMgr := &mockCredentialManager{}
	appSvc = &app.AppServices{
		ConnManager:       &mockConnManager{},
		CryptoKeyManager:  mockKeyMgr,
		CredentialManager: credMgr,
		Logger:            zap.NewNop(),
	}
	cred := &model.Credential{
		Login:    "user",
		Password: "pass",
		Metadata: "data",
	}
	err = appSvc.CreateCredential(ctx, cred)
	require.NoError(t, err)
	require.True(t, credMgr.setClientCalled)

	// Проверим, что поля зашифрованы (не равны исходным)
	require.NotEqual(t, "user", cred.Login)
	require.NotEqual(t, "pass", cred.Password)
	require.NotEqual(t, "data", cred.Metadata)
}

func TestGetCredentialByID(t *testing.T) {
	ctx := context.Background()

	// Создаём Credential с зашифрованными полями
	plain := &model.Credential{
		ID:       "123",
		Login:    "user",
		Password: "pass",
		Metadata: "meta",
	}
	key := []byte("12345678901234567890123456789012")
	wrapper := &cryptowrap.CredentialCryptoWrapper{Credential: plain}
	err := wrapper.Encrypt(key)
	require.NoError(t, err)

	// Мок менеджера ключей
	mockKeyMgr := &mockCryptoKeyManager{
		loadKeyData: key,
	}
	// Мок CredentialManager возвращает зашифрованные данные
	mockCredMgr := &mockCredentialManager{
		getByIDResult: plain,
	}

	appSvc := &app.AppServices{
		ConnManager:       &mockConnManager{},
		CryptoKeyManager:  mockKeyMgr,
		CredentialManager: mockCredMgr,
		Logger:            zap.NewNop(),
	}

	cred, err := appSvc.GetCredentialByID(ctx, "123")
	require.NoError(t, err)
	require.Equal(t, "123", cred.ID)
	require.Equal(t, "user", cred.Login) // Проверяем, что расшифровалось
	require.Equal(t, "pass", cred.Password)
	require.Equal(t, "meta", cred.Metadata)

	// Ошибка подключения
	appSvc.ConnManager = &mockConnManager{connectErr: errors.New("connect error")}
	cred, err = appSvc.GetCredentialByID(ctx, "123")
	require.Error(t, err)
	require.Nil(t, cred)
	require.EqualError(t, err, "connect error")
}

func TestGetCredentials(t *testing.T) {
	ctx := context.Background()

	key := []byte("12345678901234567890123456789012")

	// Создаём список с зашифрованными Credential
	plainList := []model.Credential{
		{ID: "1", Login: "login1", Password: "pass1", Metadata: "meta1"},
		{ID: "2", Login: "login2", Password: "pass2", Metadata: "meta2"},
	}

	for i := range plainList {
		wrapper := &cryptowrap.CredentialCryptoWrapper{Credential: &plainList[i]}
		err := wrapper.Encrypt(key)
		require.NoError(t, err)
	}

	mockKeyMgr := &mockCryptoKeyManager{loadKeyData: key}
	mockCredMgr := &mockCredentialManager{getByUserIDResult: plainList}

	appSvc := &app.AppServices{
		ConnManager:       &mockConnManager{},
		CryptoKeyManager:  mockKeyMgr,
		CredentialManager: mockCredMgr,
		Logger:            zap.NewNop(),
	}

	creds, err := appSvc.GetCredentials(ctx)
	require.NoError(t, err)
	require.Len(t, creds, 2)

	// Проверяем, что данные расшифровались
	require.Equal(t, "login1", creds[0].Login)
	require.Equal(t, "pass1", creds[0].Password)
	require.Equal(t, "meta1", creds[0].Metadata)
	require.Equal(t, "login2", creds[1].Login)
	require.Equal(t, "pass2", creds[1].Password)
	require.Equal(t, "meta2", creds[1].Metadata)

	// Ошибка подключения
	appSvc.ConnManager = &mockConnManager{connectErr: errors.New("connect error")}
	creds, err = appSvc.GetCredentials(ctx)
	require.Error(t, err)
	require.Nil(t, creds)
	require.EqualError(t, err, "connect error")
}

func TestUpdateCredential(t *testing.T) {
	ctx := context.Background()
	mockKeyMgr := &mockCryptoKeyManager{
		loadKeyData: []byte("12345678901234567890123456789012"),
	}

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:       &mockConnManager{connectErr: errors.New("connect failed")},
		CryptoKeyManager:  mockKeyMgr,
		CredentialManager: &mockCredentialManager{},
		Logger:            zap.NewNop(),
	}
	err := appSvc.UpdateCredential(ctx, &model.Credential{})
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешное обновление
	credMgr := &mockCredentialManager{}
	appSvc = &app.AppServices{
		ConnManager:       &mockConnManager{},
		CryptoKeyManager:  mockKeyMgr,
		CredentialManager: credMgr,
		Logger:            zap.NewNop(),
	}

	cred := &model.Credential{
		Login:    "user",
		Password: "pass",
		Metadata: "meta",
	}
	err = appSvc.UpdateCredential(ctx, cred)
	require.NoError(t, err)
	require.True(t, credMgr.setClientCalled)
	require.NotEqual(t, "user", cred.Login)
	require.NotEqual(t, "pass", cred.Password)
	require.NotEqual(t, "meta", cred.Metadata)
}

func TestDeleteCredential(t *testing.T) {
	ctx := context.Background()

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:       &mockConnManager{connectErr: errors.New("connect failed")},
		CredentialManager: &mockCredentialManager{},
		Logger:            zap.NewNop(),
	}
	err := appSvc.DeleteCredential(ctx, "id")
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешное удаление
	credMgr := &mockCredentialManager{}
	appSvc = &app.AppServices{
		ConnManager:       &mockConnManager{},
		CredentialManager: credMgr,
		Logger:            zap.NewNop(),
	}
	err = appSvc.DeleteCredential(ctx, "id")
	require.NoError(t, err)
	require.True(t, credMgr.setClientCalled)
}
