package postgres

import (
	"context"
	"errors"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	"github.com/Luclpor/url_shortener.git/internal/service"
	errors2 "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type URLRepository struct {
	pool *pgxpool.Pool
}

func NewURLRepository(pool *pgxpool.Pool) *URLRepository {
	return &URLRepository{pool: pool}
}

func (r *URLRepository) FindByShortURL(ctx context.Context, shortURL string) (*model.ShortenURL, bool) {
	const query = `
		SELECT short_url, original_url
		FROM url_shortener
		WHERE short_url = $1
	`

	var u model.ShortenURL
	err := r.pool.QueryRow(ctx, query, shortURL).Scan(&u.ShortURL, &u.OriginalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false
		}
		return nil, false
	}

	return &u, true
}

func (r *URLRepository) FindByOriginalURL(ctx context.Context, longURL string) (*model.ShortenURL, error) {
	const query = `
		SELECT short_url, original_url, correlation_id
		FROM url_shortener
		WHERE original_url = $1
	`
	var u model.ShortenURL
	err := r.pool.QueryRow(ctx, query, longURL).Scan(&u.ShortURL, &u.OriginalURL, &u.CorrelationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors2.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *URLRepository) Save(ctx context.Context, shortURL string, originalURL string) (*model.ShortenURL, error) {
	const query = `
		INSERT INTO url_shortener (short_url, original_url)
		VALUES ($1, $2)
		on conflict(original_url) do nothing
		RETURNING short_url, original_url
	`

	var u = new(model.ShortenURL)
	err := r.pool.QueryRow(ctx, query, shortURL, originalURL).Scan(&u.ShortURL, &u.OriginalURL)
	if errors.Is(err, pgx.ErrNoRows) {
		u, err = r.FindByOriginalURL(ctx, originalURL)
		if err != nil {
			return nil, err
		}
		return u, errors2.ErrAlreadyExists
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *URLRepository) SaveBatch(ctx context.Context, dtos []dto.URLDto) ([]model.ShortenURL, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	shortenURLs := make([]model.ShortenURL, 0, len(dtos))
	const query = `
		INSERT INTO url_shortener (short_url, original_url, correlation_id)
		VALUES ($1, $2, $3)
		RETURNING id, short_url, original_url, correlation_id, created_at, updated_at
	`

	for _, v := range dtos {
		var m model.ShortenURL
		err = tx.QueryRow(ctx, query, v.ShortURL, v.OriginalURL, v.CorrelationID).
			Scan(
				&m.ID,
				&m.ShortURL,
				&m.OriginalURL,
				&m.CorrelationID,
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

var _ service.URLRepository = (*URLRepository)(nil)
