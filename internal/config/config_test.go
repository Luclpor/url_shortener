package config

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestInitConfigEnablesHTTPSWithFlag(t *testing.T) {
	resetConfigTestState(t, []string{"shortener", "-s"})
	clearConfigEnv(t)

	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}
	if !cfg.EnableHTTPS {
		t.Fatal("InitConfig() should enable HTTPS with -s flag")
	}
}

func TestInitConfigEnablesHTTPSWithEnv(t *testing.T) {
	resetConfigTestState(t, []string{"shortener"})
	clearConfigEnv(t)
	t.Setenv("ENABLE_HTTPS", "true")

	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}
	if !cfg.EnableHTTPS {
		t.Fatal("InitConfig() should enable HTTPS with ENABLE_HTTPS=true")
	}
}

func TestInitConfigFlagEnablesHTTPSWhenEnvIsFalse(t *testing.T) {
	resetConfigTestState(t, []string{"shortener", "-s"})
	clearConfigEnv(t)
	t.Setenv("ENABLE_HTTPS", "false")

	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}
	if !cfg.EnableHTTPS {
		t.Fatal("InitConfig() should enable HTTPS when -s is set")
	}
}

func TestInitConfigLoadsJSONConfigFileFromFlag(t *testing.T) {
	configPath := writeConfigFile(t, `{
		"server_address": "127.0.0.1:9090",
		"base_url": "https://short.test",
		"file_storage_path": "/tmp/urls.db",
		"database_dsn": "postgres://user:pass@localhost:5432/shortener",
		"enable_https": true,
		"audit_file": "/tmp/audit.log",
		"audit_url": "http://audit.test/events"
	}`)
	resetConfigTestState(t, []string{"shortener", "-c", configPath})
	clearConfigEnv(t)

	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	assertConfigValue(t, "ServerAddress", cfg.ServerAddress, "127.0.0.1:9090")
	assertConfigValue(t, "BaseAddressShort", cfg.BaseAddressShort, "https://short.test")
	assertConfigValue(t, "BaseURL", cfg.BaseURL, "https://short.test")
	assertConfigValue(t, "FileStoragePath", cfg.FileStoragePath, "/tmp/urls.db")
	assertConfigValue(t, "DataBaseDSN", cfg.Postgres.DataBaseDSN, "postgres://user:pass@localhost:5432/shortener")
	assertConfigValue(t, "AuditFile", cfg.AuditFile, "/tmp/audit.log")
	assertConfigValue(t, "AuditURL", cfg.AuditURL, "http://audit.test/events")
	if !cfg.EnableHTTPS {
		t.Fatal("InitConfig() should enable HTTPS from config file")
	}
}

func TestInitConfigLoadsJSONConfigFileFromEnv(t *testing.T) {
	configPath := writeConfigFile(t, `{
		"server_address": "127.0.0.1:7070",
		"base_url": "https://env-config.test",
		"file_storage_path": "/tmp/env-config.db",
		"database_dsn": "config-env-dsn",
		"enable_https": true
	}`)
	resetConfigTestState(t, []string{"shortener"})
	clearConfigEnv(t)
	t.Setenv("CONFIG", configPath)

	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	assertConfigValue(t, "ServerAddress", cfg.ServerAddress, "127.0.0.1:7070")
	assertConfigValue(t, "BaseAddressShort", cfg.BaseAddressShort, "https://env-config.test")
	assertConfigValue(t, "BaseURL", cfg.BaseURL, "https://env-config.test")
	assertConfigValue(t, "FileStoragePath", cfg.FileStoragePath, "/tmp/env-config.db")
	assertConfigValue(t, "DataBaseDSN", cfg.Postgres.DataBaseDSN, "config-env-dsn")
	if !cfg.EnableHTTPS {
		t.Fatal("InitConfig() should enable HTTPS from CONFIG file")
	}
}

