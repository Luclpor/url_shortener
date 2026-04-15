package postgres

import (
	"context"
	"fmt"

	"github.com/Luclpor/url_shortener.git/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg *config.PostgresConfig) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(cfg.DataBaseDSN)
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}
	pgxConfig.MaxConns = cfg.MaxConns
	pgxConfig.MinConns = cfg.MinConns
	pgxConfig.MaxConnLifetime = cfg.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	pgxConfig.HealthCheckPeriod = cfg.HealthCheckPeriod
	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}
