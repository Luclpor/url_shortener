package service

import (
	"fmt"

	"github.com/Luclpor/url_shortener.git/internal/model"
)

type URLRepository interface {
	FindByShortURL(shortUrl string) (*model.URL, bool)
	FindByLongURL(longUrl string) (*model.URL, bool)
	Save(shortUrl string, fullURl string) (*model.URL, error)
}

type URLManager struct {
	repo URLRepository
}

func NewURLManager(repo URLRepository) *URLManager {
	return &URLManager{repo: repo}
}

func (m *URLManager) TryCreateShortURL(longURL string) (string, bool, error) {
	if url, b := m.repo.FindByLongURL(longURL); b {
		return url.ShortURL, false, nil
	}
	key, b := m.getUniqueKey(longURL, 0)
	if !b {
		return "", false, fmt.Errorf("short url already exists")
	}
	url, err := m.repo.Save(key, longURL)
	if err != nil {
		return "", false, fmt.Errorf("failed to save url: %s", longURL)
	}
	return url.ShortURL, true, nil

}

func (m *URLManager) GetURL(shortURL string) (*model.URL, error) {
	url, b := m.repo.FindByShortURL(shortURL)
	if !b {
		return nil, fmt.Errorf("short url not found")
	}
	return url, nil
}

func (m *URLManager) getUniqueKey(longURL string, count int) (string, bool) {
	if count > 100 {
		return "", false
	}
	key := GenerateRandomString(5)
	_, ok := m.repo.FindByShortURL(key)
	if ok {
		m.getUniqueKey(longURL, count)
	}
	return key, true
}
