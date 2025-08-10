package service

import (
	"context"
)

// AuthService описывает контракт сервисов аутентификации и регистрации
type AuthService interface {
	Register(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) (accessToken string, err error)
}
