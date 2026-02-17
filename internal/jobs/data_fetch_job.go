package jobs

import (
	"context"
	"fmt"
	"homelab-dashboard/internal/data"
	"homelab-dashboard/internal/middlewares"
	"log/slog"
	"time"
)

type DataFetchJob struct {
	dataService *data.Service
	appCtx      *middlewares.AppContext
	interval    time.Duration
	logger      *slog.Logger
}

func NewDataFetchJob(dataService *data.Service, appCtx *middlewares.AppContext, interval time.Duration, logger *slog.Logger) *DataFetchJob {
	return &DataFetchJob{
		dataService: dataService,
		appCtx:      appCtx,
		interval:    interval,
		logger:      logger,
	}
}

func (j *DataFetchJob) Name() string {
	return "data_fetch"
}

func (j *DataFetchJob) RequiresLeadership() bool {
	return true
}

func (j *DataFetchJob) Interval() time.Duration {
	return j.interval
}

func (j *DataFetchJob) Run(ctx context.Context) error {
	if j.interval <= 0 {
		j.logger.Error("initial query execution failed: ticker interval must not be zero")
		return fmt.Errorf("non-positive ticker interval: %s", j.interval)
	}

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	j.logger.Debug("Starting background data fetching", "interval", j.interval)

	if err := j.dataService.ExecuteQueries(ctx, j.appCtx.Cache); err != nil {
		j.logger.Error("initial data fetch failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			j.logger.Debug("Background data fetching canceled")
			return ctx.Err()
		case <-ticker.C:
			if err := j.dataService.ExecuteQueries(ctx, j.appCtx.Cache); err != nil {
				j.logger.Error(fmt.Sprintf("Background data fetch failed, trying again in %s", j.interval.String()), "error", err)
			}
		}
	}
}
