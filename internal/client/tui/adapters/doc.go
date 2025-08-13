// Package adapters содержит адаптеры для преобразования специализированных сервисов
// доменной модели в унифицированный интерфейс DataService, используемый в TUI.
//
// В частности, CredentialAdapter позволяет работать с учётными данными
// (*model.Credential) через интерфейс DataService, предоставляя методы для
// просмотра списка, получения, создания, обновления и удаления записей.
//
// Пример использования:
//
//	credAdapter := adapters.NewCredentialAdapter(credService)
//	items, err := credAdapter.List(ctx)
//	if err != nil {
//	    // обработка ошибки
//	}
package adapters
