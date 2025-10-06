package data

import (
	"homelab-dashboard/internal/config"
	"log/slog"
)

//go:generate mockgen -source=cache_provider.go -destination=../mocks/cache.go -package=mocks

// NewCacheProvider returns a new CacheProvider
func NewCacheProvider(config *config.Config, logger *slog.Logger) (CacheProvider, error) {
	switch config.Cache.Type {
	case "redis":
		return NewRedisCache(config, logger)
	case "memory":
		fallthrough
	default:
		return NewMemCache(config, logger)
	}
}
