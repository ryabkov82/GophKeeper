// Package app инкапсулирует «сервисы» клиентского приложения GophKeeper и
// предоставляет высокоуровневый API для аутентификации и работы с данными
// пользователя (учётки, банковские карты, заметки, бинарные файлы).
//
// Ключевая сущность — контейнер зависимостей AppServices. Он создаётся через
// NewAppServices(cfg) и содержит готовые менеджеры, логгер и менеджер gRPC‑
// соединения. Такой подход упрощает передачу зависимостей в TUI и тестирование.
//
// Состав AppServices:
//
//   - AuthManager       — регистрация/логин пользователя, хранение токена.
//   - CredentialManager — CRUD для учётных данных (логин/пароль/метаданные).
//   - BankCardManager   — CRUD для банковских карт (имя, номер, срок, CVV, мета).
//   - TextDataManager   — CRUD для текстовых заметок.
//   - BinaryDataManager — загрузка/обновление/скачивание/удаление бинарных файлов,
//     а также выдача списка и метаданных.
//   - CryptoKeyManager  — генерация/сохранение/загрузка симметричного ключа,
//     используемого для шифрования пользовательских данных.
//   - ConnManager       — управление gRPC‑подключением к серверу (TLS/без TLS).
//   - Logger            — zap‑логгер (уровень и каталог задаются конфигурацией).
//
// Инициализация
//
//	NewAppServices(cfg) поднимает логгер, настраивает ConnManager, подготавливает
//	файловые стораджи для токена и ключа шифрования и создаёт менеджеры доменов.
//
//	getGRPCConn(ctx) скрывает логику установления/восстановления соединения
//	(через ConnManager) и используется всеми ensure*Client.
//
// Жизненный цикл
//
//	Close() корректно закрывает ресурсы (gRPC) с защитой через sync.Once и
//	ожиданием завершения до 5s. Рекомендуется всегда вызывать Close().
//
// Аутентификация
//
//	ensureAuthClient(ctx) — создаёт gRPC‑клиента Auth и пробрасывает его в
//	AuthManager.
//
//	LoginUser(ctx, login, password) — логинится, получает от сервера соль,
//	генерирует и сохраняет симметричный ключ шифрования через CryptoKeyManager.
//
//	RegisterUser(ctx, login, password) — регистрирует пользователя и затем
//	вызывает LoginUser.
//
// Данные пользователя и шифрование
//
//	Для всех сущностей перед отправкой на сервер конфиденциальные поля шифруются,
//	а после получения — расшифровываются. Шифрование инкапсулировано в
//	обёртках из internal/client/cryptowrap.
//
//	Credentials:
//	  - CreateCredential / GetCredentialByID / GetCredentials / UpdateCredential / DeleteCredential
//	  - Шифруются: Login, Password, Metadata.
//
//	Bank cards:
//	  - CreateBankCard / GetBankCardByID / GetBankCards / UpdateBankCard / DeleteBankCard
//	  - Шифруются: CardholderName, CardNumber, ExpiryDate, CVV, Metadata.
//
//	Text notes:
//	  - CreateTextData / GetTextDataByID / GetTextDataTitles / UpdateTextData / DeleteTextData
//	  - Шифруются: Content, Metadata. Заголовок (Title) хранится в открытом виде,
//	    что позволяет отдавать списки заголовков без расшифровки.
//
//	Binary files:
//	  - UploadBinaryData / UpdateBinaryData — потоковое шифрование содержимого файла
//	    при отправке (EncryptStream). Метаданные шифруются перед RPC.
//	  - DownloadBinaryData — потоковая расшифровка (DecryptStream) и запись в файл.
//	  - GetBinaryDataInfo — получение и расшифровка только метаданных.
//	  - ListBinaryData / DeleteBinaryData — работа со списком и удалением.
//	  - Для отображения прогресса используются каналы:
//	      * при upload/update — chan ProgressMsg { Done, Total },
//	      * при download — chan int64 (накопленный байт‑каунтер).
//
// Клиенты gRPC
//
//	Каждый доменный метод начинается с ensure*Client(ctx), который запрашивает
//	активное соединение (getGRPCConn) и создаёт конкретный gRPC‑клиент
//	(из internal/pkg/proto), после чего передаёт его соответствующему Manager’у.
//	Это отделяет сетевую часть от бизнес‑логики и упрощает мокирование в тестах.
//
// Запуск TUI
//
//	RunWithServices(cfg, runTUI) — удобная обёртка для старта приложения:
//	  1) создаёт AppServices;
//	  2) формирует контекст, отменяемый по SIGINT/SIGTERM;
//	  3) вызывает предоставленную функцию runTUI(ctx, services, progFactory);
//	  4) гарантирует закрытие ресурсов (services.Close, logger.Close).
//
// Обработка ошибок и логирование
//
//	Ошибки подключения и криптоопераций прозрачно возвращаются вызывающему коду.
//	Важные операции логируются через zap с учётом уровня из конфигурации.
//
// Пример (упрощённо):
//
//	cfg := &config.ClientConfig{ ServerAddress: "...", LogLevel: "info", ... }
//	if err := app.RunWithServices(cfg, func(ctx context.Context, s *app.AppServices, _ tuiiface.ProgramFactory) error {
//	    if err := s.LoginUser(ctx, "alice", "secret"); err != nil {
//	        return err
//	    }
//	    cred := &model.Credential{ Login: "site_user", Password: "p@ss", Metadata: "example" }
//	    if err := s.CreateCredential(ctx, cred); err != nil {
//	        return err
//	    }
//	    list, err := s.GetCredentials(ctx)
//	    if err != nil {
//	        return err
//	    }
//	    // использовать list...
//	    return nil
//	}); err != nil {
//	    log.Fatal(err)
//	}
//
// Пакет ориентирован на использование внутри TUI, но может применяться в любом
// клиентском слое, где требуется безопасная работа с данными через gRPC.
//
// ВНИМАНИЕ: для корректной работы криптографии после логина должен быть
// сгенерирован и сохранён симметричный ключ (LoginUser/RegisterUser), иначе
// методы шифрования/дешифрования вернут ошибку.
package app
