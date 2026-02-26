package handler

import (
	"io"
	"net/http"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/service"
)

func CreatedShortURL(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	su, exs := service.TryCreateShortURL(string(b))
	if !exs {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	_, err = w.Write([]byte(config.GlobalConfig.Host + "/" + su))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetShortURL(w http.ResponseWriter, r *http.Request) {
	s := service.GetShortURL(r.PathValue("id"))
	if s == "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	w.Header().Add("Location", s)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
