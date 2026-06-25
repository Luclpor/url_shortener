package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthRepository reports PostgreSQL connection health.
type HealthRepository struct {
	pool *pgxpool.Pool
}

// NewHealthRepository creates a PostgreSQL health checker.
func NewHealthRepository(pool *pgxpool.Pool) *HealthRepository {
	return &HealthRepository{pool: pool}
}

// Ping verifies that the PostgreSQL connection pool is reachable.
func (r *HealthRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
