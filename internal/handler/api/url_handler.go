package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/go-chi/render"
)

func NewCreateShortenUlrJSONHandler(cfg *config.Config, manager *service.URLManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		var model *api.CreateShortenReq
		err := json.NewDecoder(r.Body).Decode(&model)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err)
		}
		w.Header().Set("Content-Type", "application/json")
		responseModel, exs, err := manager.TryCreateShortURL(ctx, model.URL)
		responseModel.ShortenURL = cfg.BaseAddressShort + "/" + responseModel.ShortenURL
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, err)
		}
		if !exs {
			render.Status(r, http.StatusOK)
		}
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, responseModel)
	}
}

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
		_, err = w.Write([]byte(cfg.BaseAddressShort + "/" + su.ShortenURL))
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
