package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/repository"
	router2 "github.com/Luclpor/url_shortener.git/internal/router"
	"github.com/Luclpor/url_shortener.git/internal/service"
)

const (
	shutdownTimeout = 10 * time.Second
)

type Server struct {
	httpServer *http.Server
}

func NewServer() *Server {
	cfg := config.InitConfig()
	repo, err := repository.NewRepository(cfg.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	manager := service.NewURLManager(repo)
	router, err := router2.NewRouter(cfg, manager)
	if err != nil {
		log.Fatal(err)
	}

	server := &Server{
		&http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      router,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}

	return server
}

func (s *Server) Start() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	go func() {
		log.Printf("Server listening on %s\n", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
