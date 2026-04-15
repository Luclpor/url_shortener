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
	"github.com/Luclpor/url_shortener.git/internal/config/db"
	router2 "github.com/Luclpor/url_shortener.git/internal/router"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/Luclpor/url_shortener.git/internal/storage/inmemory"
	"github.com/Luclpor/url_shortener.git/internal/storage/postgres"
)

const (
	shutdownTimeout = 10 * time.Second
)

type Server struct {
	httpServer *http.Server
	closers    []func() error
}

func NewServer() *Server {
	cfg := config.InitConfig()

	var repo service.URLRepository
	var healthChecker service.HealthChecker
	var closers []func() error
	if cfg.Postgres.DataBaseDSN != "" {
		pool, err := postgres.NewPool(context.Background(), cfg.Postgres)
		if err != nil {
			log.Fatal(err)
		}
		repo = postgres.NewURLRepository(pool)
		healthChecker = postgres.NewHealthRepository(pool)
		err = db.RunMigrations(cfg.Postgres.DataBaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		closers = append(closers, func() error {
			pool.Close()
			return nil
		})
	} else {
		fileStorage, err := inmemory.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		memRepo, err := inmemory.NewRepository(fileStorage)
		healthChecker = inmemory.NewHealthRepository()
		if err != nil {
			log.Fatal(err)
		}
		repo = memRepo
		closers = append(closers, memRepo.Close)
	}
	healthService := service.NewHealthService(healthChecker)
	manager := service.NewURLManager(repo)
	router, err := router2.NewRouter(cfg, manager, healthService)
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
		closers,
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
	for _, closeFn := range s.closers {
		if err := closeFn(); err != nil {
			log.Printf("close error: %v", err)
		}
	}
	log.Println("Server exited properly")
}
