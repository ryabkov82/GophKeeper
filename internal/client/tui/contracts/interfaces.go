package contracts

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// AuthService определяет интерфейс для аутентификации пользователя.
// Он используется для выполнения операций входа и регистрации в пользовательском интерфейсе (TUI).
type AuthService interface {
	// LoginUser выполняет вход пользователя с указанным логином и паролем.
	// Возвращает ошибку, если вход не удался.
	LoginUser(ctx context.Context, login, password string) error

	// RegisterUser регистрирует нового пользователя с заданным логином и паролем.
	// Возвращает ошибку, если регистрация не удалась.
	RegisterUser(ctx context.Context, login, password string) error
}

// CredentialService описывает интерфейс управления учётными данными (логины/пароли).
type CredentialService interface {
	CreateCredential(ctx context.Context, cred *model.Credential) error
	GetCredentialByID(ctx context.Context, id string) (*model.Credential, error)
	GetCredentials(ctx context.Context) ([]model.Credential, error)
	UpdateCredential(ctx context.Context, cred *model.Credential) error
	DeleteCredential(ctx context.Context, id string) error
}

// BankCardService описывает интерфейс управления данными банковских карт.
type BankCardService interface {
	// CreateBankCard создает новую запись банковской карты.
	// Перед сохранением все чувствительные данные должны быть зашифрованы.
	CreateBankCard(ctx context.Context, card *model.BankCard) error

	// GetBankCardByID возвращает банковскую карту по её идентификатору.
	// При необходимости выполняет дешифровку данных карты.
	GetBankCardByID(ctx context.Context, id string) (*model.BankCard, error)

	// GetBankCards возвращает все банковские карты пользователя.
	// При необходимости выполняет дешифровку данных карт.
	GetBankCards(ctx context.Context) ([]model.BankCard, error)

	// UpdateBankCard обновляет существующую запись банковской карты.
	// Перед сохранением все чувствительные данные должны быть зашифрованы.
	UpdateBankCard(ctx context.Context, card *model.BankCard) error

	// DeleteBankCard удаляет запись банковской карты по идентификатору.
	DeleteBankCard(ctx context.Context, id string) error
}

// TextDataService описывает интерфейс управления текстовыми данными.
type TextDataService interface {
	// CreateTextData создаёт новый текстовый объект на сервере с шифрованием содержимого
	CreateTextData(ctx context.Context, text *model.TextData) error

	// GetTextDataByID получает текстовые данные по ID с расшифровкой содержимого
	GetTextDataByID(ctx context.Context, id string) (*model.TextData, error)

	// GetTextDataTitles получает только заголовки текстовых данных (без расшифровки контента)
	GetTextDataTitles(ctx context.Context) ([]*model.TextData, error)

	// UpdateTextData обновляет текстовые данные с шифрованием содержимого
	UpdateTextData(ctx context.Context, text *model.TextData) error

	// DeleteTextData удаляет текстовые данные по ID
	DeleteTextData(ctx context.Context, id string) error
}

type BinaryDataService interface {
	// UploadBinaryData загружает бинарный объект на сервер с шифрованием содержимого и прогрессом
	UploadBinaryData(ctx context.Context, data *model.BinaryData, filePath string, progressChan chan<- int64) error

	// DownloadBinaryData скачивает бинарный объект с сервера, расшифровывает и отправляет прогресс через канал
	DownloadBinaryData(ctx context.Context, dataID, destPath string, progressCh chan<- int64) error

	// GetBinaryDataInfo получает метаданные бинарного объекта по ID (без содержимого)
	GetBinaryDataInfo(ctx context.Context, id string) (*model.BinaryData, error)

	// ListBinaryData возвращает список всех бинарных объектов пользователя (только метаданные)
	ListBinaryData(ctx context.Context) ([]model.BinaryData, error)

	// DeleteBinaryData удаляет бинарный объект по ID
	DeleteBinaryData(ctx context.Context, id string) error

	// CreateBinaryDataInfo сохраняет метаданные бинарного объекта
	CreateBinaryDataInfo(ctx context.Context, data *model.BinaryData) error

	// CreateBinaryDataInfo обновляет метаданные бинарного объекта
	UpdateBinaryDataInfo(ctx context.Context, data *model.BinaryData) error
}
