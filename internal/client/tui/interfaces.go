package tui

import "context"

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
