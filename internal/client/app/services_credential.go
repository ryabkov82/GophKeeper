package app

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/client/app/cryptowrap"
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	"github.com/ryabkov82/gophkeeper/internal/pkg/proto"
)

// ensureCredentialClient гарантирует создание gRPC клиента для Credential сервиса и установку его в CredentialManager.
//
// ctx — контекст запроса.
//
// Возвращает ошибку при сбое подключения.
func (s *AppServices) ensureCredentialClient(ctx context.Context) error {
	conn, err := s.getGRPCConn(ctx)
	if err != nil {
		return err
	}

	client := proto.NewCredentialServiceClient(conn)
	s.CredentialManager.SetClient(client)
	return nil
}

// CreateCredential создаёт новую учётную запись (credential) на сервере.
// Перед отправкой данные логина, пароля и метаданных шифруются с помощью
// симметричного ключа, загружаемого из CryptoKeyManager.
//
// ctx — контекст запроса.
// cred — данные учётных данных для создания.
//
// Возвращает ошибку при сбое RPC вызова или шифрования.
func (s *AppServices) CreateCredential(ctx context.Context, cred *model.Credential) error {
	err := s.ensureCredentialClient(ctx)
	if err != nil {
		return err
	}

	key, err := s.CryptoKeyManager.LoadKey() // загрузка ключа шифрования
	if err != nil {
		return err
	}

	wrapper := &cryptowrap.CredentialCryptoWrapper{Credential: cred}
	if err := wrapper.Encrypt(key); err != nil {
		return err
	}

	return s.CredentialManager.CreateCredential(ctx, cred)
}

// GetCredentialByID получает учётные данные по их уникальному идентификатору.
// После получения с сервера происходит расшифровка полей логина, пароля и метаданных
// с помощью симметричного ключа из CryptoKeyManager.
//
// ctx — контекст запроса.
// id — идентификатор учётных данных.
//
// Возвращает найденные учётные данные или ошибку при RPC вызове или дешифровании.
func (s *AppServices) GetCredentialByID(ctx context.Context, id string) (*model.Credential, error) {
	err := s.ensureCredentialClient(ctx)
	if err != nil {
		return nil, err
	}
	cred, err := s.CredentialManager.GetCredentialByID(ctx, id)
	if err != nil {
		return nil, err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return nil, err
	}

	wrapper := &cryptowrap.CredentialCryptoWrapper{Credential: cred}
	if err := wrapper.Decrypt(key); err != nil {
		return nil, err
	}

	return cred, nil
}

// GetCredentials возвращает список учётных данных для заданного пользователя (из контекста).
// Все записи, полученные с сервера, расшифровываются по отдельности.
//
// ctx — контекст запроса.
//
// Возвращает срез учётных данных или ошибку при RPC вызове или дешифровании.
func (s *AppServices) GetCredentials(ctx context.Context) ([]model.Credential, error) {
	err := s.ensureCredentialClient(ctx)
	if err != nil {
		return nil, err
	}
	creds, err := s.CredentialManager.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}

	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return nil, err
	}

	for i := range creds {
		wrapper := &cryptowrap.CredentialCryptoWrapper{Credential: &creds[i]}
		if err := wrapper.Decrypt(key); err != nil {
			return nil, err
		}
	}

	return creds, nil
}

// UpdateCredential обновляет существующую учётную запись.
// Перед отправкой обновлённые данные логина, пароля и метаданных шифруются
// с помощью ключа из CryptoKeyManager.
//
// ctx — контекст запроса.
// cred — обновлённые данные учётных данных.
//
// Возвращает ошибку при сбое RPC вызова или шифрования.
func (s *AppServices) UpdateCredential(ctx context.Context, cred *model.Credential) error {
	err := s.ensureCredentialClient(ctx)
	if err != nil {
		return err
	}
	key, err := s.CryptoKeyManager.LoadKey()
	if err != nil {
		return err
	}

	wrapper := &cryptowrap.CredentialCryptoWrapper{Credential: cred}
	if err := wrapper.Encrypt(key); err != nil {
		return err
	}

	return s.CredentialManager.UpdateCredential(ctx, cred)
}

// DeleteCredential удаляет учётные данные по идентификатору.
//
// ctx — контекст запроса.
// id — идентификатор учётных данных для удаления.
//
// Возвращает ошибку при сбое RPC вызова.
func (s *AppServices) DeleteCredential(ctx context.Context, id string) error {
	err := s.ensureCredentialClient(ctx)
	if err != nil {
		return err
	}
	return s.CredentialManager.DeleteCredential(ctx, id)
}
