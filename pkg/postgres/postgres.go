// Package postgres opens PostgreSQL connection pools using pgx.
package postgres

import (
	"context"
	"fmt"

	"tvtn/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// New opens a pool and verifies connectivity before returning it.
// Use a context with a deadline to bound connection setup. The caller must
// close the returned pool when it is no longer needed.
func New(ctx context.Context, cfg *config.Postgres) (*pgxpool.Pool, error) {
	dsn, err := cfg.DSN()
	if err != nil {
		return nil, fmt.Errorf("postgres.New: %w", err)
	}

	connectionString, err := dsn.ExposeAsString()
	if err != nil {
		return nil, fmt.Errorf("postgres.New: %w", err)
	}

	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("postgres.New: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres.New: %w", err)
	}

	return pool, nil
}
