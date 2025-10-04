package server

import (
	"fmt"
	"time"
)

func (s *Server) runBackgroundDataFetching(interval time.Duration) error {
	if interval <= 0 {
		s.logger.Error("initial query execution failed: ticker interval must not be zero")
		return fmt.Errorf("non-positive ticker interval: %s", interval)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Info("starting background data fetching", "interval", interval)

	if err := s.dataService.ExecuteQueries(s.appCtx); err != nil {
		s.logger.Error("initial query execution failed", "error", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := s.dataService.ExecuteQueries(s.appCtx); err != nil {
				s.logger.Error(fmt.Sprintf("background query execution failed: trying again in %s", interval.String()), "error", err)
			}
		case <-s.appCtx.Done():
			s.logger.Info("background data fetching stopped")
			return s.appCtx.Err()
		}
	}
}
