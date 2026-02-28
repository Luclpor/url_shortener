package config

import "time"

type Config struct {
	HTTPServer
}

type HTTPServer struct {
	Host        string
	Timeout     time.Duration
	IdleTimeout time.Duration
}

func InitConfig() *Config {
	cfg := &Config{
		HTTPServer: HTTPServer{
			Host:        "localhost:8080",
			Timeout:     time.Second * 4,
			IdleTimeout: time.Second * 30,
		},
	}
	return cfg
}
