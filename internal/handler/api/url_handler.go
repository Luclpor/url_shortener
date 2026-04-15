package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/service"
	appErrors "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/go-chi/render"
)

type Handler struct {
	cfg     *config.Config
	manager *service.URLManager
	health  *service.HealthService
}

func NewHandler(cfg *config.Config, health *service.HealthService, manager *service.URLManager) *Handler {
	return &Handler{
		cfg:     cfg,
		health:  health,
		manager: manager,
	}
}

func (h *Handler) NewCreateBatchShortenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()
		var model []api.CreateShortenReq
		err := json.NewDecoder(r.Body).Decode(&model)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		responseModel, err := h.manager.CreateBatchURL(ctx, model)
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
		for i := range responseModel {
			responseModel[i].ShortURL = h.cfg.BaseAddressShort + "/" + responseModel[i].ShortURL
		}
		render.JSON(w, r, responseModel)
	}
}

func (h *Handler) NewCreateShortenUlrJSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()
		var model api.CreateShortenReq
		err := json.NewDecoder(r.Body).Decode(&model)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		responseModel, err := h.manager.CreateShortURL(ctx, model.URL)
		if err != nil && !errors.Is(err, appErrors.ErrAlreadyExists) {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, err)
			return
		}
		responseModel.Result = h.cfg.BaseAddressShort + "/" + responseModel.Result
		if err != nil && errors.Is(err, appErrors.ErrAlreadyExists) {
			render.Status(r, http.StatusConflict)
		} else {
			render.Status(r, http.StatusCreated)
		}
		render.JSON(w, r, responseModel)
	}
}

func (h *Handler) NewCreateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.PlainText(w, r, http.StatusText(http.StatusBadRequest))
			return
		}
		su, createErr := h.manager.CreateShortURL(ctx, string(body))
		if createErr != nil && !errors.Is(createErr, appErrors.ErrAlreadyExists) {
			render.Status(r, http.StatusInternalServerError)
			render.PlainText(w, r, http.StatusText(http.StatusInternalServerError))
			return
		}

		if errors.Is(createErr, appErrors.ErrAlreadyExists) {
			render.Status(r, http.StatusConflict)
		} else {
			render.Status(r, http.StatusCreated)
		}

		render.PlainText(w, r, h.cfg.BaseAddressShort+"/"+su.Result)
	}
}

func (h *Handler) NewGetterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()
		s, err := h.manager.GetURL(ctx, r.PathValue("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Add("Location", s.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
