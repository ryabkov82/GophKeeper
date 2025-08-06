package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("Default config", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("default", flag.PanicOnError)
		os.Args = []string{"cmd"}

		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, "localhost:50051", cfg.ServerAddress)
		require.Equal(t, false, cfg.UseTLS)
		require.Equal(t, false, cfg.TLSSkipVerify)
		require.Equal(t, "certs/ca.crt", cfg.CACertPath)
		require.Equal(t, 10*time.Second, cfg.Timeout)
		require.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("JSON config", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("json", flag.PanicOnError)
		os.Args = []string{"cmd"}

		json := `{
			"server_address": "jsonhost:50000",
			"use_tls": false,
			"tls_skip_verify": true,
			"ca_cert_path": "/custom/ca.pem",
			"timeout": "30s",
			"log_level": "debug"
		}`
		tmp := filepath.Join(t.TempDir(), "config.json")
		require.NoError(t, os.WriteFile(tmp, []byte(json), 0644))
		t.Setenv("CONFIG", tmp)

		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, "jsonhost:50000", cfg.ServerAddress)
		require.Equal(t, false, cfg.UseTLS)
		require.Equal(t, true, cfg.TLSSkipVerify)
		require.Equal(t, "/custom/ca.pem", cfg.CACertPath)
		require.Equal(t, 30*time.Second, cfg.Timeout)
		require.Equal(t, "debug", cfg.LogLevel)
	})

	t.Run("Environment override", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("env", flag.PanicOnError)
		os.Args = []string{"cmd"}
		t.Setenv("SERVER_ADDRESS", "envhost:60090")
		t.Setenv("USE_TLS", "false")
		t.Setenv("TLS_SKIP_VERIFY", "true")
		t.Setenv("TIMEOUT", "15s")
		t.Setenv("LOG_LEVEL", "warn")

		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, "envhost:60090", cfg.ServerAddress)
		require.Equal(t, false, cfg.UseTLS)
		require.Equal(t, true, cfg.TLSSkipVerify)
		require.Equal(t, 15*time.Second, cfg.Timeout)
		require.Equal(t, "warn", cfg.LogLevel)
	})

	t.Run("Invalid server address", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("invalid_addr", flag.PanicOnError)
		os.Args = []string{"cmd"}
		t.Setenv("SERVER_ADDRESS", "invalid_address")

		_, err := Load()
		require.Error(t, err)
	})

	t.Run("Invalid TLS env values", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("bad_tls", flag.PanicOnError)
		os.Args = []string{"cmd"}
		t.Setenv("USE_TLS", "not_a_boolean")

		_, err := Load()
		require.Error(t, err)
	})

	t.Run("Flag override", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("flags", flag.PanicOnError)
		os.Args = []string{"cmd",
			"-addr=flaghost:60070",
			"-tls=false",
			"-skip-verify=true",
			"-timeout=5s",
			"-log-level=error"}

		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, "flaghost:60070", cfg.ServerAddress)
		require.Equal(t, false, cfg.UseTLS)
		require.Equal(t, true, cfg.TLSSkipVerify)
		require.Equal(t, 5*time.Second, cfg.Timeout)
		require.Equal(t, "error", cfg.LogLevel)
	})

	t.Run("validateServerAddress", func(t *testing.T) {
		cases := []struct {
			addr    string
			wantErr bool
		}{
			{"localhost:8080", true},
			{"example.com:49152", false},
			{"127.0.0.1:65535", false},
			{"[::1]:50000", false},
			{"invalid", true},
			{"", true},
			{"localhost", true},
			{"localhost:1000", true},  // порт слишком маленький
			{"localhost:70000", true}, // порт слишком большой
		}

		for _, c := range cases {
			err := validateServerAddress(c.addr)
			if c.wantErr {
				require.Error(t, err, "addr: %s", c.addr)
			} else {
				require.NoError(t, err, "addr: %s", c.addr)
			}
		}
	})

	t.Run("mergeConfigs", func(t *testing.T) {
		dst := &ClientConfig{
			ServerAddress: "default:8080",
			UseTLS:        true,
			Timeout:       10 * time.Second,
		}
		src := &ClientConfig{
			ServerAddress: "merged:9090",
			TLSSkipVerify: true,
			LogLevel:      "debug",
			Timeout:       20 * time.Second,
		}

		mergeConfigs(dst, src)

		require.Equal(t, "merged:9090", dst.ServerAddress)
		require.True(t, dst.UseTLS) // не перезаписывалось
		require.True(t, dst.TLSSkipVerify)
		require.Equal(t, "debug", dst.LogLevel)
		require.Equal(t, 20*time.Second, dst.Timeout)
	})

	t.Run("CA cert validation", func(t *testing.T) {
		dir := t.TempDir()
		caCert := filepath.Join(dir, "ca.crt")
		require.NoError(t, os.WriteFile(caCert, []byte("test cert"), 0644))

		flag.CommandLine = flag.NewFlagSet("ca_check", flag.PanicOnError)
		os.Args = []string{"cmd", "-ca-cert", caCert, "-tls=true"}

		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, caCert, cfg.CACertPath)

		// Проверка несуществующего CA
		os.Args = []string{"cmd", "-ca-cert=nonexistent.crt", "-tls=true"}
		_, err = Load()
		require.Error(t, err)
	})
}
