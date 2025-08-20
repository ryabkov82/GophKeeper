package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/cryptowrap"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateTextData(t *testing.T) {
	ctx := context.Background()

	mockKeyMgr := &mockCryptoKeyManager{
		loadKeyData: []byte("12345678901234567890123456789012"),
	}

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:      &mockConnManager{connectErr: errors.New("connect failed")},
		CryptoKeyManager: mockKeyMgr,
		TextDataManager:  &mockTextDataManager{},
		Logger:           zap.NewNop(),
	}
	err := appSvc.CreateTextData(ctx, &model.TextData{})
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешный кейс
	textMgr := &mockTextDataManager{}
	appSvc = &app.AppServices{
		ConnManager:      &mockConnManager{},
		CryptoKeyManager: mockKeyMgr,
		TextDataManager:  textMgr,
		Logger:           zap.NewNop(),
	}
	td := &model.TextData{
		Title:    "My Note",
		Content:  []byte("Secret content"),
		Metadata: "meta",
	}
	err = appSvc.CreateTextData(ctx, td)
	require.NoError(t, err)
	require.True(t, textMgr.setClientCalled)

	// Проверяем, что зашифровалось (Title не шифруется!)
	require.Equal(t, "My Note", td.Title)
	require.NotEqual(t, "Secret content", td.Content)
	require.NotEqual(t, "meta", td.Metadata)
}

func TestGetTextDataByID(t *testing.T) {
	ctx := context.Background()
	key := []byte("12345678901234567890123456789012")

	plain := &model.TextData{
		ID:       "t1",
		Title:    "Note",
		Content:  []byte("Secret"),
		Metadata: "meta",
	}
	wrapper := &cryptowrap.TextDataCryptoWrapper{TextData: plain}
	err := wrapper.Encrypt(key)
	require.NoError(t, err)

	mockKeyMgr := &mockCryptoKeyManager{loadKeyData: key}
	mockTextMgr := &mockTextDataManager{getByIDResult: plain}

	appSvc := &app.AppServices{
		ConnManager:      &mockConnManager{},
		CryptoKeyManager: mockKeyMgr,
		TextDataManager:  mockTextMgr,
		Logger:           zap.NewNop(),
	}

	td, err := appSvc.GetTextDataByID(ctx, "t1")
	require.NoError(t, err)
	require.Equal(t, "t1", td.ID)
	require.Equal(t, "Note", td.Title)
	require.Equal(t, "Secret", string(td.Content))
	require.Equal(t, "meta", td.Metadata)

	// Ошибка подключения
	appSvc.ConnManager = &mockConnManager{connectErr: errors.New("connect error")}
	td, err = appSvc.GetTextDataByID(ctx, "t1")
	require.Error(t, err)
	require.Nil(t, td)
	require.EqualError(t, err, "connect error")
}

func TestGetTextDataTitles(t *testing.T) {
	ctx := context.Background()
	key := []byte("12345678901234567890123456789012")

	plainList := []*model.TextData{
		{ID: "1", Title: "t1", Content: []byte("c1"), Metadata: "m1"},
		{ID: "2", Title: "t2", Content: []byte("c2"), Metadata: "m2"},
	}
	for i := range plainList {
		w := &cryptowrap.TextDataCryptoWrapper{TextData: plainList[i]}
		err := w.Encrypt(key)
		require.NoError(t, err)
	}

	mockKeyMgr := &mockCryptoKeyManager{loadKeyData: key}
	mockTextMgr := &mockTextDataManager{getTitlesResult: plainList}

	appSvc := &app.AppServices{
		ConnManager:      &mockConnManager{},
		CryptoKeyManager: mockKeyMgr,
		TextDataManager:  mockTextMgr,
		Logger:           zap.NewNop(),
	}

	list, err := appSvc.GetTextDataTitles(ctx)
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Equal(t, "t1", list[0].Title)
	require.Equal(t, "t2", list[1].Title)

	// Ошибка подключения
	appSvc.ConnManager = &mockConnManager{connectErr: errors.New("connect error")}
	list, err = appSvc.GetTextDataTitles(ctx)
	require.Error(t, err)
	require.Nil(t, list)
	require.EqualError(t, err, "connect error")
}

func TestUpdateTextData(t *testing.T) {
	ctx := context.Background()
	mockKeyMgr := &mockCryptoKeyManager{loadKeyData: []byte("12345678901234567890123456789012")}

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:      &mockConnManager{connectErr: errors.New("connect failed")},
		CryptoKeyManager: mockKeyMgr,
		TextDataManager:  &mockTextDataManager{},
		Logger:           zap.NewNop(),
	}
	err := appSvc.UpdateTextData(ctx, &model.TextData{})
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешное обновление
	textMgr := &mockTextDataManager{}
	appSvc = &app.AppServices{
		ConnManager:      &mockConnManager{},
		CryptoKeyManager: mockKeyMgr,
		TextDataManager:  textMgr,
		Logger:           zap.NewNop(),
	}
	td := &model.TextData{
		Title:    "Note",
		Content:  []byte("Secret"),
		Metadata: "meta",
	}
	err = appSvc.UpdateTextData(ctx, td)
	require.NoError(t, err)
	require.True(t, textMgr.setClientCalled)

	require.Equal(t, "Note", td.Title)
	require.NotEqual(t, "Secret", td.Content)
	require.NotEqual(t, "meta", td.Metadata)
}

func TestDeleteTextData(t *testing.T) {
	ctx := context.Background()

	// Ошибка подключения
	appSvc := &app.AppServices{
		ConnManager:     &mockConnManager{connectErr: errors.New("connect failed")},
		TextDataManager: &mockTextDataManager{},
		Logger:          zap.NewNop(),
	}
	err := appSvc.DeleteTextData(ctx, "id")
	require.Error(t, err)
	require.EqualError(t, err, "connect failed")

	// Успешное удаление
	textMgr := &mockTextDataManager{}
	appSvc = &app.AppServices{
		ConnManager:     &mockConnManager{},
		TextDataManager: textMgr,
		Logger:          zap.NewNop(),
	}
	err = appSvc.DeleteTextData(ctx, "id")
	require.NoError(t, err)
	require.True(t, textMgr.setClientCalled)
}
