package api

import (
	"context"
	"net/http"
	"time"
)

// NewPingHandler returns a handler for GET /ping health checks.
func (h *Handler) NewPingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.health.Ping(ctx); err != nil {
			http.Error(w, "database is unavailable", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
