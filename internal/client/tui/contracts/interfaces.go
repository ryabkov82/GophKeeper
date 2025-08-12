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
