package repository

import (
	"context"

	"github.com/ryabkov82/gophkeeper/internal/domain/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, login, hash, salt string) error
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
}
