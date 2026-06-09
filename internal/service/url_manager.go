package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/api"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	appErrors "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockgen -source=url_manager.go -destination=../storage/mock/mock_user_repository.go -package=mock

const maxLimitMessages = 1000

// URLRepository defines storage operations required by URLManager.
type URLRepository interface {
	// FindByShortURLsAndUserID returns active URL records for the specified user and short URL keys.
	FindByShortURLsAndUserID(ctx context.Context, shortURLs []string, userID uuid.UUID) ([]model.ShortenURL, error)
	// FindAllByUserID returns all active URL records owned by a user.
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error)
	// FindByShortURL returns a URL record by its short URL key.
	FindByShortURL(ctx context.Context, shortURL string) (*model.ShortenURL, bool)
	// FindByOriginalURL returns a user's URL record by the original URL.
	FindByOriginalURL(ctx context.Context, longURL string, userID uuid.UUID) (*model.ShortenURL, error)
	// Save stores a single URL mapping.
	Save(ctx context.Context, shortURL string, fullURL string, userID uuid.UUID) (*model.ShortenURL, error)
	// SaveBatch stores several URL mappings in one operation.
	SaveBatch(ctx context.Context, dtos []dto.URLDto) ([]model.ShortenURL, error)
	// DeleteBatch marks short URL keys as deleted for their users.
	DeleteBatch(ctx context.Context, deleteShortURLs map[uuid.UUID][]string) error
}

// URLManager coordinates URL shortening, lookup, listing, and asynchronous deletion.
type URLManager struct {
	repo             URLRepository
	msgChan          chan dto.URLDto
	maxLimitMessages int
	appLogger        *zap.Logger
}

// NewURLManager creates a URLManager and starts its background deletion flusher.
func NewURLManager(repo URLRepository, appLogger *zap.Logger) *URLManager {
	um := &URLManager{
		repo:             repo,
		msgChan:          make(chan dto.URLDto, 100),
		appLogger:        appLogger,
		maxLimitMessages: maxLimitMessages,
	}

	go um.flushMessages()

	return um
}

// CreateShortURL creates or returns a short URL for one original URL.
func (m *URLManager) CreateShortURL(ctx context.Context, originalURL string, user *model.User) (*api.ShortenResp, error) {
	var resultAPIModel *api.ShortenResp
	key, b := m.getUniqueKey(ctx, originalURL, 0, user.ID)
	if !b {
		m.appLogger.Error("failed get unique key", zap.String("originalURL", originalURL))
		return nil, errors.New("failed create shorten url")
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

// CreateBatchURL creates short URLs for a batch shortening request.
func (m *URLManager) CreateBatchURL(ctx context.Context, apiModels []api.CreateShortenReq, user *model.User) ([]api.ShortenBatchResp, error) {
	toAdd := make([]dto.URLDto, 0)
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

// GetURL resolves a short URL key to its stored URL record.
func (m *URLManager) GetURL(ctx context.Context, shortURL string) (*model.ShortenURL, error) {
	url, b := m.repo.FindByShortURL(ctx, shortURL)
	if !b || url == nil {
		return nil, appErrors.ErrNotFound
	}
	if url.IsDeleted {
		return nil, appErrors.ErrURLWasDeleted
	}
	return url, nil
}

// GetBatchURLByUserID returns all active short URLs created by a user.
func (m *URLManager) GetBatchURLByUserID(ctx context.Context, user *model.User) ([]model.ShortenURL, error) {
	urls, err := m.repo.FindAllByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return urls, nil
}

func (m *URLManager) getUniqueKey(ctx context.Context, longURL string, count int, userID uuid.UUID) (string, bool) {
	if count > 100 {
		return "", false
	}
	key := GenerateRandomString(5)
	_, ok := m.repo.FindByShortURL(ctx, key)
	if ok {
		m.getUniqueKey(ctx, longURL, count, userID)
	}
	return key, true
}

// DeleteBatch queues a user's short URLs for asynchronous deletion.
func (m *URLManager) DeleteBatch(ctx context.Context, shortURLs []string, user *model.User) error {
	urls, err := m.repo.FindByShortURLsAndUserID(ctx, shortURLs, user.ID)
	if err != nil {
		return err
	}
	if len(urls) != len(shortURLs) {
		return appErrors.ErrNotFound
	}
	for _, url := range urls {
		m.msgChan <- dto.URLDto{
			ShortURL: url.ShortURL,
			UserID:   user.ID,
		}
	}
	return nil
}

func (m *URLManager) flushMessages() {
	// будем сохранять сообщения, накопленные за последние 10 секунд
	ticker := time.NewTicker(10 * time.Second)

	messages := make(map[uuid.UUID][]string)

	for {
		select {
		case msg := <-m.msgChan:
			messages[msg.UserID] = append(messages[msg.UserID], msg.ShortURL)
		case <-ticker.C:
			// подождём, пока придёт хотя бы одно сообщение
			if len(messages) == 0 {
				continue
			}
			if len(messages) > m.maxLimitMessages {
				m.appLogger.Warn("max limit exceeded", zap.Int("count", len(messages)))
				messages = make(map[uuid.UUID][]string)
			}
			// сохраним все пришедшие сообщения одновременно
			err := m.repo.DeleteBatch(context.Background(), messages)
			if err != nil {
				// не будем стирать сообщения, попробуем отправить их чуть позже
				continue
			}
			// сотрём успешно отосланные сообщения
			messages = make(map[uuid.UUID][]string)
		}
	}
}
