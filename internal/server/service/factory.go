package service

import (
	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtutils"
)

// NewServices создает и возвращает набор сервисов с реализациями интерфейсов domain/service.
func NewServices(repo *repository.Repositories, jwt *jwtutils.TokenManager) *service.Services {
	return &service.Services{
		Auth: NewAuthService(repo.User, jwt),
		// Data: NewDataService(repo.Data),
	}
}
