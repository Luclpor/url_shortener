package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/Luclpor/url_shortener.git/internal/storage"
	appErrors "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

type Handler struct {
	cfg         *config.Config
	manager     *service.URLManager
	authService storage.UserAuthentication
	health      *service.HealthService
	logger      *zap.Logger
}

func NewHandler(cfg *config.Config, health *service.HealthService, manager *service.URLManager, userAuth storage.UserAuthentication, appLogger *zap.Logger) *Handler {
	return &Handler{
		cfg:         cfg,
		health:      health,
		authService: userAuth,
		manager:     manager,
		logger:      appLogger,
	}
}

func (h *Handler) NewCreateBatchShortenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var model []api.CreateShortenReq
		err := json.NewDecoder(r.Body).Decode(&model)
		if err != nil {
			h.logger.Error("Failed to decode create shorten request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		user, err := h.authService.GetUserFromContext(r.Context())
		if err != nil {
			h.logger.Error("failed to get user from context:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		responseModel, err := h.manager.CreateBatchURL(r.Context(), model, user)
		if err != nil {
			h.logger.Error("failed to create shorten request", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(responseModel) == 0 {
			w.WriteHeader(http.StatusOK)
			return
		}
		render.Status(r, http.StatusCreated)
		for i := range responseModel {
			joined, err := url.JoinPath(h.cfg.BaseAddressShort, responseModel[i].ShortURL)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				h.logger.Error("failed to join short url:", zap.Error(err))
				return
			}
			responseModel[i].ShortURL = joined
		}
		render.JSON(w, r, responseModel)
	}
}

func (h *Handler) NewCreateShortenUlrJSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var model api.CreateShortenReq
		err := json.NewDecoder(r.Body).Decode(&model)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err)
			return
		}
		user, err := h.authService.GetUserFromContext(r.Context())
		if err != nil {
			h.logger.Error("failed to get user from context:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		responseModel, err := h.manager.CreateShortURL(r.Context(), model.URL, user)
		if err != nil && !errors.Is(err, appErrors.ErrAlreadyExists) {
			h.logger.Error("failed to create short url:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err != nil && errors.Is(err, appErrors.ErrAlreadyExists) {
			render.Status(r, http.StatusConflict)
		} else {
			render.Status(r, http.StatusCreated)
		}
		joined, err := url.JoinPath(h.cfg.BaseAddressShort, responseModel.Result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error("failed to join short url:", zap.Error(err))
			return
		}
		responseModel.Result = joined
		render.JSON(w, r, responseModel)
	}
}

func (h *Handler) NewCreateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.PlainText(w, r, err.Error())
			return
		}
		user, err := h.authService.GetUserFromContext(r.Context())
		if err != nil {
			h.logger.Error("failed to get user from context:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		su, createErr := h.manager.CreateShortURL(r.Context(), string(body), user)
		if createErr != nil && !errors.Is(createErr, appErrors.ErrAlreadyExists) {
			h.logger.Error("failed to create short url:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if errors.Is(createErr, appErrors.ErrAlreadyExists) {
			render.Status(r, http.StatusConflict)
		} else {
			render.Status(r, http.StatusCreated)
		}
		joined, err := url.JoinPath(h.cfg.BaseAddressShort, su.Result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error("failed to join short url:", zap.Error(err))
			return
		}
		render.PlainText(w, r, joined)
	}
}

func (h *Handler) NewGetterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := h.manager.GetURL(r.Context(), r.PathValue("id"))
		if err != nil {
			if errors.Is(err, appErrors.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if errors.Is(err, appErrors.ErrURLWasDeleted) {
				w.WriteHeader(http.StatusGone)
				return
			}
			render.Status(r, http.StatusBadRequest)
			render.PlainText(w, r, err.Error())
			return
		}
		w.Header().Add("Location", s.OriginalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (h *Handler) NewDeleteBatchHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := h.authService.GetUserFromContext(r.Context())
		if err != nil {
			h.logger.Error("failed to get user from context:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var model []string
		err = json.NewDecoder(r.Body).Decode(&model)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, err)
			return
		}
		err = h.manager.DeleteBatch(r.Context(), model, user)
		if err != nil {
			h.logger.Error("failed to delete short urls:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		return
	}
}

func (h *Handler) NewGetBatchHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := h.authService.GetUserFromContext(r.Context())
		if err != nil {
			h.logger.Error("failed to get user from context:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		urls, err := h.manager.GetBatchURLByUserID(r.Context(), user)
		if err != nil {
			if errors.Is(err, appErrors.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			h.logger.Error("failed to get short urls:", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for i, u := range urls {
			joined, err := url.JoinPath(h.cfg.BaseAddressShort, u.ShortURL)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				h.logger.Error("failed to join short url:", zap.Error(err))
				return
			}
			urls[i].ShortURL = joined
		}
		render.JSON(w, r, urls)
	}
}
