package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v6"
)

const (
	ProdEnv   = "production"
	SecretKey = "super_secret_key"
)

type Config struct {
	AppEnv string `env:"APP_ENV" envDefault:"development"`
	HTTPServer
	BaseAddressShort string
	FileStoragePath  string `env:"FILE_STORAGE_PATH"`
	SecretKey        string `env:"SECRET_KEY"`
}

type HTTPServer struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	StorageType   string `env:"STORAGE_TYPE"`
	Postgres      *PostgresConfig
	Timeout       time.Duration
	IdleTimeout   time.Duration
}

type PostgresConfig struct {
	DataBaseDSN       string        `env:"DATABASE_DSN"`
	MaxConns          int32         `env:"MAX_CONNS"`
	MinConns          int32         `env:"MIN_CONNS"`
	MaxConnLifetime   time.Duration `env:"MAX_CONN_LIFETIME"`
	MaxConnIdleTime   time.Duration `env:"MAX_CONN_IDLE_TIME"`
	HealthCheckPeriod time.Duration `env:"HEALTH_CHECK_PERIOD"`
}

func InitConfig() (*Config, error) {
	h := flag.String("a", "localhost:8080", "host address server")
	b := flag.String("b", "http://localhost:8080", "base url for short url")
	f := flag.String("f", "shortenest_url.txt", "file storage path")
	d := flag.String("d", "", "dsn connection to db")

	flag.Parse()

	cfg := Config{
		HTTPServer: HTTPServer{
			ServerAddress: *h,
			Timeout:       time.Second * 4,
			IdleTimeout:   time.Second * 30,
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
	cfg.SecretKey = SecretKey
	return &cfg, nil
}
