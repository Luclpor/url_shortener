package inmemory

import (
	"context"
	"sync"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	"github.com/Luclpor/url_shortener.git/pkg/errors"
)

type InMemoryDB struct {
	mu          sync.Mutex
	urls        []model.ShortenURL
	fileStorage *FileStorage
}

func NewRepository(fStorage *FileStorage) (*InMemoryDB, error) {
	urls := make([]model.ShortenURL, 0)
	err := fStorage.ScantTo(urls)
	if err != nil {
		return nil, err
	}

	return &InMemoryDB{
		urls:        urls,
		fileStorage: fStorage,
	}, nil
}

func (db *InMemoryDB) FindByShortURL(_ context.Context, shortURL string) (*model.ShortenURL, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.urls {
		if db.urls[i].ShortURL == shortURL {
			return &db.urls[i], true
		}
	}
	return nil, false
}

func (db *InMemoryDB) FindByOriginalURL(_ context.Context, originalURL string) (*model.ShortenURL, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.urls {
		if db.urls[i].OriginalURL == originalURL {
			return &db.urls[i], errors.ErrAlreadyExists
		}
	}
	return nil, nil
}

func (db *InMemoryDB) Save(_ context.Context, shortURL, fullURL string) (*model.ShortenURL, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	u := model.ShortenURL{
		ShortURL:    shortURL,
		OriginalURL: fullURL,
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
