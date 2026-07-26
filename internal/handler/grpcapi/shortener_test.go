package grpcapi

import (
	"context"
	"errors"
	"testing"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/grpc/pb"
	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/Luclpor/url_shortener.git/internal/storage/mock"
	appErrors "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type testAuth struct {
	user *model.User
}

func (a testAuth) CreateEncryptedUser() (*model.User, string, error) {
	return a.user, "token", nil
}

func (a testAuth) DecryptUser(string) (*model.User, error) {
	return a.user, nil
}

func (a testAuth) GetUserFromContext(context.Context) (*model.User, error) {
	return a.user, nil
}

func (a testAuth) SetUserOnContext(ctx context.Context, _ *model.User) context.Context {
	return ctx
}

func TestShortenerServerShortenURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRep := mock.NewMockURLRepository(ctrl)
	user := &model.User{ID: uuid.New()}

	mockRep.EXPECT().
		FindByShortURL(gomock.Any(), gomock.Any()).
		Return(nil, false)
	mockRep.EXPECT().
		Save(gomock.Any(), gomock.Any(), "https://example.com/article", user.ID).
		Return(&model.ShortenURL{
			ShortURL:    "abcde",
			OriginalURL: "https://example.com/article",
			UserID:      user.ID,
		}, nil)

	server := newTestShortenerServer(mockRep, user)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(authorizationMetadataKey, "token"))

	got, err := server.ShortenURL(ctx, &pb.URLShortenRequest{Url: "https://example.com/article"})

	require.NoError(t, err)
	assert.Equal(t, "http://short.test/abcde", got.GetResult())
}

func TestShortenerServerExpandURLNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRep := mock.NewMockURLRepository(ctrl)
	user := &model.User{ID: uuid.New()}

	mockRep.EXPECT().
		FindByShortURL(gomock.Any(), "missing").
		Return(nil, false)

	server := newTestShortenerServer(mockRep, user)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(authorizationMetadataKey, "token"))

	_, err := server.ExpandURL(ctx, &pb.URLExpandRequest{Id: "missing"})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestShortenerServerListUserURLsRequiresAuthorization(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRep := mock.NewMockURLRepository(ctrl)
	user := &model.User{ID: uuid.New()}

	server := newTestShortenerServer(mockRep, user)

	_, err := server.ListUserURLs(context.Background(), &emptypb.Empty{})

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestShortenerServerListUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRep := mock.NewMockURLRepository(ctrl)
	user := &model.User{ID: uuid.New()}

	mockRep.EXPECT().
		FindAllByUserID(gomock.Any(), user.ID).
		Return([]model.ShortenURL{
			{
				ShortURL:    "first",
				OriginalURL: "https://example.com/first",
				UserID:      user.ID,
			},
			{
				ShortURL:    "second",
				OriginalURL: "https://example.com/second",
				UserID:      user.ID,
			},
		}, nil)

	server := newTestShortenerServer(mockRep, user)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(authorizationMetadataKey, "token"))

	got, err := server.ListUserURLs(ctx, &emptypb.Empty{})

	require.NoError(t, err)
	require.Len(t, got.GetUrl(), 2)
	assert.Equal(t, "http://short.test/first", got.GetUrl()[0].GetShortUrl())
	assert.Equal(t, "https://example.com/first", got.GetUrl()[0].GetOriginalUrl())
	assert.Equal(t, "http://short.test/second", got.GetUrl()[1].GetShortUrl())
	assert.Equal(t, "https://example.com/second", got.GetUrl()[1].GetOriginalUrl())
}

func TestMapExpandURLErrorDoesNotExposeInternalDetails(t *testing.T) {
	err := mapExpandURLError(errors.New("database password is secret"))

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.NotContains(t, err.Error(), "database password")
}

func TestShortenerServerListUserURLsReturnsEmptyResponseForMissingURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRep := mock.NewMockURLRepository(ctrl)
	user := &model.User{ID: uuid.New()}

	mockRep.EXPECT().
		FindAllByUserID(gomock.Any(), user.ID).
		Return(nil, appErrors.ErrNotFound)

	server := newTestShortenerServer(mockRep, user)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(authorizationMetadataKey, "token"))

	got, err := server.ListUserURLs(ctx, &emptypb.Empty{})

	require.NoError(t, err)
	assert.Empty(t, got.GetUrl())
}

func newTestShortenerServer(repo service.URLRepository, user *model.User) *ShortenerServer {
	return NewShortenerServer(
		&config.Config{BaseAddressShort: "http://short.test"},
		service.NewURLManager(repo, zap.NewNop()),
		testAuth{user: user},
		nil,
		zap.NewNop(),
	)
}
