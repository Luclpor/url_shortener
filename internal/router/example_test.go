package router_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/Luclpor/url_shortener.git/internal/model"
	modelapi "github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	"github.com/Luclpor/url_shortener.git/internal/router"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/Luclpor/url_shortener.git/internal/service/audit"
	appErrors "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type exampleUserContextKey struct{}

type exampleAuth struct {
	user *model.User
}

func (a exampleAuth) CreateEncryptedUser() (*model.User, string, error) {
	return a.user, "example-user-token", nil
}

func (a exampleAuth) DecryptUser(string) (*model.User, error) {
	return a.user, nil
}

func (a exampleAuth) GetUserFromContext(ctx context.Context) (*model.User, error) {
	user, ok := ctx.Value(exampleUserContextKey{}).(*model.User)
	if !ok {
		return nil, appErrors.ErrUserNotFound
	}
	return user, nil
}

func (a exampleAuth) SetUserOnContext(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, exampleUserContextKey{}, user)
}

type exampleHealthChecker struct{}

func (exampleHealthChecker) Ping(context.Context) error {
	return nil
}

type exampleURLRepository struct {
	mu    sync.Mutex
	user  *model.User
	urls  map[string]model.ShortenURL
	order []string
}

func newExampleURLRepository() *exampleURLRepository {
	return &exampleURLRepository{
		user: &model.User{
			ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		},
		urls: make(map[string]model.ShortenURL),
	}
}

func (r *exampleURLRepository) add(shortURL string, originalURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.putLocked(model.ShortenURL{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      r.user.ID,
	})
}

func (r *exampleURLRepository) putLocked(url model.ShortenURL) {
	if _, ok := r.urls[url.ShortURL]; !ok {
		r.order = append(r.order, url.ShortURL)
	}
	r.urls[url.ShortURL] = url
}

func (r *exampleURLRepository) FindByShortURLsAndUserID(_ context.Context, shortURLs []string, userID uuid.UUID) ([]model.ShortenURL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	found := make([]model.ShortenURL, 0, len(shortURLs))
	for _, shortURL := range shortURLs {
		stored, ok := r.urls[shortURL]
		if ok && stored.UserID == userID && !stored.IsDeleted {
			found = append(found, stored)
		}
	}
	if len(found) == 0 {
		return nil, appErrors.ErrNotFound
	}
	return found, nil
}

func (r *exampleURLRepository) FindAllByUserID(_ context.Context, userID uuid.UUID) ([]model.ShortenURL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	found := make([]model.ShortenURL, 0, len(r.urls))
	for _, shortURL := range r.order {
		stored := r.urls[shortURL]
		if stored.UserID == userID && !stored.IsDeleted {
			found = append(found, stored)
		}
	}
	if len(found) == 0 {
		return nil, appErrors.ErrNotFound
	}
	return found, nil
}

func (r *exampleURLRepository) FindByShortURL(_ context.Context, shortURL string) (*model.ShortenURL, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	stored, ok := r.urls[shortURL]
	if !ok {
		return nil, false
	}
	return &stored, true
}

func (r *exampleURLRepository) FindByOriginalURL(_ context.Context, longURL string, userID uuid.UUID) (*model.ShortenURL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, shortURL := range r.order {
		stored := r.urls[shortURL]
		if stored.OriginalURL == longURL && stored.UserID == userID && !stored.IsDeleted {
			return &stored, nil
		}
	}
	return nil, appErrors.ErrNotFound
}

func (r *exampleURLRepository) Save(_ context.Context, _ string, fullURL string, userID uuid.UUID) (*model.ShortenURL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, shortURL := range r.order {
		stored := r.urls[shortURL]
		if stored.OriginalURL == fullURL && stored.UserID == userID && !stored.IsDeleted {
			return &stored, appErrors.ErrAlreadyExists
		}
	}

	url := model.ShortenURL{
		ShortURL:    exampleShortKey(fullURL),
		OriginalURL: fullURL,
		UserID:      userID,
	}
	r.putLocked(url)
	return &url, nil
}

