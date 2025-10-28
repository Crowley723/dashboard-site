package data

import (
	"context"
	"encoding/json"
	"fmt"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/metrics"
	"homelab-dashboard/internal/utils"
	"log/slog"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
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
		queries: queries,
  logger: logger,
	}
}

func (s *Service) ExecuteQueries(ctx context.Context, cache CacheProvider) error {
	if cache == nil {
		cache = s.cache
	}

	if cache == nil {
		s.logger.Error("cache is nil, skipping metrics update")
		return nil
	}

	cacheType := "memory"
	if _, ok := cache.(*RedisCache); ok {
		cacheType = "redis"
	}

	size := cache.Size(ctx)
	metrics.CacheItems.WithLabelValues(cacheType).Set(float64(size))

	for _, queryConfig := range s.queries {

		if queryConfig.Disabled {
			continue
		}

		if err := s.executeQuery(ctx, cache, queryConfig); err != nil {
			s.logger.Error("failed to execute query",
				"query", queryConfig.Name,
				"error", err)
			continue
		}
	}
	return nil
}

func (s *Service) executeQuery(ctx context.Context, cache CacheProvider, config config.PrometheusQuery) error {
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

		timer := prometheus.NewTimer(metrics.DataFetchDuration.WithLabelValues(config.Name, metrics.DataSourceTypeMimir))
		result, err = s.client.QueryRange(ctx, config.Query, r)
		timer.ObserveDuration()

	default:
		timer := prometheus.NewTimer(metrics.DataFetchDuration.WithLabelValues(config.Name, metrics.DataSourceTypeMimir))
		result, err = s.client.Query(ctx, config.Query, time.Now())
		timer.ObserveDuration()

	}

	if err != nil {
		metrics.DataFetchErrors.WithLabelValues(config.Name, metrics.DataSourceTypeMimir).Inc()
		return fmt.Errorf("failed to execute query %s: %w", config.Name, err)
	}

	// Add nil check for result before processing
	if result == nil {
		s.logger.Warn("query returned nil result", "query", config.Name)
		return nil
	}

	if s.cache != nil {
		ttl := config.TTL
		if ttl == 0 {
			ttl = 5 * time.Minute
		}

		cachedData := s.prepareCacheData(config.Name, result, config)

		cache.Set(ctx, config.Name, cachedData)
		s.logger.Debug("cached query result", "query", config.Name, "type", config.Type, "ttl", ttl)
	} else {
		s.logger.Warn("cache is nil, skipping cache storage", "query", config.Name)
	}

	return nil
}

func (s *Service) prepareCacheData(name string, value model.Value, config config.PrometheusQuery) CachedData {
	if value == nil {
		s.logger.Error("received nil value for cache preparation", "query", name)
		return CachedData{
			Name:          name,
			ValueType:     "unknown",
			JSONBytes:     []byte("null"),
			Timestamp:     time.Now(),
			RequireAuth:   config.RequireAuth,
			RequiredGroup: config.RequiredGroup,
		}
	}

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		s.logger.Error("failed to marshal value for cache", "query", name, "error", err)
		return CachedData{}
	}

	var typeStr string
	switch value.(type) {
	case model.Vector:
		typeStr = "vector"
	case model.Matrix:
		typeStr = "matrix"
	case *model.Scalar:
		typeStr = "scalar"
	case *model.String:
		typeStr = "string"
	default:
		if value != nil {
			typeStr = value.Type().String()
		} else {
			typeStr = "unknown"
		}
	}

	return CachedData{
		Name:          name,
		ValueType:     typeStr,
		JSONBytes:     jsonBytes,
		Timestamp:     time.Now(),
		RequireAuth:   config.RequireAuth,
		RequiredGroup: config.RequiredGroup,
	}
}
