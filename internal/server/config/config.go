// Package config предоставляет функционал загрузки и хранения конфигурации сервера GophKeeper.
// Конфигурация может быть загружена из JSON-файла, переменных окружения или флагов командной строки.
// Основное назначение — централизованное хранение всех параметров, необходимых для запуска сервера.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Config содержит параметры конфигурации сервера.
// Все поля имеют JSON-теги для удобства сериализации/десериализации.
//
// Поля:
//
//	GRPCServerAddr — адрес gRPC-сервера в формате host:port, на котором будет запущено приложение.
//	DBConnect      — строка подключения (DSN) к базе данных PostgreSQL.
//	JwtKey         — секретный ключ, используемый для подписи и проверки JWT-токенов.
//	SSLCertFile    — путь к TLS-сертификату, используемому сервером при включённом TLS.
//	SSLKeyFile     — путь к приватному ключу TLS, соответствующему сертификату из SSLCertFile.
//	EnableTLS      — флаг включения TLS (true — использовать HTTPS/gRPC-TLS, false — без шифрования).
//	LogLevel       — уровень логирования. Возможные значения: debug, info, warn, error.
//	BinaryDataStorePath — путь к директории хранения бинарных данных на локальной файловой системе.
type Config struct {
	GRPCServerAddr      string `json:"grpc_server_address"`    // host:port
	DBConnect           string `json:"database_dsn"`           // PostgreSQL DSN
	JwtKey              string `json:"jwt_secret"`             // секрет для JWT
	SSLCertFile         string `json:"ssl_cert_file"`          // путь к TLS сертификату
	SSLKeyFile          string `json:"ssl_key_file"`           // путь к TLS ключу
	EnableTLS           bool   `json:"enable_tls"`             // включить TLS
	LogLevel            string `json:"log_level"`              // Уровень логирования (debug, info, warn, error)
	BinaryDataStorePath string `json:"binary_data_store_path"` // путь к директории для бинарных файлов
}

const (
	minDynamicPort = 49152 // Начало диапазона динамических/частных портов (IANA)
	maxPort        = 65535 // Максимальный допустимый номер порта
)

// validateGRPCServerAddr проверяет корректность адреса gRPC-сервера.
func validateGRPCServerAddr(addr string) error {
	if addr == "" {
		return errors.New("gRPC server address cannot be empty")
	}

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	// Проверка порта
	portNum, err := strconv.Atoi(port)

	if err != nil || portNum < minDynamicPort || portNum > maxPort {
		return fmt.Errorf("port must be between %d and %d", minDynamicPort, maxPort)
	}

	return nil
}

func validateCertFiles(certFile, keyFile string) error {
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return fmt.Errorf("SSL cert not found: %s", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return fmt.Errorf("SSL key not found: %s", keyFile)
	}
	return nil
}

