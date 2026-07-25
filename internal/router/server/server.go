package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/config/db"
	"github.com/Luclpor/url_shortener.git/internal/grpc/pb"
	"github.com/Luclpor/url_shortener.git/internal/handler/grpcapi"
	"github.com/Luclpor/url_shortener.git/internal/logger"
	router2 "github.com/Luclpor/url_shortener.git/internal/router"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/Luclpor/url_shortener.git/internal/service/audit"
	"github.com/Luclpor/url_shortener.git/internal/service/auth"
	"github.com/Luclpor/url_shortener.git/internal/storage/inmemory"
	"github.com/Luclpor/url_shortener.git/internal/storage/postgres"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const shutdownTimeout = 10 * time.Second

// Server owns the configured HTTP server and shutdown resources.
type Server struct {
	httpServer  *http.Server
	grpcServer  *grpc.Server
	grpcAddress string
	closers     []func() error
	appLogger   *zap.Logger
	useHTTPS    bool
	tlsCertFile string
	tlsKeyFile  string
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
		grpcAddress: cfg.GRPCServerAddress,
		closers:     closers,
		appLogger:   appLogger,
		useHTTPS:    cfg.EnableHTTPS,
		tlsCertFile: cfg.TLSCertFile,
		tlsKeyFile:  cfg.TLSKeyFile,
	}
	grpcServer, err := server.newGRPCServer(cfg, manager, userAuth, eventAuditPublisher)
	if err != nil {
		return nil, err
	}
	server.grpcServer = grpcServer

	return server, nil
}

// Start runs the HTTP server until an interrupt or termination signal is received.
func (s *Server) Start() error {
	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	serverErrCh := make(chan error, 2)
	go func() {
		s.appLogger.Info("Server listening on",
			zap.String("server_address", s.httpServer.Addr),
			zap.Bool("https_enabled", s.useHTTPS),
		)
		if err := s.listenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()
	go func() {
		s.appLogger.Info("gRPC server listening on",
			zap.String("grpc_server_address", s.grpcAddress),
			zap.Bool("tls_enabled", s.useHTTPS),
		)
		if err := s.listenAndServeGRPC(); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serverErrCh <- err
		}
	}()

	select {
	case <-shutdownCtx.Done():
		stop()
		s.appLogger.Info("Server shutting down...")

		shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		shutdownErr := s.httpServer.Shutdown(shutdownTimeoutCtx)
		if shutdownErr != nil {
			s.appLogger.Error("server forced to shutdown", zap.Error(shutdownErr))
		}
		grpcShutdownErr := s.shutdownGRPC(shutdownTimeoutCtx)
		if grpcShutdownErr != nil {
			s.appLogger.Error("gRPC server forced to shutdown", zap.Error(grpcShutdownErr))
		}
		closeErr := s.closeResources()
		if closeErr != nil {
			s.appLogger.Error("server resources close failed", zap.Error(closeErr))
		}
		if err := errors.Join(shutdownErr, grpcShutdownErr, closeErr); err != nil {
			return err
		}
	case serverErr := <-serverErrCh:
		s.appLogger.Error("Error starting server", zap.Error(serverErr))

		shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		shutdownErr := s.httpServer.Shutdown(shutdownTimeoutCtx)
		if shutdownErr != nil && !errors.Is(shutdownErr, http.ErrServerClosed) {
			s.appLogger.Error("server forced to shutdown", zap.Error(shutdownErr))
		}
		grpcShutdownErr := s.shutdownGRPC(shutdownTimeoutCtx)
		if grpcShutdownErr != nil {
			s.appLogger.Error("gRPC server forced to shutdown", zap.Error(grpcShutdownErr))
		}
		closeErr := s.closeResources()
		if closeErr != nil {
			s.appLogger.Error("server resources close failed", zap.Error(closeErr))
		}
		if err := errors.Join(serverErr, shutdownErr, grpcShutdownErr, closeErr); err != nil {
			return err
		}
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

	certificate, err := s.tlsCertificateFor(s.httpServer.Addr)
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

func (s *Server) newGRPCServer(cfg *config.Config, manager *service.URLManager, userAuth auth.UserAuthentication, eventPublisher *audit.Event) (*grpc.Server, error) {
	options := make([]grpc.ServerOption, 0)
	if s.useHTTPS {
		certificate, err := s.tlsCertificateFor(s.grpcAddress)
		if err != nil {
			return nil, err
		}
		options = append(options, grpc.Creds(credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{certificate},
			MinVersion:   tls.VersionTLS12,
		})))
	}

	grpcServer := grpc.NewServer(options...)
	pb.RegisterShortenerServiceServer(grpcServer, grpcapi.NewShortenerServer(cfg, manager, userAuth, eventPublisher, s.appLogger))
	return grpcServer, nil
}

func (s *Server) listenAndServeGRPC() error {
	listener, err := net.Listen("tcp", s.grpcAddress)
	if err != nil {
		return err
	}
	return s.grpcServer.Serve(listener)
}

func (s *Server) shutdownGRPC(ctx context.Context) error {
	if s.grpcServer == nil {
		return nil
	}

	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		s.grpcServer.Stop()
		<-stopped
		return ctx.Err()
	}
}

func (s *Server) tlsCertificate() (tls.Certificate, error) {
	address := ""
	if s.httpServer != nil {
		address = s.httpServer.Addr
	}
	return s.tlsCertificateFor(address)
}

func (s *Server) tlsCertificateFor(address string) (tls.Certificate, error) {
	if (s.tlsCertFile == "") != (s.tlsKeyFile == "") {
		return tls.Certificate{}, errors.New("tls cert file and tls key file must be set together")
	}
	if s.tlsCertFile != "" || s.tlsKeyFile != "" {
		return tls.LoadX509KeyPair(s.tlsCertFile, s.tlsKeyFile)
	}
	return newSelfSignedCertificate(address)
}
