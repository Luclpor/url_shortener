package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/spf13/viper"
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
	defaultTrustedSubnet    = ""
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

const (
	keyAppEnv            = "app_env"
	keyServerAddress     = "server_address"
	keyBaseURL           = "base_url"
	keyFileStoragePath   = "file_storage_path"
	keyStorageType       = "storage_type"
	keyAuditFile         = "audit_file"
	keyAuditURL          = "audit_url"
	keyEnableHTTPS       = "enable_https"
	keyTLSCertFile       = "tls_cert_file"
	keyTLSKeyFile        = "tls_key_file"
	keyTrustedSubnet     = "trusted_subnet"
	keyConfigFilePath    = "config"
	keyDatabaseDSN       = "database_dsn"
	keyMaxConns          = "max_conns"
	keyMinConns          = "min_conns"
	keyMaxConnLifetime   = "max_conn_lifetime"
	keyMaxConnIdleTime   = "max_conn_idle_time"
	keyHealthCheckPeriod = "health_check_period"
)

// Config contains application settings loaded from flags and environment variables.
type Config struct {
	// AppEnv selects application mode, including logger configuration.
	AppEnv string
	HTTPServer
	// BaseAddressShort is the public base URL used to build shortened links.
	BaseAddressShort string
	// FileStoragePath points to the JSONL file used by the in-memory repository.
	FileStoragePath string
	// SecretKey is the AES key used by the authentication service.
	SecretKey string
}

// HTTPServer contains HTTP, storage, and audit settings for the service.
type HTTPServer struct {
	// ServerAddress is the bind address of the HTTP server.
	ServerAddress string
	// BaseURL duplicates BaseAddressShort for callers that read HTTP server settings directly.
	BaseURL string
	// StorageType names the configured storage backend.
	StorageType string
	// AuditFile enables file-based audit logging when it is not empty.
	AuditFile string
	// AuditURL enables HTTP audit delivery when it is not empty.
	AuditURL string
	// EnableHTTPS switches the web server to TLS.
	EnableHTTPS bool
	// TLSCertFile is a path to the TLS public certificate file.
	TLSCertFile string
	// TLSKeyFile is a path to the TLS private key file.
	TLSKeyFile string
	// TrustedSubnet contains CIDR notation of the subnet allowed to access internal endpoints.
	TrustedSubnet string
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
	DataBaseDSN string
	// MaxConns limits the maximum number of PostgreSQL connections.
	MaxConns int32
	// MinConns keeps a minimum number of PostgreSQL connections ready.
	MinConns int32
	// MaxConnLifetime limits how long a PostgreSQL connection can be reused.
	MaxConnLifetime time.Duration
	// MaxConnIdleTime limits how long an idle PostgreSQL connection is kept.
	MaxConnIdleTime time.Duration
	// HealthCheckPeriod controls how often the PostgreSQL pool checks idle connections.
	HealthCheckPeriod time.Duration
}

type stdFlagValue struct {
	flag      *flag.Flag
	flagSet   *flag.FlagSet
	valueType string
	key       string
}

func (v stdFlagValue) HasChanged() bool {
	changed := false
	v.flagSet.Visit(func(f *flag.Flag) {
		if f.Name == v.flag.Name {
			changed = true
		}
	})
	return changed
}

func (v stdFlagValue) Name() string {
	return v.key
}

func (v stdFlagValue) ValueString() string {
	return v.flag.Value.String()
}

func (v stdFlagValue) ValueType() string {
	return v.valueType
}

