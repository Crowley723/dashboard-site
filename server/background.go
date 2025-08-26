package server

import (
	"context"
	"homelab-dashboard/config"
	"homelab-dashboard/data"
	"log/slog"
	"time"
)

func runBackgroundDataFetching(ctx context.Context, dataService *data.Service, logger *slog.Logger, cfg *config.Config) error {
	interval := cfg.Data.TimeInterval

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("starting background data fetching", "interval", interval)

	// Run once immediately
	if err := dataService.ExecuteQueries(ctx); err != nil {
		logger.Error("initial query execution failed", "error", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := dataService.ExecuteQueries(ctx); err != nil {
				logger.Error("background query execution failed", "error", err)
			}
		case <-ctx.Done():
			logger.Info("background data fetching stopped")
			return ctx.Err()
		}
	}
}
