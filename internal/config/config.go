package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	defaultServerAddress    = "localhost:8080"
	defaultBaseURL          = "http://localhost:8080"
	defaultFileStoragePath  = "shortenest_url.txt"
	defaultDatabaseDSN      = ""
	defaultAuditFile        = ""
	defaultAuditURL         = ""
	defaultConfigFilePath   = ""
	defaultTLSCertFile      = ""
	defaultTLSKeyFile       = ""
	defaultHTTPS            = false
	defaultTimeout          = time.Second * 4
	defaultIdleTimeout      = time.Second * 30
	defaultMaxConns         = 30
	defaultMinConns         = 3
	defaultMaxConnLifetime  = time.Duration(30) * time.Minute
	defaultMaxConnIdleTime  = time.Duration(10) * time.Minute
	defaultHealthCheckDelay = time.Duration(30) * time.Second

	// ProdEnv identifies the production logger and runtime mode.
	ProdEnv = "production"
	// SecretKey is the default AES key used to encrypt the user cookie.
	SecretKey = "super_secret_key"
)

// Config contains application settings loaded from flags and environment variables.
type Config struct {
	// AppEnv selects application mode, including logger configuration.
	AppEnv string `env:"APP_ENV" envDefault:"development"`
	HTTPServer
	// BaseAddressShort is the public base URL used to build shortened links.
	BaseAddressShort string `env:"BASE_URL"`
	// FileStoragePath points to the JSONL file used by the in-memory repository.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// SecretKey is the AES key used by the authentication service.
	SecretKey string `env:"SECRET_KEY"`
}

// HTTPServer contains HTTP, storage, and audit settings for the service.
type HTTPServer struct {
	// ServerAddress is the bind address of the HTTP server.
	ServerAddress string `env:"SERVER_ADDRESS"`
	// BaseURL duplicates BaseAddressShort for callers that read HTTP server settings directly.
	BaseURL string `env:"-"`
	// StorageType names the configured storage backend.
	StorageType string `env:"STORAGE_TYPE"`
	// AuditFile enables file-based audit logging when it is not empty.
	AuditFile string `env:"AUDIT_FILE"`
	// AuditURL enables HTTP audit delivery when it is not empty.
	AuditURL string `env:"AUDIT_URL"`
	// EnableHTTPS switches the web server to TLS.
	EnableHTTPS bool `env:"ENABLE_HTTPS"`
	// TLSCertFile is a path to the TLS public certificate file.
	TLSCertFile string `env:"TLS_CERT_FILE"`
	// TLSKeyFile is a path to the TLS private key file.
	TLSKeyFile string `env:"TLS_KEY_FILE"`
	// Postgres holds PostgreSQL connection pool settings.
	Postgres *PostgresConfig
	// Timeout is applied to HTTP read and write operations.
	Timeout time.Duration
	// IdleTimeout is applied to idle HTTP connections.
	IdleTimeout time.Duration
}

// PostgresConfig contains PostgreSQL DSN and connection pool settings.
type PostgresConfig struct {
	// DataBaseDSN is the PostgreSQL connection string.
	DataBaseDSN string `env:"DATABASE_DSN"`
	// MaxConns limits the maximum number of PostgreSQL connections.
	MaxConns int32 `env:"MAX_CONNS"`
	// MinConns keeps a minimum number of PostgreSQL connections ready.
	MinConns int32 `env:"MIN_CONNS"`
	// MaxConnLifetime limits how long a PostgreSQL connection can be reused.
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME"`
	// MaxConnIdleTime limits how long an idle PostgreSQL connection is kept.
	MaxConnIdleTime time.Duration `env:"MAX_CONN_IDLE_TIME"`
	// HealthCheckPeriod controls how often the PostgreSQL pool checks idle connections.
	HealthCheckPeriod time.Duration `env:"HEALTH_CHECK_PERIOD"`
}

type cliConfig struct {
	serverAddress    string
	baseAddressShort string
	fileStoragePath  string
	databaseDSN      string
	auditFile        string
	auditURL         string
	configFilePath   string
	enableHTTPS      bool
	tlsCertFile      string
	tlsKeyFile       string
}

type fileConfig struct {
	ServerAddress   *string `json:"server_address"`
	BaseURL         *string `json:"base_url"`
	FileStoragePath *string `json:"file_storage_path"`
	DatabaseDSN     *string `json:"database_dsn"`
	EnableHTTPS     *bool   `json:"enable_https"`
	TLSCertFile     *string `json:"tls_cert_file"`
	TLSKeyFile      *string `json:"tls_key_file"`
	AuditFile       *string `json:"audit_file"`
	AuditURL        *string `json:"audit_url"`
}

