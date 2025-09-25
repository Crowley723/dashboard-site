package server

import (
	"context"
	"fmt"
	"homelab-dashboard/internal/data"
	"log/slog"
	"time"
)

func runBackgroundDataFetching(ctx context.Context, dataService *data.Service, logger *slog.Logger, interval time.Duration) error {
	if interval == 0 {
		logger.Error("initial query execution failed: ticker interval must not be zero")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("starting background data fetching", "interval", interval)

	if err := dataService.ExecuteQueries(ctx); err != nil {
		logger.Error("initial query execution failed", "error", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := dataService.ExecuteQueries(ctx); err != nil {
				logger.Error(fmt.Sprintf("background query execution failed: trying again in %s", interval.String()), "error", err)
			}
		case <-ctx.Done():
			logger.Info("background data fetching stopped")
			return ctx.Err()
		}
	}
}
