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
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL       string `env:"BASE_URL" default:"http://localhost:9090"`
	Timeout       time.Duration
	IdleTimeout   time.Duration
}

func InitConfig() *Config {
	h := flag.String("a", "localhost:7999", "host address server")
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
