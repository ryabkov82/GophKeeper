package service

import (
	"github.com/ryabkov82/gophkeeper/internal/domain/repository"
	"github.com/ryabkov82/gophkeeper/internal/domain/service"
	"github.com/ryabkov82/gophkeeper/internal/domain/storage"
	"github.com/ryabkov82/gophkeeper/internal/pkg/jwtutils"
)

// serviceFactory — конкретная реализация фабрики сервисов.
type serviceFactory struct {
	auth       service.AuthService
	credential service.CredentialService
	bankCard   service.BankCardService
	textData   service.TextDataService
	binaryData service.BinaryDataService
}

// NewServiceFactory создает фабрику сервисов.
// repoFactory — фабрика репозиториев, jwt — менеджер токенов.
func NewServiceFactory(repoFactory repository.StorageFactory, binaryDataStorage storage.BinaryDataStorage, jwt *jwtutils.TokenManager) service.ServiceFactory {
	return &serviceFactory{
		auth:       NewAuthService(repoFactory.User(), jwt),
		credential: NewCredentialService(repoFactory.Credential()),
		bankCard:   NewBankCardService(repoFactory.BankCard()),
		textData:   NewTextDataService(repoFactory.TextData()),
		binaryData: NewBinaryDataService(repoFactory.BinaryData(), binaryDataStorage),
	}
}

func (f *serviceFactory) Auth() service.AuthService {
	return f.auth
}

func (f *serviceFactory) Credential() service.CredentialService {
	return f.credential
}

func (f *serviceFactory) BankCard() service.BankCardService {
	return f.bankCard
}

func (f *serviceFactory) TextData() service.TextDataService {
	return f.textData
}

func (f *serviceFactory) BinaryData() service.BinaryDataService {
	return f.binaryData
}
