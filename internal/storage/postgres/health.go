package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthRepository struct {
	pool *pgxpool.Pool
}

func NewHealthRepository(pool *pgxpool.Pool) *HealthRepository {
	return &HealthRepository{pool: pool}
}

func (r *HealthRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
