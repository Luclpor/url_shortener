package router

import (
	"time"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/handler/api"
	customMidlleware "github.com/Luclpor/url_shortener.git/internal/handler/middleware"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/Luclpor/url_shortener.git/internal/service/audit"
	"github.com/Luclpor/url_shortener.git/internal/service/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// NewRouter wires application middleware and URL shortener routes.
func NewRouter(cfg *config.Config, manager *service.URLManager, healthService *service.HealthService, userAuth auth.UserAuthentication, eventPublisher *audit.Event, appLogger *zap.Logger) (*chi.Mux, error) {
	r := chi.NewRouter()
	handler := api.NewHandler(cfg, healthService, manager, userAuth, eventPublisher, appLogger)
	r.Use(middleware.RequestID)
	r.Use(customMidlleware.RequestLogger(appLogger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(777 * time.Second))
	r.Use(customMidlleware.CompressMiddleware)
	r.Get("/ping", handler.NewPingHandler())

	r.Route("/", func(r chi.Router) {
		r.Use(customMidlleware.Auth(userAuth, appLogger))
		r.Post("/api/shorten", handler.NewCreateShortenUlrJSONHandler())
		r.Get("/api/user/urls", handler.NewGetBatchHandler())
		r.Delete("/api/user/urls", handler.NewDeleteBatchHandler())
		r.Post("/", handler.NewCreateHandler())
		r.Post("/api/shorten/batch", handler.NewCreateBatchShortenHandler())
		r.Get("/{id}", handler.NewGetterHandler())
	})

	return r, nil
}
