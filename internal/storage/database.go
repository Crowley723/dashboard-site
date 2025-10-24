package storage

import (
	"context"
	"database/sql/driver"
	"fmt"
	"homelab-dashboard/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseProvider struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

func NewDatabaseProvider(ctx context.Context, cfg *config.Config) (*DatabaseProvider, error) {
	dbPool, err := pgxpool.New(ctx, GetConnectionStringFromConfig(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := dbPool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseProvider{pool: dbPool}, nil
}

func (p *DatabaseProvider) GetPool() *pgxpool.Pool {
	return p.pool
}

func (p *DatabaseProvider) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}

func (p *DatabaseProvider) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *DatabaseProvider) RunMigrations(ctx context.Context) error {
	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	driver, err := postgres.

}
