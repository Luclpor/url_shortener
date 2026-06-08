package service

import (
	"context"
	"testing"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	appErrors "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type benchmarkURLRepository struct {
	urlsByShort    map[string]*model.ShortenURL
	urlsByOriginal map[string]*model.ShortenURL
	saved          model.ShortenURL
}

func newBenchmarkURLRepository() *benchmarkURLRepository {
	return &benchmarkURLRepository{
		urlsByShort:    make(map[string]*model.ShortenURL),
		urlsByOriginal: make(map[string]*model.ShortenURL),
	}
}

func (r *benchmarkURLRepository) FindByShortURLsAndUserID(_ context.Context, shortURLs []string, userID uuid.UUID) ([]model.ShortenURL, error) {
	found := make([]model.ShortenURL, 0, len(shortURLs))
	for _, shortURL := range shortURLs {
		url, ok := r.urlsByShort[shortURL]
		if ok && url.UserID == userID {
			found = append(found, *url)
		}
	}
	return found, nil
}

func (r *benchmarkURLRepository) FindAllByUserID(_ context.Context, userID uuid.UUID) ([]model.ShortenURL, error) {
	found := make([]model.ShortenURL, 0, len(r.urlsByShort))
	for _, url := range r.urlsByShort {
		if url.UserID == userID {
			found = append(found, *url)
		}
	}
	return found, nil
}

func (r *benchmarkURLRepository) FindByShortURL(_ context.Context, shortURL string) (*model.ShortenURL, bool) {
	url, ok := r.urlsByShort[shortURL]
	if !ok {
		return nil, false
	}
	return url, true
}

func (r *benchmarkURLRepository) FindByOriginalURL(_ context.Context, longURL string, userID uuid.UUID) (*model.ShortenURL, error) {
	url, ok := r.urlsByOriginal[longURL]
	if !ok || url.UserID != userID {
		return nil, appErrors.ErrNotFound
	}
	return url, nil
}

func (r *benchmarkURLRepository) Save(_ context.Context, shortURL string, fullURL string, userID uuid.UUID) (*model.ShortenURL, error) {
	r.saved.ShortURL = shortURL
	r.saved.OriginalURL = fullURL
	r.saved.UserID = userID
	r.saved.IsDeleted = false
	return &r.saved, nil
}

func (r *benchmarkURLRepository) SaveBatch(_ context.Context, items []dto.URLDto) ([]model.ShortenURL, error) {
	urls := make([]model.ShortenURL, len(items))
	for i, item := range items {
		urls[i] = model.ShortenURL{
			ShortURL:      item.ShortURL,
			OriginalURL:   item.OriginalURL,
			UserID:        item.UserID,
			CorrelationID: &item.CorrelationID,
		}
		r.urlsByShort[item.ShortURL] = &urls[i]
		r.urlsByOriginal[item.OriginalURL] = &urls[i]
	}
	return urls, nil
}

func (r *benchmarkURLRepository) DeleteBatch(_ context.Context, deleteShortURLs map[uuid.UUID][]string) error {
	for userID, shortURLs := range deleteShortURLs {
		for _, shortURL := range shortURLs {
			url, ok := r.urlsByShort[shortURL]
			if ok && url.UserID == userID {
				url.IsDeleted = true
			}
		}
	}
	return nil
}

func BenchmarkGenerateRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateRandomString(5)
	}
}

func BenchmarkCreateShortURL(b *testing.B) {
	manager := &URLManager{
		repo:      newBenchmarkURLRepository(),
		appLogger: zap.NewNop(),
	}
	user := &model.User{ID: uuid.New()}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.CreateShortURL(ctx, "https://example.com/original", user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetURL(b *testing.B) {
	repo := newBenchmarkURLRepository()
	userID := uuid.New()
	url := model.ShortenURL{
		ShortURL:    "abcde",
		OriginalURL: "https://example.com/original",
		UserID:      userID,
	}
	repo.urlsByShort[url.ShortURL] = &url
	manager := &URLManager{repo: repo}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetURL(ctx, url.ShortURL)
		if err != nil {
			b.Fatal(err)
		}
	}
}
