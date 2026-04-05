package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/service"
	errors2 "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/go-chi/render"
)

func NewCreateBatchShortenHandler(cfg *config.Config, manager *service.URLManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		var model []api.CreateShortenReq
		err := json.NewDecoder(r.Body).Decode(&model)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		responseModel, err := manager.CreateBatchURL(ctx, model)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, err)
			return
		}
		if len(responseModel) == 0 {
			render.Status(r, http.StatusOK)
			return
		}
		render.Status(r, http.StatusCreated)
		for i, _ := range responseModel {
			responseModel[i].ShortURL = cfg.BaseAddressShort + "/" + responseModel[i].ShortURL
		}
		render.JSON(w, r, responseModel)
	}
}

func NewCreateShortenUlrJSONHandler(cfg *config.Config, manager *service.URLManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		var model api.CreateShortenReq
		err := json.NewDecoder(r.Body).Decode(&model)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		responseModel, err := manager.CreateShortURL(ctx, model.URL)
		if err != nil && !errors.Is(err, errors2.ErrAlreadyExists) {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, err)
			return
		}
		responseModel.Result = cfg.BaseAddressShort + "/" + responseModel.Result
		if err != nil && errors.Is(err, errors2.ErrAlreadyExists) {
			render.Status(r, http.StatusOK)
		} else {
			render.Status(r, http.StatusCreated)
		}
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
		su, err := manager.CreateShortURL(ctx, string(b))
		if err != nil && !errors.Is(err, errors2.ErrAlreadyExists) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err != nil && errors.Is(err, errors2.ErrAlreadyExists) {
			w.WriteHeader(http.StatusOK)
		}
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(cfg.BaseAddressShort + "/" + su.Result))
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
		w.Header().Add("Location", s.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
