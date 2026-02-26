package config

type Config struct {
	HTTPServer
}

type HTTPServer struct {
	Host string
}

var GlobalConfig *Config

func InitConfig() *Config {
	cfg := &Config{
		HTTPServer: HTTPServer{
			Host: "localhost:8080",
		},
	}
	GlobalConfig = cfg
	return cfg
}
