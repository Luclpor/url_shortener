package main

import (
	"log"

	"github.com/Luclpor/url_shortener.git/internal/router/server"
)

func main() {
	s, err := server.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
