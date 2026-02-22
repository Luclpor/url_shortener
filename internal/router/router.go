package router

import (
	"net/http"

	"github.com/Luclpor/url_shortener.git/internal/handler"
)

func RunServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", handler.CreatedShortUrl)
	mux.HandleFunc("GET /{id}", handler.GetShortUrl)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
