// Package textdata предоставляет функционал для управления текстовыми данными (заметками, заметками с метаданными)
// в клиентском приложении GophKeeper.
//
// TextDataManager реализует CRUD операции с текстовыми данными,
// взаимодействует с сервером по gRPC, а также логирует все операции.
//
// Основные возможности:
//   - Создание, получение, обновление и удаление текстовых данных пользователя.
//   - Получение списка заголовков всех текстовых данных пользователя без передачи полного содержимого.
//   - Проверка ограничений на размер поля Content перед отправкой на сервер.
//   - Поддержка внедрения моков для тестирования.
//
// Типы:
//   - TextDataManagerIface — интерфейс для управления текстовыми данными,
//     упрощающий мокирование и расширяемость.
//   - TextDataManager — конкретная реализация интерфейса с бизнес-логикой.
//
// Константы:
//   - MaxContentSize — максимальный размер поля Content в байтах (10 МБ), проверяется перед отправкой данных на сервер.
//
// Методы:
//   - CreateTextData — создаёт новую текстовую запись на сервере.
//   - GetTextDataByID — получает полную запись текста по ID.
//   - GetTextDataTitles — получает список заголовков текстовых записей пользователя.
//   - UpdateTextData — обновляет существующую текстовую запись.
//   - DeleteTextData — удаляет текстовую запись по ID.
//   - SetClient — устанавливает gRPC-клиент (позволяет внедрять моки в тестах).
//
// Пример использования:
//
//	logger := zap.NewExample()
//	tdManager := textdata.NewTextDataManager(logger)
//	tdManager.SetClient(grpcClient) // grpcClient реализует pb.TextDataServiceClient
//
//	ctx := context.Background()
//	text := &model.TextData{
//	    ID:      "note-123",
//	    Title:   "Моя заметка",
//	    Content: "Текст заметки",
//	}
//	err := tdManager.CreateTextData(ctx, text)
//
//	titles, err := tdManager.GetTextDataTitles(ctx)
//	note, err := tdManager.GetTextDataByID(ctx, "note-123")
package textdata
