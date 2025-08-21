# GophKeeper

GophKeeper — клиент-серверное хранилище секретов с текстовым интерфейсом, написанное на Go.

## Возможности

- хранение учётных данных, банковских карт, текстовых заметок и бинарных файлов;
- шифрование данных на стороне клиента с помощью ключей Argon2id и AES‑GCM;
- взаимодействие клиента и сервера по gRPC;
- настраиваемые файлы конфигурации и переменные окружения;
- TUI-клиент на базе библиотеки Bubble Tea.

## Структура репозитория

```text
.
├── cmd/                # точки входа
│   ├── client/         # TUI‑клиент
│   └── server/         # gRPC‑сервер
├── internal/
│   ├── client/         # код клиентского приложения
│   │   ├── app/        # инициализация и управление жизненным циклом
│   │   ├── config/     # загрузка конфигурации
│   │   ├── connection/ # установка gRPC‑соединений
│   │   ├── crypto/     # работа с ключами и шифрованием
│   │   ├── forms/      # формы интерфейса
│   │   ├── service/    # взаимодействие с сервером
│   │   ├── storage/    # локальное хранилище
│   │   └── tui/        # компоненты Bubble Tea
│   ├── server/         # код сервера
│   │   ├── config/     # параметры запуска
│   │   ├── grpc/       # gRPC‑обработчики
│   │   ├── service/    # бизнес‑логика
│   │   └── storage/    # работа с БД
│   ├── domain/         # доменные модели и интерфейсы
│   │   ├── model/
│   │   ├── repository/
│   │   ├── service/
│   │   └── storage/
│   ├── pkg/            # общие пакеты
│   │   ├── crypto/
│   │   ├── jwtauth/
│   │   ├── jwtutils/
│   │   ├── logger/
│   │   └── proto/      # protobuf‑схемы gRPC API
│   └── migrations/     # SQL‑миграции БД
├── certs/              # тестовые TLS‑сертификаты
├── Makefile            # команды сборки и запуска
├── go.mod, go.sum      # зависимости проекта
└── README.md           # документация
```

## Шифрование на стороне клиента

При первом входе пользователь задаёт мастер-пароль, а сервер возвращает
уникальную соль. На их основе локально вычисляется симметричный ключ с помощью
функции Argon2id. Используются параметры по умолчанию
`Time=1`, `Memory=64MiB`, `Threads=4`, `KeyLen=32`, что даёт 256‑битный ключ.
Параметры Argon2 и сам ключ сохраняются в файл `key_file_path` и не передаются
на сервер.

Полученный ключ применяется для шифрования и дешифрования всех записей с
использованием алгоритма AES‑GCM, обеспечивающего конфиденциальность и
целостность. На сервер отправляются только зашифрованные данные.

## Сборка и запуск

```bash
# запуск сервера
go run ./cmd/server

# запуск клиента
go run ./cmd/client
```

Также доступны команды Makefile:

```bash
make run-server
make run-client
```

## Конфигурация

Клиент и сервер поддерживают загрузку настроек из JSON-файлов, флагов командной строки и
переменных окружения. Приоритет источников: значения по умолчанию → JSON → флаги → env.

### Клиент

Параметры клиента:

- `server_address` (`SERVER_ADDRESS`) — адрес gRPC сервера `host:port`;
- `use_tls` (`USE_TLS`) — использовать ли TLS при подключении;
- `tls_skip_verify` (`TLS_SKIP_VERIFY`) — отключить проверку сертификата сервера;
- `ca_cert_path` (`CA_CERT_PATH`) — путь к CA‑сертификату;
- `timeout` (`TIMEOUT`) — таймаут установления соединения;
- `log_level` (`LOG_LEVEL`) — уровень логирования (`debug`, `info`, `warn`, `error`);
- `key_file_path` (`KEY_FILE_PATH`) — путь к файлу с ключом шифрования;
- `token_file_path` (`TOKEN_FILE_PATH`) — путь к файлу токена авторизации;
- `log_dir_path` (`LOG_DIR_PATH`) — директория для логов клиента.

Пример `client_config.json`:

```json
{
  "server_address": "localhost:50051",
  "use_tls": false,
  "timeout": "10s",
  "log_level": "debug",
  "key_file_path": "/home/user/.config/gophkeeper/key",
  "token_file_path": "/home/user/.config/gophkeeper/token",
  "log_dir_path": "/home/user/.local/share/gophkeeper/logs"
}
```

Запуск клиента с конфигом:

```bash
go run ./cmd/client --config client_config.json
# или через переменные окружения
SERVER_ADDRESS=localhost:50051 go run ./cmd/client
```

### Сервер

Параметры сервера:

- `grpc_server_address` (`GRPC_SERVER_ADDRESS`, флаг `-grpc`) — адрес прослушивания gRPC;
- `database_dsn` (`DATABASE_KEEPER`, флаг `-db`) — строка подключения к PostgreSQL;
- `jwt_secret` (`JWT_SECRET`) — секрет для подписи JWT (не короче 32 символов);
- `enable_tls` (`SSL_ENABLE`, флаг `-s`) — включить TLS;
- `ssl_cert_file` (`SSL_CERT_FILE`) — путь к TLS‑сертификату сервера;
- `ssl_key_file` (`SSL_KEY_FILE`) — путь к приватному ключу TLS;
- `log_level` (`LOG_LEVEL`, флаг `-l`) — уровень логирования;
- `binary_data_store_path` (`BINARY_DATA_PATH`, флаг `-binary-path`) — директория для бинарных файлов.

Пример `server_config.json`:

```json
{
  "grpc_server_address": "localhost:50051",
  "database_dsn": "postgres://gophkeeper:gophkeeper@localhost:5432/gophkeeper?sslmode=disable",
  "jwt_secret": "change_me_to_strong_secret",
  "enable_tls": false,
  "log_level": "info",
  "binary_data_store_path": "/var/lib/gophkeeper/binary"
}
```

Запуск сервера с конфигом:

```bash
go run ./cmd/server --config server_config.json
# или через переменные окружения
GRPC_SERVER_ADDRESS=localhost:50051 DATABASE_KEEPER=postgres://... JWT_SECRET=... go run ./cmd/server
```

## Тесты

Запуск всех модульных тестов:

```bash
make test
```
