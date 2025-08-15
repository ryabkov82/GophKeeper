package service

// ServiceFactory определяет интерфейс фабрики сервисов.
// Это позволяет собирать набор сервисов из разных реализаций репозиториев.
type ServiceFactory interface {
	Auth() AuthService
	Credential() CredentialService
	BankCard() BankCardService
}
