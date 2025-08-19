package app_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/crypto"
	"github.com/ryabkov82/gophkeeper/internal/client/cryptowrap"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Тест UploadBinaryData и UpdateBinaryData через sendBinaryData
func TestUploadUpdateBinaryData(t *testing.T) {
	key := []byte("1234567890123456")

	// Создаём временный файл
	tmpFile, _ := os.CreateTemp("", "testfile")
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("file content")
	tmpFile.Close()

	progressCh := make(chan app.ProgressMsg, 10)

	mockMgr := &mockBinaryDataManager{}
	mockCrypto := &mockCryptoKeyManager{
		loadKeyData: key,
	}

	svc := &app.AppServices{
		ConnManager:       &mockConnManager{},
		BinaryDataManager: mockMgr,
		CryptoKeyManager:  mockCrypto,
		Logger:            zap.NewNop(),
	}

	data := &model.BinaryData{Metadata: "meta"}

	// Upload
	err := svc.UploadBinaryData(context.Background(), data, tmpFile.Name(), progressCh)
	assert.NoError(t, err)
	assert.True(t, mockMgr.setClientCalled) // проверка установки клиента
	assert.True(t, mockCrypto.loadCalled)   // проверка загрузки ключа

	// Update
	err = svc.UpdateBinaryData(context.Background(), data, tmpFile.Name(), progressCh)
	assert.NoError(t, err)
}

// Тест DownloadBinaryData с проверкой прогресса
func TestDownloadBinaryData(t *testing.T) {
	key := []byte("1234567890123456")
	plainContent := []byte("download content")

	// Зашифровываем данные в буфер
	encryptedBuf := new(bytes.Buffer)
	err := crypto.EncryptStream(bytes.NewReader(plainContent), encryptedBuf, key)
	assert.NoError(t, err)

	tmpDest := t.TempDir() + "/out.txt"
	progressCh := make(chan int64, 10)

	mockMgr := &mockBinaryDataManager{
		downloadFn: func(ctx context.Context, id string) (io.ReadCloser, error) {
			// Возвращаем зашифрованный поток
			return io.NopCloser(bytes.NewReader(encryptedBuf.Bytes())), nil
		},
	}
	mockCrypto := &mockCryptoKeyManager{
		loadKeyData: key,
	}

	svc := &app.AppServices{
		ConnManager:       &mockConnManager{},
		BinaryDataManager: mockMgr,
		CryptoKeyManager:  mockCrypto,
		Logger:            zap.NewNop(),
	}

	err = svc.DownloadBinaryData(context.Background(), "id1", tmpDest, progressCh)
	assert.NoError(t, err)

	got, _ := os.ReadFile(tmpDest)
	assert.Equal(t, plainContent, got) // проверяем расшифрованный результат
}

// Тест ListBinaryData
func TestListBinaryData(t *testing.T) {
	mockMgr := &mockBinaryDataManager{
		listResult: []model.BinaryData{
			{Metadata: "a"},
			{Metadata: "b"},
		},
	}

	svc := &app.AppServices{
		ConnManager:       &mockConnManager{},
		BinaryDataManager: mockMgr,
		Logger:            zap.NewNop(),
	}

	list, err := svc.ListBinaryData(context.Background())
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

// Тест GetBinaryDataInfo
func TestGetBinaryDataInfo(t *testing.T) {
	key := []byte("1234567890123456")

	// Исходные метаданные
	plainMeta := "meta"

	// Создаём объект BinaryData и шифруем Metadata
	dataToReturn := &model.BinaryData{Metadata: plainMeta}
	wrapper := &cryptowrap.BinaryDataCryptoWrapper{BinaryData: dataToReturn}
	err := wrapper.Encrypt(key)
	assert.NoError(t, err)

	mockMgr := &mockBinaryDataManager{
		getInfoFn: func(ctx context.Context, id string) (*model.BinaryData, error) {
			// Возвращаем зашифрованный объект
			return dataToReturn, nil
		},
	}
	mockCrypto := &mockCryptoKeyManager{
		loadKeyData: key,
	}

	svc := &app.AppServices{
		ConnManager:       &mockConnManager{},
		BinaryDataManager: mockMgr,
		CryptoKeyManager:  mockCrypto,
		Logger:            zap.NewNop(),
	}

	data, err := svc.GetBinaryDataInfo(context.Background(), "id1")
	assert.NoError(t, err)
	assert.Equal(t, plainMeta, data.Metadata) // проверяем расшифрованное значение
}

// Тест DeleteBinaryData
func TestDeleteBinaryData(t *testing.T) {
	mockMgr := &mockBinaryDataManager{}
	svc := &app.AppServices{
		ConnManager:       &mockConnManager{},
		BinaryDataManager: mockMgr,
		Logger:            zap.NewNop(),
	}

	err := svc.DeleteBinaryData(context.Background(), "user1", "id1")
	assert.NoError(t, err)
}

// Тест UpdateBinaryDataInfo
func TestUpdateBinaryDataInfo(t *testing.T) {
	key := []byte("1234567890123456")

	mockMgr := &mockBinaryDataManager{}
	mockCrypto := &mockCryptoKeyManager{loadKeyData: key}

	svc := &app.AppServices{
		ConnManager:       &mockConnManager{},
		BinaryDataManager: mockMgr,
		CryptoKeyManager:  mockCrypto,
		Logger:            zap.NewNop(),
	}

	data := &model.BinaryData{ID: "id1", Title: "title", Metadata: "meta"}

	err := svc.UpdateBinaryDataInfo(context.Background(), data)
	assert.NoError(t, err)
	assert.True(t, mockMgr.setClientCalled)
	assert.True(t, mockCrypto.loadCalled)

	// проверяем, что данные шифровались
	assert.NotEqual(t, "meta", data.Metadata)
	wrapper := &cryptowrap.BinaryDataCryptoWrapper{BinaryData: data}
	err = wrapper.Decrypt(key)
	assert.NoError(t, err)
	assert.Equal(t, "meta", data.Metadata)
}

// Тест CreateBinaryDataInfo
func TestCreateBinaryDataInfo(t *testing.T) {
	key := []byte("1234567890123456")

	mockMgr := &mockBinaryDataManager{}
	mockCrypto := &mockCryptoKeyManager{loadKeyData: key}

	svc := &app.AppServices{
		ConnManager:       &mockConnManager{},
		BinaryDataManager: mockMgr,
		CryptoKeyManager:  mockCrypto,
		Logger:            zap.NewNop(),
	}

	data := &model.BinaryData{Title: "title", Metadata: "meta"}

	err := svc.CreateBinaryDataInfo(context.Background(), data)
	assert.NoError(t, err)
	assert.True(t, mockMgr.setClientCalled)
	assert.True(t, mockCrypto.loadCalled)

	// проверяем, что данные шифровались
	assert.NotEqual(t, "meta", data.Metadata)
	wrapper := &cryptowrap.BinaryDataCryptoWrapper{BinaryData: data}
	err = wrapper.Decrypt(key)
	assert.NoError(t, err)
	assert.Equal(t, "meta", data.Metadata)
}
