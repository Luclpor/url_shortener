package repository

import "github.com/Luclpor/url_shortener.git/internal/model"

type InMemoryDB struct {
	urls []model.URL
}

func NewRepository() *InMemoryDB {
	u := make([]model.URL, 0, 10)
	return &InMemoryDB{
		urls: u,
	}
}

func (db *InMemoryDB) FindByShortURL(shortURL string) (*model.URL, bool) {
	for _, u := range db.urls {
		if u.ShortURL == shortURL {
			return &u, true
		}
	}
	return nil, false
}

func (db *InMemoryDB) FindByLongURL(longURL string) (*model.URL, bool) {
	for _, u := range db.urls {
		if u.FullURL == longURL {
			return &u, true
		}
	}
	return nil, false
}

func (db *InMemoryDB) Save(shortURL string, fullURL string) (*model.URL, error) {
	u := model.URL{
		ShortURL: shortURL,
		FullURL:  fullURL,
	}
	db.urls = append(db.urls, u)
	return &u, nil
}
