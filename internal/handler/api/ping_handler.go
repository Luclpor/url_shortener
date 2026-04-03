package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Luclpor/url_shortener.git/internal/service"
)

func NewPingHandler(health *service.HealthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := health.Ping(ctx); err != nil {
			http.Error(w, "database is unavailable", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}
