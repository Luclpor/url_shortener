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
		SELECT short_url, original_url
		FROM url_shortener
		WHERE original_url = $1
	`

	var u model.ShortenURL
	err := r.pool.QueryRow(ctx, query, longURL).Scan(&u.ShortURL, &u.OriginalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &u, errors2.ErrAlreadyExists
}

func (r *URLRepository) Save(ctx context.Context, shortURL string, fullURL string) (*model.ShortenURL, error) {
	const query = `
		INSERT INTO url_shortener (short_url, original_url)
		VALUES ($1, $2)
		RETURNING short_url, full_url
	`

	var u model.ShortenURL
	err := r.pool.QueryRow(ctx, query, shortURL, fullURL).Scan(&u.ShortURL, &u.OriginalURL)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *URLRepository) SaveBatch(ctx context.Context, dtos []dto.URLDto) ([]model.ShortenURL, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	shortenURLs := make([]model.ShortenURL, 0)
	_, err = tx.Prepare(ctx, "batchSave", "insert into url_shortener (short_url, original_url, correlation_id) values ($1, $2, $3)"+
		"returning id, short_url, original_url, correlation_id, created_at, updated_at")
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}
	for _, v := range dtos {
		m := model.ShortenURL{}
		err = tx.QueryRow(ctx, "batchSave", v.ShortURL, v.OriginalURL, v.CorrelationID).
			Scan(
				&m.Id,
				&m.ShortURL,
				&m.OriginalURL,
				&m.CorrelationID,
				&m.CreatedAt,
				&m.UpdatedAt,
			)
		if err != nil {
			tx.Rollback(ctx)
			return nil, err
		}
		shortenURLs = append(shortenURLs, m)
	}
	return shortenURLs, tx.Commit(ctx)
}

var _ service.URLRepository = (*URLRepository)(nil)
