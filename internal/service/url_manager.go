package service

import (
	"context"
	"fmt"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
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
	var resultApiModel *api.ShortenResp
	key, b := m.getUniqueKey(ctx, originalURL, 0)
	if !b {
		return nil, fmt.Errorf("short url already exists")
	}
	url, err := m.repo.Save(ctx, key, originalURL)
	if url != nil {
		resultApiModel = &api.ShortenResp{Result: url.ShortURL}
	}
	if err != nil {
		return resultApiModel, fmt.Errorf("failed to save url: %s, err: %w", originalURL, err)
	}
	return resultApiModel, nil

}

func (m *URLManager) CreateBatchURL(ctx context.Context, apiModels []api.CreateShortenReq) ([]api.ShortenBatchResp, error) {
	toAdd := make([]dto.URLDto, 0)
	existsModels := make([]model.ShortenURL, 0)
	for _, v := range apiModels {
		existModel, err := m.repo.FindByOriginalURL(ctx, v.OriginalURL)
		if err != nil {
			return nil, err
		}
		if existModel != nil {
			existsModels = append(existsModels, *existModel)
			continue
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
	entities, err := m.repo.SaveBatch(ctx, toAdd)
	if err != nil {
		return nil, err
	}
	entities = append(entities, existsModels...)
	batchResps := make([]api.ShortenBatchResp, len(entities))
	for i, v := range entities {
		batchResps[i] = api.ShortenBatchResp{
			ShortURL:      v.ShortURL,
			CorrelationID: *v.CorrelationID,
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
