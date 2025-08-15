// Package connection реализует потокобезопасный менеджер подключения к gRPC-серверу
// с поддержкой TLS, автоматического переподключения и добавления токена авторизации
// в исходящие запросы.
//
// Пакет позволяет:
//   - Устанавливать и переиспользовать gRPC-соединение с учётом таймаута.
//   - Работать как в защищённом (TLS), так и в незашифрованном режиме.
//   - Добавлять токен авторизации в gRPC-запросы (кроме явно исключённых методов).
//   - Потокобезопасно получать текущее соединение и проверять его состояние.
//
// Использование:
//
//	cfg := &connection.Config{
//	    ServerAddress:  "localhost:50051",
//	    UseTLS:         true,
//	    CACertPath:     "ca.pem",
//	    ConnectTimeout: 5 * time.Second,
//	}
//
//	manager := connection.New(cfg, logger, authManager)
//
//	conn, err := manager.Connect(context.Background())
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer manager.Close()
package connection