func TestInitConfigFlagsOverrideJSONConfigFile(t *testing.T) {
	configPath := writeConfigFile(t, `{
		"server_address": "127.0.0.1:9090",
		"base_url": "https://config.test",
		"file_storage_path": "/tmp/config.db",
		"database_dsn": "config-dsn",
		"enable_https": false,
		"audit_file": "/tmp/config-audit.log",
		"audit_url": "http://config-audit.test/events"
	}`)
	resetConfigTestState(t, []string{
		"shortener",
		"-config", configPath,
		"-a", "127.0.0.1:6060",
		"-b", "https://flag.test",
		"-f", "/tmp/flag.db",
		"-d", "flag-dsn",
		"-audit-file", "/tmp/flag-audit.log",
		"-audit-url", "http://flag-audit.test/events",
		"-s",
	})
	clearConfigEnv(t)

	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	assertConfigValue(t, "ServerAddress", cfg.ServerAddress, "127.0.0.1:6060")
	assertConfigValue(t, "BaseAddressShort", cfg.BaseAddressShort, "https://flag.test")
	assertConfigValue(t, "BaseURL", cfg.BaseURL, "https://flag.test")
	assertConfigValue(t, "FileStoragePath", cfg.FileStoragePath, "/tmp/flag.db")
	assertConfigValue(t, "DataBaseDSN", cfg.Postgres.DataBaseDSN, "flag-dsn")
	assertConfigValue(t, "AuditFile", cfg.AuditFile, "/tmp/flag-audit.log")
	assertConfigValue(t, "AuditURL", cfg.AuditURL, "http://flag-audit.test/events")
	if !cfg.EnableHTTPS {
		t.Fatal("InitConfig() should enable HTTPS from -s flag")
	}
}

func TestInitConfigEnvOverridesJSONConfigFile(t *testing.T) {
	configPath := writeConfigFile(t, `{
		"server_address": "127.0.0.1:9090",
		"base_url": "https://config.test",
		"file_storage_path": "/tmp/config.db",
		"database_dsn": "config-dsn",
		"enable_https": true,
		"audit_file": "/tmp/config-audit.log",
		"audit_url": "http://config-audit.test/events"
	}`)
	resetConfigTestState(t, []string{"shortener", "-c", configPath})
	clearConfigEnv(t)
	t.Setenv("SERVER_ADDRESS", "127.0.0.1:5050")
	t.Setenv("BASE_URL", "https://env.test")
	t.Setenv("FILE_STORAGE_PATH", "/tmp/env.db")
	t.Setenv("DATABASE_DSN", "env-dsn")
	t.Setenv("ENABLE_HTTPS", "false")
	t.Setenv("AUDIT_FILE", "/tmp/env-audit.log")
	t.Setenv("AUDIT_URL", "http://env-audit.test/events")

	cfg, err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	assertConfigValue(t, "ServerAddress", cfg.ServerAddress, "127.0.0.1:5050")
	assertConfigValue(t, "BaseAddressShort", cfg.BaseAddressShort, "https://env.test")
	assertConfigValue(t, "BaseURL", cfg.BaseURL, "https://env.test")
	assertConfigValue(t, "FileStoragePath", cfg.FileStoragePath, "/tmp/env.db")
	assertConfigValue(t, "DataBaseDSN", cfg.Postgres.DataBaseDSN, "env-dsn")
	assertConfigValue(t, "AuditFile", cfg.AuditFile, "/tmp/env-audit.log")
	assertConfigValue(t, "AuditURL", cfg.AuditURL, "http://env-audit.test/events")
	if cfg.EnableHTTPS {
		t.Fatal("InitConfig() should disable HTTPS with ENABLE_HTTPS=false")
	}
}

func resetConfigTestState(t *testing.T, args []string) {
	t.Helper()

	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	os.Args = args

	t.Cleanup(func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
	})
}

func clearConfigEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"CONFIG",
		"SERVER_ADDRESS",
		"BASE_URL",
		"FILE_STORAGE_PATH",
		"DATABASE_DSN",
		"ENABLE_HTTPS",
		"AUDIT_FILE",
		"AUDIT_URL",
	}
	for _, key := range keys {
		unsetEnv(t, key)
	}
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	oldValue, wasSet := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%q) error = %v", key, err)
	}
	t.Cleanup(func() {
		if wasSet {
			if err := os.Setenv(key, oldValue); err != nil {
				t.Fatalf("Setenv(%q) error = %v", key, err)
			}
			return
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%q) error = %v", key, err)
		}
	})
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}
	return configPath
}

func assertConfigValue(t *testing.T, name, got, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %q, want %q", name, got, want)
	}
}
