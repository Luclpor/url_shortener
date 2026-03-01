package api

import (
	"context"
	"io"
	"net/http"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/service"
)

func NewCreateHandler(cfg *config.Config, manager *service.URLManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		su, exs, err := manager.TryCreateShortURL(ctx, string(b))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !exs {
			w.WriteHeader(http.StatusOK)
		}
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(cfg.BaseAddressShort + "/" + su))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func NewGetterHandler(manager *service.URLManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		s, err := manager.GetURL(ctx, r.PathValue("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Add("Location", s.FullURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
