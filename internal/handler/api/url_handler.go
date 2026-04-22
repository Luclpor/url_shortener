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
	"github.com/Luclpor/url_shortener.git/internal/service/auth"
	appErrors "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/go-chi/render"
)

type Handler struct {
	cfg         *config.Config
	manager     *service.URLManager
	authService auth.UserAuthentication
	health      *service.HealthService
}

func NewHandler(cfg *config.Config, health *service.HealthService, manager *service.URLManager, userAuth auth.UserAuthentication) *Handler {
	return &Handler{
		cfg:         cfg,
		health:      health,
		authService: userAuth,
		manager:     manager,
	}
}

func (h *Handler) NewCreateBatchShortenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*120)
		defer cancel()
		var model []api.CreateShortenReq
		err := json.NewDecoder(r.Body).Decode(&model)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err)
			return
		}
		user, err := h.authService.GetUserFromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		responseModel, err := h.manager.CreateBatchURL(ctx, model, user)
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
		user, err := h.authService.GetUserFromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		responseModel, err := h.manager.CreateShortURL(ctx, model.URL, user)
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
		user, err := h.authService.GetUserFromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		su, createErr := h.manager.CreateShortURL(ctx, string(body), user)
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
			if errors.Is(err, appErrors.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if errors.Is(err, appErrors.ErrURLWasDeleted) {
				w.WriteHeader(http.StatusGone)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Add("Location", s.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (h *Handler) NewDeleteBatchHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()
		user, err := h.authService.GetUserFromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var model api.URLDeleteBatchAPIModel
		err = json.NewDecoder(r.Body).Decode(&model)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err)
			return
		}
		err = h.manager.DeleteBatch(ctx, model, user)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			return
		}
		render.Status(r, http.StatusAccepted)
		return
	}
}

func (h *Handler) NewGetBatchHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*1330)
		defer cancel()
		user, err := h.authService.GetUserFromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		urls, err := h.manager.GetBatchURLByUserID(ctx, user)
		if err != nil {
			if errors.Is(err, appErrors.ErrNotFound) {
				render.Status(r, http.StatusNoContent)
				render.JSON(w, r, err)
				return
			}
			render.Status(r, http.StatusInternalServerError)
			return
		}
		for i, url := range urls {
			urls[i].ShortURL = h.cfg.BaseAddressShort + "/" + url.ShortURL
		}
		render.JSON(w, r, urls)
	}
}