// InitConfig parses flags and environment variables into Config.
func InitConfig() (*Config, error) {
	cli := cliConfig{
		serverAddress:    defaultServerAddress,
		baseAddressShort: defaultBaseURL,
		fileStoragePath:  defaultFileStoragePath,
		databaseDSN:      defaultDatabaseDSN,
		auditFile:        defaultAuditFile,
		auditURL:         defaultAuditURL,
		configFilePath:   configFilePathFromEnv(),
		enableHTTPS:      defaultHTTPS,
		tlsCertFile:      defaultTLSCertFile,
		tlsKeyFile:       defaultTLSKeyFile,
	}

	flag.StringVar(&cli.serverAddress, "a", cli.serverAddress, "host address server")
	flag.StringVar(&cli.baseAddressShort, "b", cli.baseAddressShort, "base url for short url")
	flag.StringVar(&cli.fileStoragePath, "f", cli.fileStoragePath, "file storage path")
	flag.StringVar(&cli.databaseDSN, "d", cli.databaseDSN, "dsn connection to db")
	flag.StringVar(&cli.auditFile, "audit-file", cli.auditFile, "file audit storage path")
	flag.StringVar(&cli.auditURL, "audit-url", cli.auditURL, "audit url")
	flag.StringVar(&cli.configFilePath, "c", cli.configFilePath, "json config file path")
	flag.StringVar(&cli.configFilePath, "config", cli.configFilePath, "json config file path")
	flag.BoolVar(&cli.enableHTTPS, "s", cli.enableHTTPS, "enable HTTPS")
	flag.StringVar(&cli.tlsCertFile, "tls-cert-file", cli.tlsCertFile, "TLS public certificate file path")
	flag.StringVar(&cli.tlsKeyFile, "tls-key-file", cli.tlsKeyFile, "TLS private key file path")

	flag.Parse()

	cfg := newDefaultConfig()
	if cli.configFilePath != defaultConfigFilePath {
		if err := loadConfigFile(&cfg, cli.configFilePath); err != nil {
			return nil, err
		}
	}

	flags := parsedFlags()
	applyFlags(&cfg, cli, flags)

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	if flags["s"] && cli.enableHTTPS {
		cfg.EnableHTTPS = true
	}
	cfg.BaseURL = cfg.BaseAddressShort
	cfg.SecretKey = SecretKey
	return &cfg, nil
}

func newDefaultConfig() Config {
	return Config{
		HTTPServer: HTTPServer{
			ServerAddress: defaultServerAddress,
			BaseURL:       defaultBaseURL,
			Timeout:       defaultTimeout,
			IdleTimeout:   defaultIdleTimeout,
			AuditFile:     defaultAuditFile,
			AuditURL:      defaultAuditURL,
			EnableHTTPS:   defaultHTTPS,
			TLSCertFile:   defaultTLSCertFile,
			TLSKeyFile:    defaultTLSKeyFile,
			Postgres: &PostgresConfig{
				DataBaseDSN:       defaultDatabaseDSN,
				MaxConns:          defaultMaxConns,
				MinConns:          defaultMinConns,
				MaxConnLifetime:   defaultMaxConnLifetime,
				MaxConnIdleTime:   defaultMaxConnIdleTime,
				HealthCheckPeriod: defaultHealthCheckDelay,
			},
		},
		BaseAddressShort: defaultBaseURL,
		FileStoragePath:  defaultFileStoragePath,
	}
}

func configFilePathFromEnv() string {
	if configFilePath, ok := os.LookupEnv("CONFIG"); ok {
		return configFilePath
	}
	return defaultConfigFilePath
}

func parsedFlags() map[string]bool {
	flags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		flags[f.Name] = true
	})
	return flags
}

func applyFlags(cfg *Config, cli cliConfig, flags map[string]bool) {
	if flags["a"] {
		cfg.ServerAddress = cli.serverAddress
	}
	if flags["b"] {
		cfg.BaseAddressShort = cli.baseAddressShort
		cfg.BaseURL = cli.baseAddressShort
	}
	if flags["f"] {
		cfg.FileStoragePath = cli.fileStoragePath
	}
	if flags["d"] {
		cfg.Postgres.DataBaseDSN = cli.databaseDSN
	}
	if flags["audit-file"] {
		cfg.AuditFile = cli.auditFile
	}
	if flags["audit-url"] {
		cfg.AuditURL = cli.auditURL
	}
	if flags["s"] {
		cfg.EnableHTTPS = cli.enableHTTPS
	}
	if flags["tls-cert-file"] {
		cfg.TLSCertFile = cli.tlsCertFile
	}
	if flags["tls-key-file"] {
		cfg.TLSKeyFile = cli.tlsKeyFile
	}
}

func loadConfigFile(cfg *Config, configFilePath string) error {
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	var fileCfg fileConfig
	if err := json.Unmarshal(content, &fileCfg); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}
	applyConfigFile(cfg, fileCfg)
	return nil
}

func applyConfigFile(cfg *Config, fileCfg fileConfig) {
	if fileCfg.ServerAddress != nil {
		cfg.ServerAddress = *fileCfg.ServerAddress
	}
	if fileCfg.BaseURL != nil {
		cfg.BaseAddressShort = *fileCfg.BaseURL
		cfg.BaseURL = *fileCfg.BaseURL
	}
	if fileCfg.FileStoragePath != nil {
		cfg.FileStoragePath = *fileCfg.FileStoragePath
	}
	if fileCfg.DatabaseDSN != nil {
		cfg.Postgres.DataBaseDSN = *fileCfg.DatabaseDSN
	}
	if fileCfg.EnableHTTPS != nil {
		cfg.EnableHTTPS = *fileCfg.EnableHTTPS
	}
	if fileCfg.TLSCertFile != nil {
		cfg.TLSCertFile = *fileCfg.TLSCertFile
	}
	if fileCfg.TLSKeyFile != nil {
		cfg.TLSKeyFile = *fileCfg.TLSKeyFile
	}
	if fileCfg.AuditFile != nil {
		cfg.AuditFile = *fileCfg.AuditFile
	}
	if fileCfg.AuditURL != nil {
		cfg.AuditURL = *fileCfg.AuditURL
	}
}
