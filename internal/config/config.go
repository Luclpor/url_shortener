package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
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
	BaseAddressShort string
	// FileStoragePath points to the JSONL file used by the in-memory repository.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// SecretKey is the AES key used by the authentication service.
	SecretKey string `env:"SECRET_KEY"`
}

// HTTPServer contains HTTP, storage, and audit settings for the service.
type HTTPServer struct {
	// ServerAddress is the bind address of the HTTP server.
	ServerAddress string `env:"SERVER_ADDRESS"`
	// BaseURL is an optional source base URL for deployments that need it.
	BaseURL string `env:"BASE_URL"`
	// StorageType names the configured storage backend.
	StorageType string `env:"STORAGE_TYPE"`
	// AuditFile enables file-based audit logging when it is not empty.
	AuditFile string `env:"AUDIT_FILE"`
	// AuditURL enables HTTP audit delivery when it is not empty.
	AuditURL string `env:"AUDIT_URL"`
	// EnableHTTPS switches the web server to TLS.
	EnableHTTPS bool `env:"ENABLE_HTTPS"`
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

// InitConfig parses flags and environment variables into Config.
func InitConfig() (*Config, error) {
	h := flag.String("a", "localhost:8080", "host address server")
	b := flag.String("b", "http://localhost:8080", "base url for short url")
	f := flag.String("f", "shortenest_url.txt", "file storage path")
	d := flag.String("d", "", "dsn connection to db")
	auditFile := flag.String("audit-file", "", "file audit storage path")
	auditURL := flag.String("audit-url", "", "audit url")
	enableHTTPS := flag.Bool("s", false, "enable HTTPS")

	flag.Parse()

	cfg := Config{
		HTTPServer: HTTPServer{
			ServerAddress: *h,
			Timeout:       time.Second * 4,
			IdleTimeout:   time.Second * 30,
			AuditFile:     *auditFile,
			AuditURL:      *auditURL,
			EnableHTTPS:   *enableHTTPS,
			Postgres: &PostgresConfig{
				DataBaseDSN:       *d,
				MaxConns:          30,
				MinConns:          3,
				MaxConnLifetime:   time.Duration(30) * time.Minute,
				MaxConnIdleTime:   time.Duration(10) * time.Minute,
				HealthCheckPeriod: time.Duration(30) * time.Second,
			},
		},
		BaseAddressShort: *b,
		FileStoragePath:  *f,
	}

	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	cfg.EnableHTTPS = cfg.EnableHTTPS || *enableHTTPS
	cfg.SecretKey = SecretKey
	return &cfg, nil
}
