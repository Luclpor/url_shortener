package router

import (
	"net/http"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/handler"
	"github.com/Luclpor/url_shortener.git/internal/repository"
	"github.com/Luclpor/url_shortener.git/internal/service"
)

func RunServer() {
	cfg := config.InitConfig()
	repo := repository.NewRepository()
	manger := service.NewURLManager(repo)
	createHandler := handler.NewCreateHandler(cfg, manger)
	getterHandler := handler.NewGetterHandler(manger)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", createHandler)
	mux.HandleFunc("GET /{id}", getterHandler)
	err := http.ListenAndServe(cfg.Host, mux)
	if err != nil {
		panic(err)
	}
}
