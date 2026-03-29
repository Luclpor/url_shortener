package router

import (
	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/handler/api"
	customMidlleware "github.com/Luclpor/url_shortener.git/internal/handler/middleware"
	"github.com/Luclpor/url_shortener.git/internal/logger"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg *config.Config, manager *service.URLManager) (*chi.Mux, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(logger.RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(customMidlleware.CompressMiddleware)

	r.Post("/api/shorten", api.NewCreateShortenUlrJSONHandler(cfg, manager))
	r.Post("/", api.NewCreateHandler(cfg, manager))
	r.Get("/{id}", api.NewGetterHandler(manager))

	return r, nil
}
