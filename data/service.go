package data

import (
	"context"
	"fmt"
	"homelab-dashboard/config"
	"log/slog"
	"time"

	"github.com/prometheus/common/model"
)

type Service struct {
	client  *MimirClient
	cache   *Cache
	logger  *slog.Logger
	queries []config.PrometheusQuery
}

func NewService(client *MimirClient, cache *Cache, logger *slog.Logger, queries []config.PrometheusQuery) *Service {
	return &Service{
		client:  client,
		cache:   cache,
		logger:  logger,
		queries: queries,
	}
}

func (s *Service) ExecuteQueries(ctx context.Context) error {
	for _, queryConfig := range s.queries {
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
	result, err := s.client.Query(ctx, config.Query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to execute query %s: %w", config.Name, err)
	}

	// Cache result with default TTL if not specified
	ttl := config.TTL
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	s.cache.Set(config.Name, result, ttl)
	s.logger.Debug("cached query result", "query", config.Name, "ttl", ttl)

	return nil
}

func (s *Service) GetCachedResult(queryName string) (model.Value, bool) {
	cached, found := s.cache.Get(queryName)
	if !found {
		return nil, false
	}
	return cached.Value, true
}
