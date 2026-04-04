package postgres

import (
	"context"
	"errors"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type URLRepository struct {
	pool *pgxpool.Pool
}

func NewURLRepository(pool *pgxpool.Pool) *URLRepository {
	return &URLRepository{pool: pool}
}

func (r *URLRepository) FindByShortURL(ctx context.Context, shortURL string) (*model.URL, bool) {
	const query = `
		SELECT short_url, full_url
		FROM url_shortener
		WHERE short_url = $1
	`

	var u model.URL
	err := r.pool.QueryRow(ctx, query, shortURL).Scan(&u.ShortURL, &u.FullURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false
		}
		return nil, false
	}

	return &u, true
}

func (r *URLRepository) FindByLongURL(ctx context.Context, longURL string) (*model.URL, bool) {
	const query = `
		SELECT short_url, full_url
		FROM url_shortener
		WHERE full_url = $1
	`

	var u model.URL
	err := r.pool.QueryRow(ctx, query, longURL).Scan(&u.ShortURL, &u.FullURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false
		}
		return nil, false
	}

	return &u, true
}

func (r *URLRepository) Save(ctx context.Context, shortURL string, fullURL string) (*model.URL, error) {
	const query = `
		INSERT INTO url_shortener (short_url, full_url)
		VALUES ($1, $2)
		RETURNING short_url, full_url
	`

	var u model.URL
	err := r.pool.QueryRow(ctx, query, shortURL, fullURL).Scan(&u.ShortURL, &u.FullURL)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

var _ service.URLRepository = (*URLRepository)(nil)
