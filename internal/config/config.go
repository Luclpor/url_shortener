package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	HTTPServer
	BaseAddressShort string
}

type HTTPServer struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
	Timeout       time.Duration
	IdleTimeout   time.Duration
}

func InitConfig() *Config {
	h := flag.String("a", "localhost:8080", "host address server")
	b := flag.String("b", "http://localhost:8080", "base url for short url")

	flag.Parse()

	cfg := Config{
		HTTPServer: HTTPServer{
			ServerAddress: *h,
			Timeout:       time.Second * 4,
			IdleTimeout:   time.Second * 30,
		},
		BaseAddressShort: *b,
	}

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg
}
