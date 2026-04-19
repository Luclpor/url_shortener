package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	"github.com/Luclpor/url_shortener.git/internal/service/auth"
	appErrors "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
)

//go:generate mockgen -source=url_manager.go -destination=../storage/mock/mock_user_repository.go -package=mock

type URLRepository interface {
	FindBatchShortURLByUserID(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error)
	FindByShortURL(ctx context.Context, shortURL string) (*model.ShortenURL, bool)
	FindByOriginalURL(ctx context.Context, longURL string, userId uuid.UUID) (*model.ShortenURL, error)
	Save(ctx context.Context, shortURL string, fullURL string, userId uuid.UUID) (*model.ShortenURL, error)
	SaveBatch(ctx context.Context, dtos []dto.URLDto) ([]model.ShortenURL, error)
}

type URLManager struct {
	repo URLRepository
}

func NewURLManager(repo URLRepository) *URLManager {
	return &URLManager{repo: repo}
}

func (m *URLManager) CreateShortURL(ctx context.Context, originalURL string) (*api.ShortenResp, error) {
	var resultAPIModel *api.ShortenResp
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	key, b := m.getUniqueKey(ctx, originalURL, 0, user.ID)
	if !b {
		return nil, fmt.Errorf("short url already exists")
	}
	url, err := m.repo.Save(ctx, key, originalURL, user.ID)
	if url != nil {
		resultAPIModel = &api.ShortenResp{Result: url.ShortURL}
	}
	if err != nil {
		return resultAPIModel, fmt.Errorf("failed to save url: %s, err: %w", originalURL, err)
	}
	return resultAPIModel, nil

}

func (m *URLManager) CreateBatchURL(ctx context.Context, apiModels []api.CreateShortenReq) ([]api.ShortenBatchResp, error) {
	toAdd := make([]dto.URLDto, 0)
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	existsModels := make([]model.ShortenURL, 0)
	for _, v := range apiModels {
		existModel, err := m.repo.FindByOriginalURL(ctx, v.OriginalURL, user.ID)
		if err != nil && !errors.Is(err, appErrors.ErrNotFound) {
			return nil, err
		}
		if existModel != nil {
			existsModels = append(existsModels, *existModel)
			continue
		}
		var sURL string
		var success bool
		if sURL, success = m.getUniqueKey(ctx, v.OriginalURL, 0, user.ID); !success {
			return nil, fmt.Errorf("short url already exists")
		}
		toAdd = append(toAdd, dto.URLDto{
			OriginalURL:   v.OriginalURL,
			ShortURL:      sURL,
			CorrelationID: v.CorrelationID,
			UserID:        user.ID,
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
		return nil, appErrors.ErrNotFound
	}
	return url, nil
}

func (m *URLManager) GetBatchURLByUserID(ctx context.Context) ([]model.ShortenURL, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	urls, err := m.repo.FindBatchShortURLByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return urls, nil
}

func (m *URLManager) getUniqueKey(ctx context.Context, longURL string, count int, userId uuid.UUID) (string, bool) {
	if count > 100 {
		return "", false
	}
	key := GenerateRandomString(5)
	_, ok := m.repo.FindByShortURL(ctx, key)
	if ok {
		m.getUniqueKey(ctx, longURL, count, userId)
	}
	return key, true
}
