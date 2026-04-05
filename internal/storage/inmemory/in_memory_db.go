package inmemory

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	"github.com/Luclpor/url_shortener.git/pkg/errors"
)

type InMemoryDB struct {
	mu      sync.Mutex
	urls    []model.ShortenURL
	file    *os.File
	encoder *json.Encoder
}

func NewRepository(filePath string) (*InMemoryDB, error) {
	urls := make([]model.ShortenURL, 0)

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var u model.ShortenURL
		if err := json.Unmarshal(scanner.Bytes(), &u); err != nil {
			_ = file.Close()
			return nil, err
		}
		urls = append(urls, u)
	}
	if err := scanner.Err(); err != nil {
		_ = file.Close()
		return nil, err
	}

	return &InMemoryDB{
		urls:    urls,
		file:    file,
		encoder: json.NewEncoder(file),
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

	if err := db.encoder.Encode(u); err != nil {
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
			CorrelationID: v.CorrelationID,
		}
		if err := db.encoder.Encode(m); err != nil {
			return nil, err
		}
		db.urls = append(db.urls, m)
		models = append(models, m)
	}
	return models, nil
}

func (db *InMemoryDB) Close() error {
	return db.file.Close()
}
