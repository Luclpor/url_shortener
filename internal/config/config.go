package config

type Config struct {
	HTTPServer
}

type HTTPServer struct {
	Host string
}

func InitConfig() *Config {
	cfg := &Config{
		HTTPServer: HTTPServer{
			Host: "localhost:8080",
		},
	}
	return cfg
}