// Load загружает конфигурацию из нескольких источников:
// 1) значения по умолчанию,
// 2) JSON-файл (если указан через флаг или env),
// 3) флаги командной строки,
// 4) переменные окружения.
func Load() (*Config, error) {
	cfg := &Config{
		GRPCServerAddr:      "localhost:50051",
		DBConnect:           "postgres://gophkeeper:gophkeeper@localhost:5432/gophkeeper?sslmode=disable",
		LogLevel:            "info",
		JwtKey:              "your_strong_secret_here",
		EnableTLS:           false,
		SSLCertFile:         "certs/server.crt",
		SSLKeyFile:          "certs/server.key",
		BinaryDataStorePath: "/var/gophkeeper/binary_data",
	}

	// 1. Сначала загрузка из JSON-файла (если указан)
	if configFile := getConfigFilePath(); configFile != "" {
		fileCfg, err := loadFromJSON(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
		mergeConfigs(cfg, fileCfg)
	}

	// 2. Загрузка из флагов командной строки
	if err := loadFromFlags(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	// 3. Переопределение переменными окружения
	if err := loadFromEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	// 4. Валидация адреса gRPC сервера
	if err := validateGRPCServerAddr(cfg.GRPCServerAddr); err != nil {
		return nil, fmt.Errorf("gRPC server address validation failed: %w", err)
	}

	// Валидация SSL файлов если TLS включен
	if cfg.EnableTLS {
		if err := validateCertFiles(cfg.SSLCertFile, cfg.SSLKeyFile); err != nil {
			return nil, fmt.Errorf("TLS configuration invalid: %w", err)
		}
	}

	// Проверка директории для хранения бинарных данных
	if cfg.BinaryDataStorePath != "" {
		if err := os.MkdirAll(cfg.BinaryDataStorePath, 0o755); err != nil {
			return nil, fmt.Errorf("cannot create binary data directory: %w", err)
		}
	}

	return cfg, nil
}

// getConfigFilePath возвращает путь к конфигу из флагов или переменных окружения
func getConfigFilePath() string {
	// Проверка флагов командной строки
	for i, arg := range os.Args[1:] {
		if arg == "-c" || arg == "--config" {
			if i+1 < len(os.Args) {
				return os.Args[i+2]
			}
		}
		if strings.HasPrefix(arg, "-c=") {
			return strings.TrimPrefix(arg, "-c=")
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}
	}
	// Переменная окружения CONFIG
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		return envConfig
	}
	return ""
}

// loadFromJSON читает конфиг из JSON-файла
func loadFromJSON(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// mergeConfigs копирует непустые поля из src в dst
func mergeConfigs(dst, src *Config) {
	if src.GRPCServerAddr != "" {
		dst.GRPCServerAddr = src.GRPCServerAddr
	}
	if src.DBConnect != "" {
		dst.DBConnect = src.DBConnect
	}
	if src.JwtKey != "" {
		dst.JwtKey = src.JwtKey
	}
	if src.EnableTLS {
		dst.EnableTLS = src.EnableTLS
	}
	if src.SSLCertFile != "" {
		dst.SSLCertFile = src.SSLCertFile
	}
	if src.SSLKeyFile != "" {
		dst.SSLKeyFile = src.SSLKeyFile
	}
	if src.LogLevel != "" {
		dst.LogLevel = src.LogLevel
	}
	if src.BinaryDataStorePath != "" {
		dst.BinaryDataStorePath = src.BinaryDataStorePath
	}
}

// loadFromFlags читает конфиг из аргументов командной строки
func loadFromFlags(cfg *Config) error {
	var validationErr error

	flag.Func("grpc", "gRPC server address host:port", func(s string) error {
		if err := validateGRPCServerAddr(s); err != nil {
			validationErr = fmt.Errorf("invalid grpc server address: %w", err)
			return validationErr
		}
		cfg.GRPCServerAddr = s
		return nil
	})

	flag.StringVar(&cfg.LogLevel, "l", cfg.LogLevel, "Log level (debug, info, warn, error)")
	flag.StringVar(&cfg.DBConnect, "db", cfg.DBConnect, "Database connection string")
	flag.BoolVar(&cfg.EnableTLS, "s", cfg.EnableTLS, "Enable TLS server")
	flag.StringVar(&cfg.BinaryDataStorePath, "binary-path", cfg.BinaryDataStorePath, "Path for storing binary data files")

	flag.Parse()

	return validationErr
}

// loadFromEnv читает конфиг из переменных окружения
func loadFromEnv(cfg *Config) error {
	if val := os.Getenv("GRPC_SERVER_ADDRESS"); val != "" {
		if err := validateGRPCServerAddr(val); err != nil {
			return err
		}
		cfg.GRPCServerAddr = val
	}
	if val := os.Getenv("DATABASE_KEEPER"); val != "" {
		cfg.DBConnect = val
	}

	if envJWT := os.Getenv("JWT_SECRET"); envJWT != "" {
		if len(envJWT) < 32 {
			return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
		}
		cfg.JwtKey = envJWT
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if envBinaryPath := os.Getenv("BINARY_DATA_PATH"); envBinaryPath != "" {
		cfg.BinaryDataStorePath = envBinaryPath
	}

	// Обработка HTTPS настроек
	if envEnableHTTPS := os.Getenv("SSL_ENABLE"); envEnableHTTPS != "" {
		if v, err := strconv.ParseBool(envEnableHTTPS); err == nil {
			cfg.EnableTLS = v
		} else {
			return fmt.Errorf("invalid SSL_ENABLE value: %w", err)
		}
	}

	if envCert := os.Getenv("SSL_CERT_FILE"); envCert != "" {
		cfg.SSLCertFile = envCert
	}

	if envKey := os.Getenv("SSL_KEY_FILE"); envKey != "" {
		cfg.SSLKeyFile = envKey
	}

	return nil
}
