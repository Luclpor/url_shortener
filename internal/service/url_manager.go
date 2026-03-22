package service

import (
	"context"
	"fmt"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/api"
)

//go:generate mockgen -source=url_manager.go -destination=mock/mock_user_repository.go -package=mock

type URLRepository interface {
	FindByShortURL(ctx context.Context, shortURL string) (*model.URL, bool)
	FindByLongURL(ctx context.Context, longURL string) (*model.URL, bool)
	Save(ctx context.Context, shortURL string, fullURL string) (*model.URL, error)
}

type URLManager struct {
	repo URLRepository
}

func NewURLManager(repo URLRepository) *URLManager {
	return &URLManager{repo: repo}
}

func (m *URLManager) TryCreateShortURL(ctx context.Context, longURL string) (*api.ShortenResp, bool, error) {
	if url, b := m.repo.FindByLongURL(ctx, longURL); b {
		return &api.ShortenResp{Result: url.ShortURL}, false, nil
	}
	key, b := m.getUniqueKey(ctx, longURL, 0)
	if !b {
		return nil, false, fmt.Errorf("short url already exists")
	}
	url, err := m.repo.Save(ctx, key, longURL)
	if err != nil {
		return nil, false, fmt.Errorf("failed to save url: %s", longURL)
	}
	return &api.ShortenResp{Result: url.ShortURL}, true, nil

}

func (m *URLManager) GetURL(ctx context.Context, shortURL string) (*model.URL, error) {
	url, b := m.repo.FindByShortURL(ctx, shortURL)
	if !b {
		return nil, fmt.Errorf("short url not found")
	}
	return url, nil
}

func (m *URLManager) getUniqueKey(ctx context.Context, longURL string, count int) (string, bool) {
	if count > 100 {
		return "", false
	}
	key := GenerateRandomString(5)
	_, ok := m.repo.FindByShortURL(ctx, key)
	if ok {
		m.getUniqueKey(ctx, longURL, count)
	}
	return key, true
}
