package postgres

import (
	"context"
	"errors"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	"github.com/Luclpor/url_shortener.git/internal/service"
	appErrors "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// URLRepository stores URL and user records in PostgreSQL.
type URLRepository struct {
	pool *pgxpool.Pool
}

// NewURLRepository creates a PostgreSQL-backed URL repository.
func NewURLRepository(pool *pgxpool.Pool) *URLRepository {
	return &URLRepository{pool: pool}
}

// FindByShortURLsAndUserID returns active URL records for the specified user and short keys.
func (r *URLRepository) FindByShortURLsAndUserID(ctx context.Context, shortURLs []string, userID uuid.UUID) ([]model.ShortenURL, error) {
	const query = `
		SELECT short_url, original_url, user_id
		FROM url_shortener
		WHERE short_url = ANY($1)
		  AND user_id = $2 AND is_deleted = false
	`
	urls := []model.ShortenURL{}
	rows, err := r.pool.Query(ctx, query, shortURLs, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		url := model.ShortenURL{}
		err = rows.Scan(&url.ShortURL, &url.OriginalURL, &url.UserID)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, appErrors.ErrNotFound
	}
	return urls, nil
}

// FindAllByUserID returns all active URL records owned by a user.
func (r *URLRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error) {
	const query = `
		SELECT short_url, original_url, user_id
		FROM url_shortener
		WHERE user_id = $1 AND is_deleted = false
	`

	urls := []model.ShortenURL{}
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		url := model.ShortenURL{}
		err = rows.Scan(&url.ShortURL, &url.OriginalURL, &url.UserID)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, appErrors.ErrNotFound
	}
	return urls, nil
}

// FindByShortURL returns a URL record by its short key.
func (r *URLRepository) FindByShortURL(ctx context.Context, shortURL string) (*model.ShortenURL, bool) {
	const query = `
		SELECT short_url, original_url, user_id, is_deleted
		FROM url_shortener
		WHERE short_url = $1
	`

	var u model.ShortenURL
	err := r.pool.QueryRow(ctx, query, shortURL).Scan(&u.ShortURL, &u.OriginalURL, &u.UserID, &u.IsDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false
		}
		return nil, false
	}

	return &u, true
}

// FindByOriginalURL returns a user's active URL record by the original URL.
func (r *URLRepository) FindByOriginalURL(ctx context.Context, longURL string, userID uuid.UUID) (*model.ShortenURL, error) {
	const query = `
		SELECT short_url, original_url, correlation_id, user_id
		FROM url_shortener
		WHERE original_url = $1 AND user_id = $2 AND is_deleted = false
	`
	var u model.ShortenURL
	err := r.pool.QueryRow(ctx, query, longURL, userID).Scan(&u.ShortURL, &u.OriginalURL, &u.CorrelationID, &u.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Save inserts one URL mapping or returns an existing mapping for the same user.
func (r *URLRepository) Save(ctx context.Context, shortURL string, originalURL string, userID uuid.UUID) (*model.ShortenURL, error) {
	const query = `
		INSERT INTO url_shortener (short_url, original_url, user_id)
		VALUES ($1, $2, $3)
		RETURNING short_url, original_url, user_id
	`
	existModel, err := r.FindByOriginalURL(ctx, originalURL, userID)
	if existModel != nil && err == nil {
		return existModel, appErrors.ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, appErrors.ErrNotFound) {
		return nil, err
	}
	var u = new(model.ShortenURL)
	err = r.pool.QueryRow(ctx, query, shortURL, originalURL, userID).Scan(&u.ShortURL, &u.OriginalURL, &u.UserID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// SaveBatch inserts several URL mappings in one transaction.
func (r *URLRepository) SaveBatch(ctx context.Context, dtos []dto.URLDto) ([]model.ShortenURL, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	shortenURLs := make([]model.ShortenURL, 0, len(dtos))
	const query = `
		INSERT INTO url_shortener (short_url, original_url, correlation_id, user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, short_url, original_url, correlation_id, user_id, created_at, updated_at
	`

	for _, v := range dtos {
		var m model.ShortenURL
		err = tx.QueryRow(ctx, query, v.ShortURL, v.OriginalURL, v.CorrelationID, v.UserID).
			Scan(
				&m.ID,
				&m.ShortURL,
				&m.OriginalURL,
				&m.CorrelationID,
				&m.UserID,
				&m.CreatedAt,
				&m.UpdatedAt,
			)
		if err != nil {
			return nil, err
		}
		shortenURLs = append(shortenURLs, m)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return shortenURLs, nil
}

// DeleteBatch marks URL mappings as deleted for each user in one transaction.
func (r *URLRepository) DeleteBatch(ctx context.Context, deleteShortURLs map[uuid.UUID][]string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const query = `
		UPDATE url_shortener
		SET is_deleted = true
		WHERE short_url = ANY($1)
		  AND user_id = $2
	`

	for k, v := range deleteShortURLs {
		if len(v) == 0 {
			continue
		}
		_, err = tx.Exec(ctx, query, v, k)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// GetStats returns service-wide URL and user counts.
func (r *URLRepository) GetStats(ctx context.Context) (int, int, error) {
	const query = `
		SELECT
			(SELECT COUNT(*) FROM url_shortener),
			(SELECT COUNT(*) FROM service_user)
	`
	var urls int64
	var users int64
	err := r.pool.QueryRow(ctx, query).Scan(&urls, &users)
	if err != nil {
		return 0, 0, err
	}
	return int(urls), int(users), nil
}

var _ service.URLRepository = (*URLRepository)(nil)
