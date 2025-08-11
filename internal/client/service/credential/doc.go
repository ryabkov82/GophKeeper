// Package credential предоставляет функционал для управления учётными данными (логинами, паролями)
// в клиентском приложении GophKeeper.
//
// CredentialManager реализует CRUD операции с учётными данными,
// взаимодействует с сервером по gRPC, добавляя в контекст токен авторизации,
// а также логирует все операции.
//
// Основные возможности:
//   - Создание, получение, обновление и удаление учётных данных пользователя.
//   - Работа с gRPC соединениями через abstraction ConnManager.
//   - Автоматическое добавление токена доступа в метаданные контекста запросов.
//   - Поддержка внедрения моков для тестирования.
//
// Типы:
//   - CredentialManagerIface — интерфейс для управления учётными данными,
//     упрощающий мокирование и расширяемость.
//   - CredentialManager — конкретная реализация интерфейса с бизнес-логикой.
//
// Пример использования:
//
//	connManager := connection.New(...)
//	authManager := auth.NewAuthManager(...)
//	logger := zap.NewExample()
//	credManager := credential.NewCredentialManager(connManager, authManager, logger)
//
//	ctx := context.Background()
//	creds, err := credManager.GetCredentialsByUserID(ctx, "user-123")
package credential
