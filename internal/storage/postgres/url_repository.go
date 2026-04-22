package postgres

import (
	"context"
	"errors"

	"github.com/Luclpor/url_shortener.git/internal/logger"
	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/Luclpor/url_shortener.git/internal/model/dto"
	"github.com/Luclpor/url_shortener.git/internal/service"
	appErrors "github.com/Luclpor/url_shortener.git/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type URLRepository struct {
	pool *pgxpool.Pool
}

func NewURLRepository(pool *pgxpool.Pool) *URLRepository {
	return &URLRepository{pool: pool}
}

func (r *URLRepository) FindBatchShortURLsByUserID(ctx context.Context, shortURLs []string, userID uuid.UUID) ([]model.ShortenURL, error) {
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

func (r *URLRepository) FindBatchShortURLByUserID(ctx context.Context, userID uuid.UUID) ([]model.ShortenURL, error) {
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

func (r *URLRepository) FindByOriginalURL(ctx context.Context, longURL string, userId uuid.UUID) (*model.ShortenURL, error) {
	const query = `
		SELECT short_url, original_url, correlation_id, user_id
		FROM url_shortener
		WHERE original_url = $1 AND user_id = $2 AND is_deleted = false
	`
	logger.SugarLogger.Infow("user",
		"user_id", userId.String(),
	)
	var u model.ShortenURL
	err := r.pool.QueryRow(ctx, query, longURL, userId).Scan(&u.ShortURL, &u.OriginalURL, &u.CorrelationID, &u.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, appErrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *URLRepository) Save(ctx context.Context, shortURL string, originalURL string, userId uuid.UUID) (*model.ShortenURL, error) {
	const query = `
		INSERT INTO url_shortener (short_url, original_url, user_id)
		VALUES ($1, $2, $3)
		RETURNING short_url, original_url, user_id
	`
	logger.SugarLogger.Infow("user",
		"user_id", userId.String(),
	)
	existModel, err := r.FindByOriginalURL(ctx, originalURL, userId)
	if existModel != nil && err == nil {
		return existModel, appErrors.ErrAlreadyExists
	}
	if err != nil && !errors.Is(err, appErrors.ErrNotFound) {
		return nil, err
	}
	var u = new(model.ShortenURL)
	err = r.pool.QueryRow(ctx, query, shortURL, originalURL, userId).Scan(&u.ShortURL, &u.OriginalURL, &u.UserID)
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

var _ service.URLRepository = (*URLRepository)(nil)
