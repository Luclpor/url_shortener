package handler

import (
	"io"
	"net/http"

	"github.com/Luclpor/url_shortener.git/internal/service"
)

type Url struct {
	names map[string]string
}

var urls *Url

func CreatedShortUrl(w http.ResponseWriter, r *http.Request) {
	if urls == nil {
		urls = &Url{make(map[string]string)}
	}
	if r.Method == http.MethodPost {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		key := service.GenerateRandomString(5)
		urls.names[key] = string(b)
		resp := "http://" + r.Host + "/" + key

		w.Write([]byte(resp))
	}
}

func GetShortUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		//b, err := io.ReadAll(r.Body)
		//if err != nil {
		//	w.WriteHeader(http.StatusBadRequest)
		//	return
		//}
		//v := r.URL.Query()
		//_ = v
		s := urls.names[r.PathValue("id")]
		//w.Header().Set("Location", s)
		w.Header().Add("Location", s)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
