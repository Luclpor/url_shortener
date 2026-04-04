package mocks

import (
	"context"
	"sync"

	"github.com/Luclpor/url_shortener.git/internal/model"
)

type URLRepoMock struct {
	mu sync.RWMutex

	ByShort map[string]string // short -> full
	ByFull  map[string]string // full  -> short

	// опционально: для проверок в тестах
	SaveCalls int
}

func NewURLRepoMock() *URLRepoMock {
	return &URLRepoMock{
		ByShort: make(map[string]string),
		ByFull:  make(map[string]string),
	}
}

func (r *URLRepoMock) FindByShortURL(_ context.Context, short string) (*model.URL, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	full, ok := r.ByShort[short]
	if !ok {
		return nil, false
	}
	u := model.URL{ShortURL: short, FullUrl: full}
	return &u, true
}

func (r *URLRepoMock) FindByLongURL(_ context.Context, full string) (*model.URL, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	short, ok := r.ByFull[full]
	if !ok {
		return nil, false
	}
	u := model.URL{ShortURL: short, FullUrl: full}
	return &u, true
}

func (r *URLRepoMock) Save(_ context.Context, shortURL, fullURL string) (*model.URL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.SaveCalls++
	r.ByShort[shortURL] = fullURL
	r.ByFull[fullURL] = shortURL

	u := model.URL{ShortURL: shortURL, FullUrl: fullURL}
	return &u, nil
}
