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
	progressChan chan<- int64
}

// Read реализует io.Reader и отправляет информацию о количестве
// прочитанных байт в канал прогресса.
func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.sent += int64(n)
		select {
		case r.progressChan <- r.sent:
		default: // если канал заблокирован, пропускаем
		}
	}
	return n, err
}

// progressWriter считает суммарно записанные байты и шлёт прогресс в канал.
type progressWriter struct {
	w  io.Writer
	n  int64
	ch chan<- int64
}

// Write реализует io.Writer и сообщает о прогрессе записи через канал.
func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	if n > 0 && pw.ch != nil {
		pw.n += int64(n)
		select {
		case pw.ch <- pw.n:
		default:
		}
	}
	return n, err
}

// ctxReader прерывает чтение, если контекст отменён.
type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

// Read реализует io.Reader с учётом отмены контекста.
func (cr *ctxReader) Read(p []byte) (int, error) {
	select {
	case <-cr.ctx.Done():
		return 0, cr.ctx.Err()
	default:
		return cr.r.Read(p)
	}
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
	progressChan chan<- int64,
	method func(ctx context.Context, data *model.BinaryData, content io.Reader) error, // upload или update
) error {
	if err := s.ensureBinaryDataClient(ctx); err != nil {
		return err
	}

	data.ClientPath = filePath
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
func (s *AppServices) UploadBinaryData(ctx context.Context, data *model.BinaryData, filePath string, progressChan chan<- int64) error {
	return s.sendBinaryData(ctx, data, filePath, progressChan, s.BinaryDataManager.Upload)
}

// UpdateBinaryDataInfo обновляет только метаданные бинарных данных без пересылки содержимого
func (s *AppServices) UpdateBinaryDataInfo(ctx context.Context, data *model.BinaryData) error {
	if err := s.ensureBinaryDataClient(ctx); err != nil {
		return err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return err
	}

	wrapper := &cryptowrap.BinaryDataCryptoWrapper{BinaryData: data}
	if err := wrapper.Encrypt(key); err != nil {
		return err
	}

	return s.BinaryDataManager.UpdateInfo(ctx, data)
}

// CreateBinaryDataInfo создаёт запись метаданных без отправки содержимого файла
func (s *AppServices) CreateBinaryDataInfo(ctx context.Context, data *model.BinaryData) error {
	if err := s.ensureBinaryDataClient(ctx); err != nil {
		return err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return err
	}

	wrapper := &cryptowrap.BinaryDataCryptoWrapper{BinaryData: data}
	if err := wrapper.Encrypt(key); err != nil {
		return err
	}

	return s.BinaryDataManager.CreateInfo(ctx, data)
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

	src, err := s.BinaryDataManager.Download(ctx, dataID)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Оборачиваем вход и выход: вход — реагирует на cancel контекста,
	// выход — считает записанные байты и репортит прогресс.
	in := &ctxReader{ctx: ctx, r: src}
	out := &progressWriter{w: dst, ch: progressCh}

	// Потоковая дешифровка напрямую в файл без промежуточного буфера/пайпа.
	if err := crypto.DecryptStream(in, out, key); err != nil {
		return err
	}

	return nil

}

// DeleteBinaryData удаляет бинарные данные по ID
func (s *AppServices) DeleteBinaryData(ctx context.Context, id string) error {
	if err := s.ensureBinaryDataClient(ctx); err != nil {
		return err
	}
	return s.BinaryDataManager.Delete(ctx, id)
}
