package main

import (
	"github.com/Luclpor/url_shortener.git/internal/router/server"
)

func main() {
	s := server.NewServer()
	s.Start()
}
