package inmemory

import (
	"context"
	"sync"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	"github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
)

type InMemoryDB struct {
	mu          sync.Mutex
	urls        []model.ShortenURL
	users       []model.User
	fileStorage *FileStorage
}

func NewRepository(fStorage *FileStorage) (*InMemoryDB, error) {
	urls := make([]model.ShortenURL, 0)
	users := make([]model.User, 0)
	err := fStorage.ScantTo(urls)
	if err != nil {
		return nil, err
	}

	return &InMemoryDB{
		urls:        urls,
		users:       users,
		fileStorage: fStorage,
	}, nil
}

func (db *InMemoryDB) FindBatchShortURLByUserID(_ context.Context, userID uuid.UUID) ([]model.ShortenURL, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	urls := make([]model.ShortenURL, 0)
	for i := range db.urls {
		if db.urls[i].UserID == userID {
			urls = append(urls, db.urls[i])
		}
	}
	return urls, nil
}

func (db *InMemoryDB) FindByShortURL(_ context.Context, shortURL string, userID uuid.UUID) (*model.ShortenURL, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.urls {
		if db.urls[i].ShortURL == shortURL && db.urls[i].UserID == userID {
			return &db.urls[i], true
		}
	}
	return nil, false
}

func (db *InMemoryDB) FindByOriginalURL(_ context.Context, originalURL string, userID uuid.UUID) (*model.ShortenURL, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.urls {
		if db.urls[i].OriginalURL == originalURL && db.urls[i].UserID == userID {
			return &db.urls[i], errors.ErrAlreadyExists
		}
	}
	return nil, nil
}

func (db *InMemoryDB) Save(_ context.Context, shortURL, fullURL string, userID uuid.UUID) (*model.ShortenURL, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	u := model.ShortenURL{
		ShortURL:    shortURL,
		OriginalURL: fullURL,
		UserID:      userID,
	}

	if err := db.fileStorage.SaveInFile(u); err != nil {
		return nil, err
	}

	db.urls = append(db.urls, u)
	return &u, nil
}

func (db *InMemoryDB) SaveBatch(_ context.Context, dtos []dto.URLDto) ([]model.ShortenURL, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	models := make([]model.ShortenURL, 0)
	for _, v := range dtos {
		m := model.ShortenURL{
			UserID:        v.UserID,
			ShortURL:      v.ShortURL,
			OriginalURL:   v.OriginalURL,
			CorrelationID: &v.CorrelationID,
		}
		db.urls = append(db.urls, m)
		models = append(models, m)
	}
	if err := db.fileStorage.SaveInFileBatch(models); err != nil {
		return nil, err
	}
	return models, nil
}

func (db *InMemoryDB) Close() error {
	return db.fileStorage.Close()
}
