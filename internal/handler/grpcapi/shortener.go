package grpcapi

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/grpc/pb"
	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/Luclpor/url_shortener.git/internal/service/audit"
	"github.com/Luclpor/url_shortener.git/internal/service/auth"
	appErrors "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const authorizationMetadataKey = "authorization"

// ShortenerServer implements the gRPC ShortenerService API.
type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer

	cfg            *config.Config
	manager        *service.URLManager
	authService    auth.UserAuthentication
	logger         *zap.Logger
	eventPublisher *audit.Event
}

// NewShortenerServer creates the gRPC facade for the URL shortener API.
func NewShortenerServer(cfg *config.Config, manager *service.URLManager, userAuth auth.UserAuthentication, eventPublisher *audit.Event, appLogger *zap.Logger) *ShortenerServer {
	if cfg == nil {
		cfg = &config.Config{}
	}
	if appLogger == nil {
		appLogger = zap.NewNop()
	}
	return &ShortenerServer{
		cfg:            cfg,
		manager:        manager,
		authService:    userAuth,
		logger:         appLogger,
		eventPublisher: eventPublisher,
	}
}

// ShortenURL creates or returns a shortened URL for the authenticated user.
func (s *ShortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	user, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	responseModel, err := s.manager.CreateShortURL(ctx, req.GetUrl(), user)
	if err != nil && !errors.Is(err, appErrors.ErrAlreadyExists) {
		s.logger.Error("failed to create short url:", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create short url")
	}
	if responseModel == nil {
		return nil, status.Error(codes.Internal, "failed to create short url")
	}

	result, err := url.JoinPath(s.cfg.BaseAddressShort, responseModel.Result)
	if err != nil {
		s.logger.Error("failed to join short url:", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to build short url")
	}
	s.publishAudit(&audit.EventAudit{
		URL:       req.GetUrl(),
		UserID:    user.ID,
		Action:    "shorten",
		Timestamp: time.Now().Unix(),
	})

	return &pb.URLShortenResponse{Result: result}, nil
}

// ExpandURL resolves a short URL key to the original URL.
func (s *ShortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	user, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	shortURL, err := s.manager.GetURL(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "url not found")
		}
		if errors.Is(err, appErrors.ErrURLWasDeleted) {
			return nil, status.Error(codes.FailedPrecondition, "url was deleted")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	s.publishAudit(&audit.EventAudit{
		URL:       shortURL.OriginalURL,
		Action:    "follow",
		UserID:    user.ID,
		Timestamp: time.Now().Unix(),
	})

	return &pb.URLExpandResponse{Result: shortURL.OriginalURL}, nil
}

// ListUserURLs returns all active shortened URLs owned by the authenticated user.
func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	user, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	urls, err := s.manager.GetBatchURLByUserID(ctx, user)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			return &pb.UserURLsResponse{}, nil
		}
		s.logger.Error("failed to get short urls:", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user urls")
	}

	response := &pb.UserURLsResponse{
		Url: make([]*pb.URLData, 0, len(urls)),
	}
	for _, item := range urls {
		shortURL, err := url.JoinPath(s.cfg.BaseAddressShort, item.ShortURL)
		if err != nil {
			s.logger.Error("failed to join short url:", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to build short url")
		}
		response.Url = append(response.Url, &pb.URLData{
			ShortUrl:    shortURL,
			OriginalUrl: item.OriginalURL,
		})
	}
	return response, nil
}

func (s *ShortenerServer) authenticate(ctx context.Context) (*model.User, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get(authorizationMetadataKey)) == 0 {
		return s.createUser(ctx)
	}

	token := strings.TrimSpace(md.Get(authorizationMetadataKey)[0])
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimPrefix(token, "bearer ")
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "authorization metadata is empty")
	}

	user, err := s.authService.DecryptUser(token)
	if err != nil || user == nil || user.ID == uuid.Nil {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization metadata")
	}
	return user, nil
}

func (s *ShortenerServer) createUser(ctx context.Context) (*model.User, error) {
	user, encryptedUserID, err := s.authService.CreateEncryptedUser()
	if err != nil {
		s.logger.Error("error creating new user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create user")
	}
	_ = grpc.SetHeader(ctx, metadata.Pairs(authorizationMetadataKey, encryptedUserID))
	return user, nil
}

func (s *ShortenerServer) publishAudit(event *audit.EventAudit) {
	if s.eventPublisher == nil {
		return
	}
	s.eventPublisher.Update(event)
}
