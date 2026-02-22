package router

import (
	"net/http"

	"github.com/Luclpor/url_shortener.git/internal/handler"
)

func RunServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", handler.CreatedShortURL)
	mux.HandleFunc("GET /{id}", handler.GetShortURL)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
