package repository

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

// UserRepository определяет контракт доступа к данным пользователей.
//
// Интерфейс инкапсулирует операции над сущностью пользователя (model.User),
// позволяя реализовать различные способы хранения (например, PostgreSQL, in-memory и т.д.).
//
// Используется в слое бизнес-логики (AuthService) для абстракции от конкретной СУБД.
type UserRepository interface {
	// CreateUser сохраняет нового пользователя с указанным логином, хешом пароля и солью.
	//
	// Возвращает ошибку, если операция завершилась неудачей (например, логин уже существует).
	CreateUser(ctx context.Context, login, hash, salt string) error

	// GetUserByLogin возвращает пользователя по логину.
	//
	// Если пользователь не найден, возвращает (nil, nil).
	// В случае других ошибок возвращается (nil, error).
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
}
