package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	errors2 "github.com/Luclpor/url_shortener.git/pkg/errors"
)

//go:generate mockgen -source=url_manager.go -destination=../storage/mock/mock_user_repository.go -package=mock

type URLRepository interface {
	FindByShortURL(ctx context.Context, shortURL string) (*model.ShortenURL, bool)
	FindByOriginalURL(ctx context.Context, longURL string) (*model.ShortenURL, error)
	Save(ctx context.Context, shortURL string, fullURL string) (*model.ShortenURL, error)
	SaveBatch(ctx context.Context, dtos []dto.URLDto) ([]model.ShortenURL, error)
}

type URLManager struct {
	repo URLRepository
}

func NewURLManager(repo URLRepository) *URLManager {
	return &URLManager{repo: repo}
}

func (m *URLManager) CreateShortURL(ctx context.Context, originalURL string) (*api.ShortenResp, error) {
	if url, err := m.repo.FindByOriginalURL(ctx, originalURL); err != nil {
		if errors.Is(err, errors2.ErrAlreadyExists) {
			return &api.ShortenResp{Result: url.ShortURL}, errors2.ErrAlreadyExists
		}
		return nil, err
	}
	key, b := m.getUniqueKey(ctx, originalURL, 0)
	if !b {
		return nil, fmt.Errorf("short url already exists")
	}
	url, err := m.repo.Save(ctx, key, originalURL)
	if err != nil {
		return nil, fmt.Errorf("failed to save url: %s", originalURL)
	}
	return &api.ShortenResp{Result: url.ShortURL}, nil

}

func (m *URLManager) CreateBatchURL(ctx context.Context, apiModels []api.CreateShortenReq) ([]api.ShortenBatchResp, error) {
	toAdd := make([]dto.URLDto, 0)
	for _, v := range apiModels {
		if _, err := m.repo.FindByOriginalURL(ctx, v.OriginalURL); err != nil {
			if errors.Is(err, errors2.ErrAlreadyExists) {
				continue
			} else {
				return nil, err
			}
		}
		var sURL string
		var success bool
		if sURL, success = m.getUniqueKey(ctx, v.OriginalURL, 0); !success {
			return nil, fmt.Errorf("short url already exists")
		}
		toAdd = append(toAdd, dto.URLDto{
			OriginalURL:   v.OriginalURL,
			ShortURL:      sURL,
			CorrelationID: v.CorrelationID,
		})
	}
	enities, err := m.repo.SaveBatch(ctx, toAdd)
	if err != nil {
		return nil, err
	}
	batchResps := make([]api.ShortenBatchResp, len(enities))
	for i, v := range enities {
		batchResps[i] = api.ShortenBatchResp{
			ShortURL:      v.ShortURL,
			CorrelationID: v.CorrelationID,
		}
	}
	return batchResps, nil
}

func (m *URLManager) GetURL(ctx context.Context, shortURL string) (*model.ShortenURL, error) {
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
