package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/Luclpor/url_shortener.git/internal/model"
)

type InMemoryDB struct {
	mu      sync.Mutex
	urls    []model.URL
	file    *os.File
	encoder *json.Encoder
}

func NewRepository(filePath string) (*InMemoryDB, error) {
	urls := make([]model.URL, 0)

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var u model.URL
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

func (db *InMemoryDB) FindByShortURL(_ context.Context, shortURL string) (*model.URL, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.urls {
		if db.urls[i].ShortURL == shortURL {
			return &db.urls[i], true
		}
	}
	return nil, false
}

func (db *InMemoryDB) FindByLongURL(_ context.Context, longURL string) (*model.URL, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.urls {
		if db.urls[i].FullURL == longURL {
			return &db.urls[i], true
		}
	}
	return nil, false
}

func (db *InMemoryDB) Save(_ context.Context, shortURL, fullURL string) (*model.URL, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	u := model.URL{
		ShortURL: shortURL,
		FullURL:  fullURL,
	}

	if err := db.encoder.Encode(u); err != nil {
		return nil, err
	}

	db.urls = append(db.urls, u)
	return &u, nil
}

func (db *InMemoryDB) Close() error {
	return db.file.Close()
}
