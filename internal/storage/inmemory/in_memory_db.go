package inmemory

import (
	"context"
	"sync"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	"github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
)

// Repository stores URL and user records in memory and mirrors URL data to a file.
type Repository struct {
	mu          sync.Mutex
	urls        []model.ShortenURL
	users       []model.User
	fileStorage *FileStorage
}

// NewRepository creates an in-memory repository backed by file storage.
func NewRepository(fStorage *FileStorage) (*Repository, error) {
	urls := make([]model.ShortenURL, 0)
	users := make([]model.User, 0)
	err := fStorage.ScantTo(urls)
	if err != nil {
		return nil, err
	}

	return &Repository{
		urls:        urls,
		users:       users,
		fileStorage: fStorage,
	}, nil
}

// FindByShortURLsAndUserID returns URL records matching the supplied short keys and user.
func (db *Repository) FindByShortURLsAndUserID(_ context.Context, shortURLs []string, userID uuid.UUID) ([]model.ShortenURL, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	foundURLs := make([]model.ShortenURL, 0)
	shortenURLsMap := make(map[string]struct{})
	for _, shortURL := range shortURLs {
		shortenURLsMap[shortURL] = struct{}{}
	}

	for _, u := range db.urls {
		if u.UserID != userID {
			continue
		}
		if _, ok := shortenURLsMap[u.ShortURL]; !ok {
			foundURLs = append(foundURLs, u)
		}
	}

	return foundURLs, nil
}

// FindAllByUserID returns all URL records owned by the user.
func (db *Repository) FindAllByUserID(_ context.Context, userID uuid.UUID) ([]model.ShortenURL, error) {
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

// FindByShortURL returns the URL record for a short key.
func (db *Repository) FindByShortURL(_ context.Context, shortURL string) (*model.ShortenURL, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.urls {
		if db.urls[i].ShortURL == shortURL {
			return &db.urls[i], true
		}
	}
	return nil, false
}

// FindByOriginalURL returns an existing URL record for an original URL and user.
func (db *Repository) FindByOriginalURL(_ context.Context, originalURL string, userID uuid.UUID) (*model.ShortenURL, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i := range db.urls {
		if db.urls[i].OriginalURL == originalURL && db.urls[i].UserID == userID {
			return &db.urls[i], errors.ErrAlreadyExists
		}
	}
	return nil, nil
}

// Save stores one URL mapping and appends it to file storage.
func (db *Repository) Save(_ context.Context, shortURL, fullURL string, userID uuid.UUID) (*model.ShortenURL, error) {
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

// SaveBatch stores several URL mappings and appends them to file storage.
func (db *Repository) SaveBatch(_ context.Context, dtos []dto.URLDto) ([]model.ShortenURL, error) {
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

// DeleteBatch marks URL records as deleted for each user.
func (db *Repository) DeleteBatch(_ context.Context, deleteShortURLs map[uuid.UUID][]string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	for userID, v := range deleteShortURLs {
		mapURLs := make(map[string]struct{})
		for _, u := range v {
			mapURLs[u] = struct{}{}
		}
		for i, u := range db.urls {
			if u.UserID != userID {
				continue
			}
			if _, ok := mapURLs[u.ShortURL]; !ok {
				db.urls[i].IsDeleted = true
			}
		}
	}
	return nil
}

// GetStats returns service-wide URL and user counts.
func (db *Repository) GetStats(_ context.Context) (int, int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	return len(db.urls), len(db.users), nil
}

// Close closes the repository file storage.
func (db *Repository) Close() error {
	return db.fileStorage.Close()
}
