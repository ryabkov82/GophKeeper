// Package storage предоставляет фабрики и реализации интерфейсов репозиториев
// и хранилищ бинарных данных для различных типов хранилищ данных в GophKeeper.
//
// Основная задача пакета — объединить конкретные реализации репозиториев
// (PostgreSQL, локальная память и т.п.) и бинарных хранилищ (локальная ФС, S3)
// за едиными интерфейсами StorageFactory и BinaryDataStorageFactory, чтобы
// верхние слои приложения могли работать с данными независимо от конкретного хранилища.
//
// Пакет содержит:
//
//   - postgresFactory — фабрика репозиториев для PostgreSQL.
//   - NewPostgresFactory — создаёт экземпляр postgresFactory с подключением к БД.
//   - NewStorageFactory — универсальная функция для создания фабрики репозиториев
//     для заданного драйвера.
//   - binaryDataFactory — фабрика для работы с BinaryDataStorage (локальная ФС, S3 и т.п.).
//   - NewBinaryDataFactory — создаёт конкретную реализацию BinaryDataStorageFactory
//     в зависимости от конфигурации.
//
// Пример использования:
//
//	db, err := sql.Open("pgx", dsn)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	repoFactory, err := storage.NewStorageFactory("pgx", db)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	binaryFactory := storage.NewBinaryDataFactory(cfg)
//	binaryStorage := binaryFactory.BinaryData()
//
//	userRepo := repoFactory.User()
//	textRepo := repoFactory.TextData()
//	credentialRepo := repoFactory.Credential()
//	bankCardRepo := repoFactory.BankCard()
//	binaryDataRepo := repoFactory.BinaryData() // метаданные в БД
//
//	// сервисный слой получает уже готовые реализации репозиториев и бинарного хранилища
package storage
