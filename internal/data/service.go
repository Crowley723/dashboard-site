package data

import (
	"context"
	"fmt"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/utils"
	"log/slog"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Service struct {
	client  *MimirClient
	cache   CacheProvider
	logger  *slog.Logger
	queries []config.PrometheusQuery
}

func NewService(client *MimirClient, cache CacheProvider, logger *slog.Logger, queries []config.PrometheusQuery) *Service {
	return &Service{
		client:  client,
		cache:   cache,
		logger:  logger,
		queries: queries,
	}
}

func (s *Service) ExecuteQueries(ctx context.Context) error {
	for _, queryConfig := range s.queries {

		if queryConfig.Disabled {
			continue
		}

		if err := s.executeQuery(ctx, queryConfig); err != nil {
			s.logger.Error("failed to execute query",
				"query", queryConfig.Name,
				"error", err)
			continue
		}
	}
	return nil
}

func (s *Service) executeQuery(ctx context.Context, config config.PrometheusQuery) error {
	var result model.Value
	var err error

	switch config.Type {
	case "range":
		rangeDuration, err := utils.ParseDurationString(config.Range)
		if err != nil {
			return fmt.Errorf("invalid range duration %s: %w", config.Range, err)
		}

		stepDuration, err := utils.ParseDurationString(config.Step)
		if err != nil {
			return fmt.Errorf("invalid step duration %s: %w", config.Step, err)
		}

		end := time.Now()
		start := end.Add(-rangeDuration)

		r := v1.Range{
			Start: start,
			End:   end,
			Step:  stepDuration,
		}

		result, err = s.client.QueryRange(ctx, config.Query, r)

	default:
		result, err = s.client.Query(ctx, config.Query, time.Now())

	}

	if err != nil {
		return fmt.Errorf("failed to execute query %s: %w", config.Name, err)
	}

	if s.cache != nil {
		ttl := config.TTL
		if ttl == 0 {
			ttl = 5 * time.Minute
		}
		s.cache.Set(config.Name, result, config.RequireAuth, config.RequiredGroup)
		s.logger.Debug("cached query result", "query", config.Name, "type", config.Type, "ttl", ttl)
	} else {
		s.logger.Warn("cache is nil, skipping cache storage", "query", config.Name)
	}

	return nil
}

func (s *Service) GetCachedResult(queryName string) (model.Value, bool) {
	cached, found := s.cache.Get(queryName)
	if !found {
		return nil, false
	}
	return cached.Value, true
}