func (r *exampleURLRepository) SaveBatch(_ context.Context, dtos []dto.URLDto) ([]model.ShortenURL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	urls := make([]model.ShortenURL, 0, len(dtos))
	for _, item := range dtos {
		correlationID := item.CorrelationID
		url := model.ShortenURL{
			ShortURL:      exampleShortKey(item.OriginalURL),
			OriginalURL:   item.OriginalURL,
			UserID:        item.UserID,
			CorrelationID: &correlationID,
		}
		r.putLocked(url)
		urls = append(urls, url)
	}
	return urls, nil
}

func (r *exampleURLRepository) DeleteBatch(_ context.Context, deleteShortURLs map[uuid.UUID][]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for userID, shortURLs := range deleteShortURLs {
		for _, shortURL := range shortURLs {
			stored, ok := r.urls[shortURL]
			if ok && stored.UserID == userID {
				stored.IsDeleted = true
				r.urls[shortURL] = stored
			}
		}
	}
	return nil
}

func exampleShortKey(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err == nil {
		if base := path.Base(parsedURL.Path); base != "." && base != "/" {
			return base
		}
		if parsedURL.Hostname() != "" {
			return strings.Split(parsedURL.Hostname(), ".")[0]
		}
	}
	return "short"
}

func newExampleHandler(repo *exampleURLRepository) http.Handler {
	cfg := &config.Config{BaseAddressShort: "http://short.test"}
	appLogger := zap.NewNop()
	manager := service.NewURLManager(repo, appLogger)
	healthService := service.NewHealthService(exampleHealthChecker{})
	eventPublisher := audit.NewEvent(appLogger)
	handler, err := router.NewRouter(cfg, manager, healthService, exampleAuth{user: repo.user}, eventPublisher, appLogger)
	if err != nil {
		panic(err)
	}
	return handler
}

func ExampleNewRouter_plainTextShortening() {
	handler := newExampleHandler(newExampleURLRepository())
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com/article"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(res.StatusCode)
	fmt.Println(strings.HasPrefix(res.Header.Get("Content-Type"), "text/plain"))
	fmt.Println(strings.TrimSpace(string(body)))

	// Output:
	// 201
	// true
	// http://short.test/article
}

func ExampleNewRouter_jsonShortening() {
	handler := newExampleHandler(newExampleURLRepository())
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://example.com/json"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	var response modelapi.ShortenResp
	_ = json.NewDecoder(res.Body).Decode(&response)

	fmt.Println(res.StatusCode)
	fmt.Println(strings.HasPrefix(res.Header.Get("Content-Type"), "application/json"))
	fmt.Println(response.Result)

	// Output:
	// 201
	// true
	// http://short.test/json
}

func ExampleNewRouter_redirect() {
	repo := newExampleURLRepository()
	repo.add("practicum", "https://practicum.yandex.ru/")
	handler := newExampleHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/practicum", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	fmt.Println(res.Header.Get("Location"))

	// Output:
	// 307
	// https://practicum.yandex.ru/
}

func ExampleNewRouter_batchShortening() {
	handler := newExampleHandler(newExampleURLRepository())
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(`[
		{"correlation_id":"first","original_url":"https://example.com/a"},
		{"correlation_id":"second","original_url":"https://example.com/b"}
	]`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	var response []modelapi.ShortenBatchResp
	_ = json.NewDecoder(res.Body).Decode(&response)

	fmt.Println(res.StatusCode)
	for _, item := range response {
		fmt.Printf("%s %s\n", item.CorrelationID, item.ShortURL)
	}

	// Output:
	// 201
	// first http://short.test/a
	// second http://short.test/b
}

func ExampleNewRouter_userURLs() {
	repo := newExampleURLRepository()
	repo.add("article", "https://example.com/article")
	repo.add("docs", "https://example.com/docs")
	handler := newExampleHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	var response []model.ShortenURL
	_ = json.NewDecoder(res.Body).Decode(&response)

	fmt.Println(res.StatusCode)
	for _, item := range response {
		fmt.Printf("%s -> %s\n", item.ShortURL, item.OriginalURL)
	}

	// Output:
	// 200
	// http://short.test/article -> https://example.com/article
	// http://short.test/docs -> https://example.com/docs
}

func ExampleNewRouter_deleteUserURLs() {
	repo := newExampleURLRepository()
	repo.add("article", "https://example.com/article")
	handler := newExampleHandler(repo)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`["article"]`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	// Output:
	// 202
}
