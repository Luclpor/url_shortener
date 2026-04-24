package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/config/db"
	"github.com/Luclpor/url_shortener.git/internal/logger"
	router2 "github.com/Luclpor/url_shortener.git/internal/router"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/Luclpor/url_shortener.git/internal/service/auth"
	"github.com/Luclpor/url_shortener.git/internal/storage"
	"github.com/Luclpor/url_shortener.git/internal/storage/inmemory"
	"github.com/Luclpor/url_shortener.git/internal/storage/postgres"
	"go.uber.org/zap"
)

const (
	shutdownTimeout = 10 * time.Second
)

type Server struct {
	httpServer *http.Server
	closers    []func() error
	appLogger  *zap.Logger
}

func NewServer() (*Server, error) {
	cfg, err := config.InitConfig()
	if err != nil {
		return nil, err
	}
	appLogger, err := logger.InitLogger(cfg.AppEnv)
	if err != nil {
		return nil, err
	}
	var repo service.URLRepository
	var healthChecker service.HealthChecker
	var userAuth storage.UserAuthentication
	var closers []func() error
	if cfg.Postgres.DataBaseDSN != "" {
		pool, err := postgres.NewPool(context.Background(), cfg.Postgres, appLogger)
		if err != nil {
			return nil, err
		}
		urlRepo := postgres.NewURLRepository(pool)
		repo = urlRepo
		healthChecker = postgres.NewHealthRepository(pool)
		userAuth, err = auth.InitAuthService([]byte(cfg.SecretKey), urlRepo)
		if err != nil {
			appLogger.Error("Could not initialize auth service", zap.Error(err))
			return nil, err
		}
		err = db.RunMigrations(cfg.Postgres.DataBaseDSN)
		if err != nil {
			appLogger.Error("Could not run migrations", zap.Error(err))
			return nil, err
		}
		closers = append(closers, func() error {
			pool.Close()
			return nil
		})
	} else {
		fileStorage, err := inmemory.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			appLogger.Error("Could not initialize file storage", zap.Error(err))
			return nil, err
		}
		memRepo, err := inmemory.NewRepository(fileStorage)
		if err != nil {
			appLogger.Error("Could not initialize in memory repository", zap.Error(err))
			return nil, err
		}
		healthChecker = inmemory.NewHealthRepository()
		userAuth, err = auth.InitAuthService([]byte(cfg.SecretKey), memRepo)
		if err != nil {
			appLogger.Error("Could not initialize auth service", zap.Error(err))
			return nil, err
		}
		repo = memRepo
		closers = append(closers, memRepo.Close)
	}
	healthService := service.NewHealthService(healthChecker)
	manager := service.NewURLManager(repo, appLogger)
	router, err := router2.NewRouter(cfg, manager, healthService, userAuth, appLogger)
	if err != nil {
		appLogger.Error("Could not initialize router", zap.Error(err))
		return nil, err
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
		appLogger,
	}

	return server, nil
}

func (s *Server) Start() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	go func() {
		s.appLogger.Info("Server listening on",
			zap.String("server_address", s.httpServer.Addr),
		)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.appLogger.Fatal("Error starting server", zap.Error(err))
		}
	}()

	<-quit
	s.appLogger.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.appLogger.Error("server forced to shutdown", zap.Error(err))
		return err
	}
	for _, closeFn := range s.closers {
		if err := closeFn(); err != nil {
			s.appLogger.Error("close function failed", zap.Error(err))
			return err
		}
	}
	s.appLogger.Info("Server exited properly")
	return nil
}
