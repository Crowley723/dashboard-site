package server

import (
	"context"
	"fmt"
	"time"
)

func (s *Server) runBackgroundDataFetching(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		s.logger.Error("initial query execution failed: ticker interval must not be zero")
		return fmt.Errorf("non-positive ticker interval: %s", interval)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("starting background data fetching", "interval", interval)

	if err := s.dataService.ExecuteQueries(ctx, s.appCtx.Cache); err != nil {
		s.logger.Error("initial query execution failed", "error", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := s.dataService.ExecuteQueries(ctx, s.appCtx.Cache); err != nil {
				s.logger.Error(fmt.Sprintf("background query execution failed: trying again in %s", interval.String()), "error", err)
			}
		case <-ctx.Done():
			s.logger.Info("background data fetching canceled")
			return ctx.Err()
		}
	}
}
