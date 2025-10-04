package data

import (
	"homelab-dashboard/internal/config"
	"log/slog"

	"github.com/prometheus/common/model"
)

//go:generate mockgen -source=cache_provider.go -destination=../mocks/cache.go -package=mocks

type CacheProvider interface {
	Get(queryName string) (CachedData, bool)
	ListAll() []string
	Set(queryName string, value model.Value, requireAuth bool, requiredGroup string)
	Delete(query string)
	Size() int
	EstimateSize() (int, error)
}

// NewCacheProvider returns a new CacheProvider
func NewCacheProvider(config *config.Config, logger *slog.Logger) CacheProvider {
	switch config.Cache.Type {
	case "redis":
		return NewRedisCache(config, logger)
	case "memory":
		fallthrough
	default:
		return NewMemCache(config, logger)
	}
}
