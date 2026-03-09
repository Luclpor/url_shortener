package config

import (
	"flag"
	"time"
)

type Config struct {
	HTTPServer
	BaseAddressShort string
}

type HTTPServer struct {
	Host        string
	Timeout     time.Duration
	IdleTimeout time.Duration
}

func InitConfig() *Config {
	h := flag.String("a", "localhost:8080", "host address server")
	b := flag.String("b", "http://localhost:8080", "base url for short url")

	flag.Parse()

	cfg := &Config{
		HTTPServer: HTTPServer{
			Host:        *h,
			Timeout:     time.Second * 4,
			IdleTimeout: time.Second * 30,
		},
		BaseAddressShort: *b,
	}
	return cfg
}
