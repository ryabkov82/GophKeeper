package app

import (
	"context"
	"io"
	"os"

	"github.com/ryabkov82/gophkeeper/internal/client/crypto"
	"github.com/ryabkov82/gophkeeper/internal/client/cryptowrap"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto"
)

// progressReader оборачивает io.Reader и отправляет прогресс в канал
type progressReader struct {
	reader       io.Reader
	total        int64
	sent         int64
	progressChan chan<- ProgressMsg
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.sent += int64(n)
		select {
		case r.progressChan <- ProgressMsg{Done: r.sent, Total: r.total}:
		default: // если канал заблокирован, пропускаем
		}
	}
	return n, err
}

// progressMsg используется для отправки прогресса в Bubble Tea
type ProgressMsg struct {
	Done  int64
	Total int64
}

// ensureBinaryDataClient гарантирует создание gRPC клиента для BinaryData сервиса
func (s *AppServices) ensureBinaryDataClient(ctx context.Context) error {
	conn, err := s.getGRPCConn(ctx)
	if err != nil {
		return err
	}

	client := proto.NewBinaryDataServiceClient(conn)
	s.BinaryDataManager.SetClient(client)
	return nil
}

// GetBinaryDataInfo получает метаданные бинарных данных по ID (без содержимого)
func (s *AppServices) GetBinaryDataInfo(ctx context.Context, id string) (*model.BinaryData, error) {
	if err := s.ensureBinaryDataClient(ctx); err != nil {
		return nil, err
	}

	data, err := s.BinaryDataManager.GetInfo(ctx, id)
	if err != nil {
		return nil, err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return nil, err
	}

	wrapper := &cryptowrap.BinaryDataCryptoWrapper{BinaryData: data}
	if err := wrapper.Decrypt(key); err != nil {
		return nil, err
	}

	return data, nil
}

// ListBinaryData возвращает список всех бинарных данных пользователя (только метаданные)
func (s *AppServices) ListBinaryData(ctx context.Context) ([]model.BinaryData, error) {
	if err := s.ensureBinaryDataClient(ctx); err != nil {
		return nil, err
	}

	list, err := s.BinaryDataManager.List(ctx)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// sendBinaryData загружает или обновляет бинарные данные на сервер с шифрованием и прогрессом
func (s *AppServices) sendBinaryData(
	ctx context.Context,
	data *model.BinaryData,
	filePath string,
	progressChan chan<- ProgressMsg,
	method func(ctx context.Context, data *model.BinaryData, content io.Reader) error, // upload или update
) error {
	if err := s.ensureBinaryDataClient(ctx); err != nil {
		return err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return err
	}

	// Шифруем Metadata
	wrapper := &cryptowrap.BinaryDataCryptoWrapper{BinaryData: data}
	if err := wrapper.Encrypt(key); err != nil {
		return err
	}

	src, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer src.Close()

	fi, err := src.Stat()
	if err != nil {
		return err
	}
	totalSize := fi.Size()

	progReader := &progressReader{
		reader:       src,
		total:        totalSize,
		progressChan: progressChan,
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		if err := crypto.EncryptStream(progReader, pw, key); err != nil {
			_ = pw.CloseWithError(err)
		}
	}()

	done := make(chan error, 1)
	go func() {
		done <- method(ctx, data, pr)
	}()

	select {
	case <-ctx.Done():
		_ = pr.CloseWithError(ctx.Err())
		return ctx.Err()
	case err := <-done:
		_ = pr.CloseWithError(err)
		return err
	}
}

// UploadBinaryData загружает файл на сервер с потоковым шифрованием
func (s *AppServices) UploadBinaryData(ctx context.Context, data *model.BinaryData, filePath string, progressChan chan<- ProgressMsg) error {
	return s.sendBinaryData(ctx, data, filePath, progressChan, s.BinaryDataManager.Upload)
}

func (s *AppServices) UpdateBinaryData(ctx context.Context, data *model.BinaryData, filePath string, progressChan chan<- ProgressMsg) error {
	return s.sendBinaryData(ctx, data, filePath, progressChan, s.BinaryDataManager.Update)
}

// DownloadBinaryData скачивает файл с сервера, расшифровывает его и отправляет прогресс через канал
func (s *AppServices) DownloadBinaryData(
	ctx context.Context,
	dataID, destPath string,
	progressCh chan<- int64,
) error {
	if err := s.ensureBinaryDataClient(ctx); err != nil {
		return err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return err
	}

	stream, err := s.BinaryDataManager.Download(ctx, dataID)
	if err != nil {
		return err
	}
	defer stream.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	pr, pw := io.Pipe()
	done := make(chan error, 1)

	// Горутина для дешифровки
	go func() {
		defer pw.Close()
		done <- crypto.DecryptStream(stream, pw, key)
	}()

	// Чтение из пайпа и запись в файл с отправкой прогресса
	buf := make([]byte, 32*1024)
	var totalWritten int64
	for {
		select {
		case <-ctx.Done():
			_ = pr.CloseWithError(ctx.Err())
			return ctx.Err()
		default:
		}

		n, err := pr.Read(buf)
		if n > 0 {
			if _, wErr := dst.Write(buf[:n]); wErr != nil {
				_ = pr.CloseWithError(wErr)
				return wErr
			}
			totalWritten += int64(n)
			if progressCh != nil {
				progressCh <- totalWritten
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	// Дожидаемся окончания дешифровки
	if derr := <-done; derr != nil {
		return derr
	}

	return nil
}

// DeleteBinaryData удаляет бинарные данные по ID
func (s *AppServices) DeleteBinaryData(ctx context.Context, userID, id string) error {
	if err := s.ensureBinaryDataClient(ctx); err != nil {
		return err
	}
	return s.BinaryDataManager.Delete(ctx, id)
}