// InitConfig parses flags and environment variables into Config.
func InitConfig() (*Config, error) {
	flagSet := flag.CommandLine

	flagSet.String("a", defaultServerAddress, "host address server")
	flagSet.String("b", defaultBaseURL, "base url for short url")
	flagSet.String("f", defaultFileStoragePath, "file storage path")
	flagSet.String("d", defaultDatabaseDSN, "dsn connection to db")
	flagSet.String("audit-file", defaultAuditFile, "file audit storage path")
	flagSet.String("audit-url", defaultAuditURL, "audit url")
	configFilePathShort := flagSet.String("c", defaultConfigFilePath, "json config file path")
	configFilePathLong := flagSet.String("config", defaultConfigFilePath, "json config file path")
	flagSet.Bool("s", defaultHTTPS, "enable HTTPS")
	flagSet.String("tls-cert-file", defaultTLSCertFile, "TLS public certificate file path")
	flagSet.String("tls-key-file", defaultTLSKeyFile, "TLS private key file path")
	flagSet.String("t", defaultTrustedSubnet, "trusted subnet in CIDR notation")

	flag.Parse()

	v := viper.New()
	setDefaults(v)
	bindEnv(v)
	if err := bindFlags(v, flagSet); err != nil {
		return nil, err
	}

	configFilePath := configFilePath(v, flagSet, *configFilePathShort, *configFilePathLong)
	if configFilePath != defaultConfigFilePath {
		v.SetConfigFile(configFilePath)
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	return buildConfig(v), nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault(keyAppEnv, "development")
	v.SetDefault(keyServerAddress, defaultServerAddress)
	v.SetDefault(keyBaseURL, defaultBaseURL)
	v.SetDefault(keyFileStoragePath, defaultFileStoragePath)
	v.SetDefault(keyStorageType, "")
	v.SetDefault(keyAuditFile, defaultAuditFile)
	v.SetDefault(keyAuditURL, defaultAuditURL)
	v.SetDefault(keyEnableHTTPS, defaultHTTPS)
	v.SetDefault(keyTLSCertFile, defaultTLSCertFile)
	v.SetDefault(keyTLSKeyFile, defaultTLSKeyFile)
	v.SetDefault(keyTrustedSubnet, defaultTrustedSubnet)
	v.SetDefault(keyConfigFilePath, defaultConfigFilePath)
	v.SetDefault(keyDatabaseDSN, defaultDatabaseDSN)
	v.SetDefault(keyMaxConns, defaultMaxConns)
	v.SetDefault(keyMinConns, defaultMinConns)
	v.SetDefault(keyMaxConnLifetime, defaultMaxConnLifetime)
	v.SetDefault(keyMaxConnIdleTime, defaultMaxConnIdleTime)
	v.SetDefault(keyHealthCheckPeriod, defaultHealthCheckDelay)
}

func bindEnv(v *viper.Viper) {
	envBindings := map[string]string{
		keyAppEnv:            "APP_ENV",
		keyServerAddress:     "SERVER_ADDRESS",
		keyBaseURL:           "BASE_URL",
		keyFileStoragePath:   "FILE_STORAGE_PATH",
		keyStorageType:       "STORAGE_TYPE",
		keyAuditFile:         "AUDIT_FILE",
		keyAuditURL:          "AUDIT_URL",
		keyEnableHTTPS:       "ENABLE_HTTPS",
		keyTLSCertFile:       "TLS_CERT_FILE",
		keyTLSKeyFile:        "TLS_KEY_FILE",
		keyTrustedSubnet:     "TRUSTED_SUBNET",
		keyConfigFilePath:    "CONFIG",
		keyDatabaseDSN:       "DATABASE_DSN",
		keyMaxConns:          "MAX_CONNS",
		keyMinConns:          "MIN_CONNS",
		keyMaxConnLifetime:   "MAX_CONN_LIFETIME",
		keyMaxConnIdleTime:   "MAX_CONN_IDLE_TIME",
		keyHealthCheckPeriod: "HEALTH_CHECK_PERIOD",
	}
	for key, envName := range envBindings {
		_ = v.BindEnv(key, envName)
	}
}

func bindFlags(v *viper.Viper, flagSet *flag.FlagSet) error {
	flagBindings := []struct {
		flagName  string
		key       string
		valueType string
	}{
		{flagName: "a", key: keyServerAddress, valueType: "string"},
		{flagName: "b", key: keyBaseURL, valueType: "string"},
		{flagName: "f", key: keyFileStoragePath, valueType: "string"},
		{flagName: "d", key: keyDatabaseDSN, valueType: "string"},
		{flagName: "audit-file", key: keyAuditFile, valueType: "string"},
		{flagName: "audit-url", key: keyAuditURL, valueType: "string"},
		{flagName: "s", key: keyEnableHTTPS, valueType: "bool"},
		{flagName: "tls-cert-file", key: keyTLSCertFile, valueType: "string"},
		{flagName: "tls-key-file", key: keyTLSKeyFile, valueType: "string"},
		{flagName: "t", key: keyTrustedSubnet, valueType: "string"},
	}
	for _, binding := range flagBindings {
		f := flagSet.Lookup(binding.flagName)
		if f == nil {
			return fmt.Errorf("flag %q is not registered", binding.flagName)
		}
		err := v.BindFlagValue(binding.key, stdFlagValue{
			flag:      f,
			flagSet:   flagSet,
			valueType: binding.valueType,
			key:       binding.key,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func configFilePath(v *viper.Viper, flagSet *flag.FlagSet, shortValue, longValue string) string {
	if flagWasChanged(flagSet, "config") {
		return longValue
	}
	if flagWasChanged(flagSet, "c") {
		return shortValue
	}
	return v.GetString(keyConfigFilePath)
}

func flagWasChanged(flagSet *flag.FlagSet, name string) bool {
	changed := false
	flagSet.Visit(func(f *flag.Flag) {
		if f.Name == name {
			changed = true
		}
	})
	return changed
}

func buildConfig(v *viper.Viper) *Config {
	baseURL := v.GetString(keyBaseURL)
	return &Config{
		AppEnv: v.GetString(keyAppEnv),
		HTTPServer: HTTPServer{
			ServerAddress: v.GetString(keyServerAddress),
			BaseURL:       baseURL,
			StorageType:   v.GetString(keyStorageType),
			Timeout:       defaultTimeout,
			IdleTimeout:   defaultIdleTimeout,
			AuditFile:     v.GetString(keyAuditFile),
			AuditURL:      v.GetString(keyAuditURL),
			EnableHTTPS:   v.GetBool(keyEnableHTTPS),
			TLSCertFile:   v.GetString(keyTLSCertFile),
			TLSKeyFile:    v.GetString(keyTLSKeyFile),
			TrustedSubnet: v.GetString(keyTrustedSubnet),
			Postgres: &PostgresConfig{
				DataBaseDSN:       v.GetString(keyDatabaseDSN),
				MaxConns:          int32(v.GetInt(keyMaxConns)),
				MinConns:          int32(v.GetInt(keyMinConns)),
				MaxConnLifetime:   v.GetDuration(keyMaxConnLifetime),
				MaxConnIdleTime:   v.GetDuration(keyMaxConnIdleTime),
				HealthCheckPeriod: v.GetDuration(keyHealthCheckPeriod),
			},
		},
		BaseAddressShort: baseURL,
		FileStoragePath:  v.GetString(keyFileStoragePath),
		SecretKey:        SecretKey,
	}
}
