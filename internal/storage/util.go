package storage

import (
	"homelab-dashboard/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetConnectionStringFromConfig(cfg *config.Config) string {
	pgConnCfg := &pgconn.Config{
		Host:     cfg.Storage.Host,
		Port:     uint16(cfg.Storage.Port),
		Database: cfg.Storage.Database,
		User:     cfg.Storage.Username,
		Password: cfg.Storage.Password,
	}

	connConfig := &pgx.ConnConfig{
		Config: *pgConnCfg,
	}

	p := &pgxpool.Config{ConnConfig: connConfig}

	return p.ConnString()
}
