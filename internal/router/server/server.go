package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/config/db"
	"github.com/Luclpor/url_shortener.git/internal/logger"
	router2 "github.com/Luclpor/url_shortener.git/internal/router"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/Luclpor/url_shortener.git/internal/service/audit"
	"github.com/Luclpor/url_shortener.git/internal/service/auth"
	"github.com/Luclpor/url_shortener.git/internal/storage/inmemory"
	"github.com/Luclpor/url_shortener.git/internal/storage/postgres"
	"go.uber.org/zap"
)

// Server owns the configured HTTP server and shutdown resources.
type Server struct {
	httpServer *http.Server
	closers    []func() error
	appLogger  *zap.Logger
	useHTTPS   bool
}

// NewServer creates a fully configured URL shortener server.
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
	var userAuth auth.UserAuthentication
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
	storageAuditSbcr, closer, err := audit.NewStorageAuditor(cfg.AuditFile)
	if err != nil {
		return nil, err
	}
	if closer != nil {
		closers = append(closers, closer)
	}
	extAuditSbcr, err := audit.NewRetryableHTTPClient(cfg.AuditURL)
	if err != nil {
		return nil, err
	}
	eventAuditPublisher := audit.NewEvent(appLogger)
	eventAuditPublisher.Register(storageAuditSbcr)
	eventAuditPublisher.Register(extAuditSbcr)
	healthService := service.NewHealthService(healthChecker)
	manager := service.NewURLManager(repo, appLogger)
	router, err := router2.NewRouter(cfg, manager, healthService, userAuth, eventAuditPublisher, appLogger)
	if err != nil {
		appLogger.Error("Could not initialize router", zap.Error(err))
		return nil, err
	}

	server := &Server{
		httpServer: &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      router,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		closers:   closers,
		appLogger: appLogger,
		useHTTPS:  cfg.EnableHTTPS,
	}

	return server, nil
}

// Start runs the HTTP server until an interrupt or termination signal is received.
func (s *Server) Start() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(quit)

	go func() {
		s.appLogger.Info("Server listening on",
			zap.String("server_address", s.httpServer.Addr),
			zap.Bool("https_enabled", s.useHTTPS),
		)
		if err := s.listenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.appLogger.Fatal("Error starting server", zap.Error(err))
		}
	}()

	sig := <-quit
	s.appLogger.Info("Server shutting down...", zap.String("signal", sig.String()))

	shutdownErr := s.httpServer.Shutdown(context.Background())
	if shutdownErr != nil {
		s.appLogger.Error("server forced to shutdown", zap.Error(shutdownErr))
	}
	closeErr := s.closeResources()
	if closeErr != nil {
		s.appLogger.Error("server resources close failed", zap.Error(closeErr))
	}
	if err := errors.Join(shutdownErr, closeErr); err != nil {
		return err
	}
	s.appLogger.Info("Server exited properly")
	return nil
}

func (s *Server) closeResources() error {
	var closeErr error
	for _, closeFn := range s.closers {
		if err := closeFn(); err != nil {
			closeErr = errors.Join(closeErr, err)
		}
	}
	return closeErr
}

func (s *Server) listenAndServe() error {
	if !s.useHTTPS {
		return s.httpServer.ListenAndServe()
	}

	certificate, err := newSelfSignedCertificate(s.httpServer.Addr)
	if err != nil {
		return err
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
	}
	listener, err := tls.Listen("tcp", s.httpServer.Addr, tlsConfig)
	if err != nil {
		return err
	}
	return s.httpServer.Serve(listener)
}
