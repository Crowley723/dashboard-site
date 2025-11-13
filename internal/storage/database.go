package storage

import (
	"context"
	"embed"
	"fmt"
	"homelab-dashboard/internal/config"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*
var migrationsFS embed.FS

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

	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
		    version VARCHAR(255) PRIMARY KEY,
		    applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		    )
		`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			migrations = append(migrations, entry.Name())
		}
	}

	sort.Strings(migrations)

	for _, filename := range migrations {
		version := strings.Split(filename, "_")[0]

		var exists bool
		err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)

		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", version, err)
		}

		if exists {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + filename)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		_, err = tx.Exec(ctx, string(content))
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", version, err)
		}
	}

	return nil
}
