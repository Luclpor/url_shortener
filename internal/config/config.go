package config

import "time"

type Config struct {
	HTTPServer
}

type HTTPServer struct {
	Host        string
	Port        string
	Timeout     time.Duration
	IdleTimeout time.Duration
}

func InitConfig() *Config {
	cfg := &Config{
		HTTPServer: HTTPServer{
			Host: "localhost:8080",
		},
	}
	return cfg
}
