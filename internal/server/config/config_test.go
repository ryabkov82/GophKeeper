package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("Default config", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("default", flag.PanicOnError)
		os.Args = []string{"cmd"}

		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, "localhost:50051", cfg.GRPCServerAddr)
		require.Equal(t, "info", cfg.LogLevel)
		require.Equal(t, false, cfg.EnableTLS)
	})

	t.Run("JSON config", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("json", flag.PanicOnError)
		os.Args = []string{"cmd"}

		json := `{
			"grpc_server_address": "jsonhost:55555",
			"log_level": "debug",
			"enable_tls": true,
			"ssl_cert_file": "certs/cert.pem",
			"ssl_key_file": "certs/key.pem",
			"jwt_secret": "json_secret_1234567890123456789012345678"
		}`
		tmp := filepath.Join(t.TempDir(), "config.json")
		require.NoError(t, os.WriteFile(tmp, []byte(json), 0644))
		t.Setenv("CONFIG", tmp)

		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, "jsonhost:55555", cfg.GRPCServerAddr)
		require.Equal(t, "debug", cfg.LogLevel)
		require.Equal(t, true, cfg.EnableTLS)
	})

	t.Run("Environment override", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("env", flag.PanicOnError)
		os.Args = []string{"cmd"}
		t.Setenv("GRPC_SERVER_ADDRESS", "envhost:50123")
		t.Setenv("LOG_LEVEL", "warn")
		t.Setenv("JWT_SECRET", "env_secret_123456789012345678901234567890")

		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, "envhost:50123", cfg.GRPCServerAddr)
		require.Equal(t, "warn", cfg.LogLevel)
		require.Equal(t, "env_secret_123456789012345678901234567890", cfg.JwtKey)
	})

	t.Run("Invalid gRPC address", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("invalid_grpc", flag.PanicOnError)
		os.Args = []string{"cmd"}
		t.Setenv("GRPC_SERVER_ADDRESS", "invalid")

		_, err := Load()
		require.Error(t, err)
	})

	t.Run("Invalid JWT secret from env", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("bad_jwt", flag.PanicOnError)
		os.Args = []string{"cmd"}
		t.Setenv("JWT_SECRET", "short")

		_, err := Load()
		require.Error(t, err)
	})

	t.Run("Flag override", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("flags", flag.PanicOnError)
		os.Args = []string{"cmd", "-grpc", "flaghost:52345", "-l", "error", "-s=true"}

		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, "flaghost:52345", cfg.GRPCServerAddr)
		require.Equal(t, "error", cfg.LogLevel)
		require.True(t, cfg.EnableTLS)
	})

	t.Run("validateCertFiles", func(t *testing.T) {
		dir := t.TempDir()
		cert := filepath.Join(dir, "cert.pem")
		key := filepath.Join(dir, "key.pem")

		require.NoError(t, os.WriteFile(cert, []byte("data"), 0644))
		require.NoError(t, os.WriteFile(key, []byte("data"), 0644))

		err := validateCertFiles(cert, key)
		require.NoError(t, err)

		err = validateCertFiles("nonexistent.pem", key)
		require.Error(t, err)

		err = validateCertFiles(cert, "missing.pem")
		require.Error(t, err)
	})

	t.Run("validateGRPCServerAddr", func(t *testing.T) {
		cases := []struct {
			addr    string
			wantErr bool
		}{
			{"localhost:50051", false},
			{"127.0.0.1:49152", false},
			{"[::1]:60000", false},
			{"localhost", true},
			{"localhost:1000", true},
			{"", true},
		}

		for _, c := range cases {
			err := validateGRPCServerAddr(c.addr)
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		}
	})

	t.Run("mergeConfigs", func(t *testing.T) {
		dst := &Config{
			GRPCServerAddr: "default:1234",
			LogLevel:       "info",
		}
		src := &Config{
			GRPCServerAddr: "merged:5555",
			LogLevel:       "debug",
			JwtKey:         "mergedsecret",
			EnableTLS:      true,
		}

		mergeConfigs(dst, src)

		require.Equal(t, "merged:5555", dst.GRPCServerAddr)
		require.Equal(t, "debug", dst.LogLevel)
		require.Equal(t, "mergedsecret", dst.JwtKey)
		require.True(t, dst.EnableTLS)
	})
}
