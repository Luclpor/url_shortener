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

func NewRouter(cfg *config.Config, manager *service.URLManager, healthService *service.HealthService) (*chi.Mux, error) {
	r := chi.NewRouter()
	handler := api.NewHandler(cfg, healthService, manager)
	r.Use(middleware.RequestID)
	r.Use(logger.RequestLogger)
	r.Use(middleware.Recoverer)
	r.Use(customMidlleware.CompressMiddleware)

	r.Post("/api/shorten", handler.NewCreateShortenUlrJSONHandler())
	r.Post("/", handler.NewCreateHandler())
	r.Post("/api/shorten/batch", handler.NewCreateBatchShortenHandler())
	r.Get("/{id}", handler.NewGetterHandler())
	r.Get("/ping", handler.NewPingHandler())
	return r, nil
}
