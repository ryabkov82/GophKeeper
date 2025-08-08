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
	"time"
)

// ClientConfig содержит параметры конфигурации клиента
type ClientConfig struct {
	ServerAddress string        `json:"server_address" env:"SERVER_ADDRESS"`
	UseTLS        bool          `json:"use_tls" env:"USE_TLS"`
	TLSSkipVerify bool          `json:"tls_skip_verify" env:"TLS_SKIP_VERIFY"`
	CACertPath    string        `json:"ca_cert_path" env:"CA_CERT_PATH"`
	Timeout       time.Duration `json:"timeout" env:"TIMEOUT"`
	ConfigPath    string        `json:"-" env:"CONFIG"` // Путь к конфиг-файлу
	LogLevel      string        `json:"log_level" env:"LOG_LEVEL"`
}

const (
	minDynamicPort = 49152
	maxPort        = 65535
)

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		ServerAddress: "localhost:50051",
		UseTLS:        false,
		TLSSkipVerify: false,
		CACertPath:    "certs/ca.crt",
		Timeout:       10 * time.Second,
		LogLevel:      "info",
	}
}

// Load загружает конфигурацию в порядке:
// 1) значения по умолчанию
// 2) JSON-файл (если указан)
// 3) флаги командной строки
// 4) переменные окружения
func Load() (*ClientConfig, error) {
	cfg := DefaultConfig()

	// 1. Загрузка из JSON-файла
	if configFile := getConfigFilePath(); configFile != "" {
		fileCfg, err := loadFromJSON(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
		mergeConfigs(cfg, fileCfg)
	}

	// 2. Загрузка из флагов
	if err := loadFromFlags(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	// 3. Загрузка из env
	if err := loadFromEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to load env vars: %w", err)
	}

	// Валидация
	if err := validateServerAddress(cfg.ServerAddress); err != nil {
		return nil, fmt.Errorf("server address validation failed: %w", err)
	}

	if cfg.UseTLS {
		if _, err := os.Stat(cfg.CACertPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("CA certificate not found: %s", cfg.CACertPath)
		}
	}

	return cfg, nil
}

// Валидация адреса сервера
func validateServerAddress(addr string) error {
	if addr == "" {
		return errors.New("server address cannot be empty")
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	if host == "" {
		return errors.New("host cannot be empty")
	}

	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < minDynamicPort || portNum > maxPort {
		return fmt.Errorf("port must be between %d and %d", minDynamicPort, maxPort)
	}

	return nil
}

// Вспомогательные методы:

func getConfigFilePath() string {

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

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		return envConfig
	}
	return ""
}

func loadFromJSON(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ClientConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// Дополнительная валидация
	if cfg.Timeout < 0 {
		return nil, errors.New("timeout cannot be negative")
	}
	return &cfg, nil
}

func mergeConfigs(dst, src *ClientConfig) {
	if src.ServerAddress != "" {
		dst.ServerAddress = src.ServerAddress
	}
	if src.UseTLS {
		dst.UseTLS = src.UseTLS
	}
	if src.TLSSkipVerify {
		dst.TLSSkipVerify = src.TLSSkipVerify
	}
	if src.CACertPath != "" {
		dst.CACertPath = src.CACertPath
	}
	if src.Timeout != 0 {
		dst.Timeout = src.Timeout
	}
	if src.LogLevel != "" {
		dst.LogLevel = src.LogLevel
	}
}

func loadFromFlags(cfg *ClientConfig) error {
	// Создаем новый изолированный FlagSet
	flagset := flag.NewFlagSet("client-flags", flag.ContinueOnError)
	var validationErr error

	// Регистрируем все флаги в изолированном FlagSet
	flagset.Func("addr", "Server address (host:port)", func(s string) error {
		if err := validateServerAddress(s); err != nil {
			validationErr = fmt.Errorf("invalid server address: %w", err)
			return validationErr
		}
		cfg.ServerAddress = s
		return nil
	})

	flagset.BoolVar(&cfg.UseTLS, "tls", cfg.UseTLS, "Use TLS")
	flagset.BoolVar(&cfg.TLSSkipVerify, "skip-verify", cfg.TLSSkipVerify, "Skip TLS verification")
	flagset.StringVar(&cfg.CACertPath, "ca-cert", cfg.CACertPath, "Path to CA certificate")
	flagset.DurationVar(&cfg.Timeout, "timeout", cfg.Timeout, "Connection timeout")
	flagset.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Logging level")

	// Парсим аргументы, игнорируя нераспознанные флаги
	if err := flagset.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return err // Возвращаем ошибку помощи отдельно
		}
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	return validationErr
}

func loadFromEnv(cfg *ClientConfig) error {
	if val := os.Getenv("SERVER_ADDRESS"); val != "" {
		if err := validateServerAddress(val); err != nil {
			return err
		}
		cfg.ServerAddress = val
	}

	if val := os.Getenv("USE_TLS"); val != "" {
		if v, err := strconv.ParseBool(val); err == nil {
			cfg.UseTLS = v
		} else {
			return fmt.Errorf("invalid USE_TLS value: %w", err)
		}
	}

	if val := os.Getenv("TLS_SKIP_VERIFY"); val != "" {
		if v, err := strconv.ParseBool(val); err == nil {
			cfg.TLSSkipVerify = v
		} else {
			return fmt.Errorf("invalid TLS_SKIP_VERIFY value: %w", err)
		}
	}

	if val := os.Getenv("CA_CERT_PATH"); val != "" {
		cfg.CACertPath = val
	}

	if val := os.Getenv("TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			cfg.Timeout = d
		} else {
			return fmt.Errorf("invalid TIMEOUT value: %w", err)
		}
	}

	if val := os.Getenv("LOG_LEVEL"); val != "" {
		cfg.LogLevel = val
	}

	return nil
}

func (c *ClientConfig) UnmarshalJSON(data []byte) error {
	type Alias ClientConfig
	aux := &struct {
		Timeout string `json:"timeout"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Timeout != "" {
		duration, err := time.ParseDuration(aux.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %w", err)
		}
		c.Timeout = duration
	}

	return nil
}
