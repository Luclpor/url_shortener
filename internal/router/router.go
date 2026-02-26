package router

import (
	"net/http"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/handler"
)

type Server struct {
	Host string
}

func RunServer() {
	cfg := config.InitConfig()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", handler.CreatedShortURL)
	mux.HandleFunc("GET /{id}", handler.GetShortURL)
	err := http.ListenAndServe(cfg.Host, mux)
	if err != nil {
		panic(err)
	}
}
