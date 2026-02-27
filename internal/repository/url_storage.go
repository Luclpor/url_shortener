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

func (db *InMemoryDB) FindByShortURL(shortUrl string) (*model.URL, bool) {
	for _, u := range db.urls {
		if u.ShortURL == shortUrl {
			return &u, true
		}
	}
	return nil, false
}

func (db *InMemoryDB) FindByLongURL(longUrl string) (*model.URL, bool) {
	for _, u := range db.urls {
		if u.FullURL == longUrl {
			return &u, true
		}
	}
	return nil, false
}

func (db *InMemoryDB) Save(shortUrl string, fullURl string) (*model.URL, error) {
	u := model.URL{
		ShortURL: shortUrl,
		FullURL:  fullURl,
	}
	db.urls = append(db.urls, u)
	return &u, nil
}
