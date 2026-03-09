package repository

import (
	"context"
	"sync"

	"github.com/Luclpor/url_shortener.git/internal/model"
)

type InMemoryDB struct {
	m    sync.Mutex
	urls []model.URL
}

func NewRepository() *InMemoryDB {
	u := make([]model.URL, 0)
	return &InMemoryDB{
		urls: u,
		m:    sync.Mutex{},
	}
}

func (db *InMemoryDB) FindByShortURL(_ context.Context, shortURL string) (*model.URL, bool) {
	db.m.Lock()
	defer db.m.Unlock()
	for _, u := range db.urls {
		if u.ShortURL == shortURL {
			return &u, true
		}
	}
	return nil, false
}

func (db *InMemoryDB) FindByLongURL(_ context.Context, longURL string) (*model.URL, bool) {
	db.m.Lock()
	defer db.m.Unlock()
	for _, u := range db.urls {
		if u.FullURL == longURL {
			return &u, true
		}
	}
	return nil, false
}

func (db *InMemoryDB) Save(_ context.Context, shortURL string, fullURL string) (*model.URL, error) {
	db.m.Lock()
	defer db.m.Unlock()
	u := model.URL{
		ShortURL: shortURL,
		FullURL:  fullURL,
	}
	db.urls = append(db.urls, u)
	return &u, nil
}
