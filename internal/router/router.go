package router

import (
	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/handler/api"
	"github.com/Luclpor/url_shortener.git/internal/logger"
	"github.com/Luclpor/url_shortener.git/internal/repository"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(logger.RequestLogger)
	r.Use(middleware.Recoverer)

	repo := repository.NewRepository()
	manager := service.NewURLManager(repo)

	r.Post("/", api.NewCreateHandler(cfg, manager))
	r.Get("/{id}", api.NewGetterHandler(manager))

	return r
}
