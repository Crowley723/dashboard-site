package data

import (
	"context"
	"homelab-dashboard/internal/config"
	"log/slog"
	"time"
)

//go:generate mockgen -source=cache_provider.go -destination=../mocks/cache.go -package=mocks

type CacheProvider interface {
	Get(ctx context.Context, queryName string) (CachedData, bool)
	ListAll(ctx context.Context) []string
	Set(ctx context.Context, queryName string, data CachedData)
	Delete(ctx context.Context, query string)
	Size(ctx context.Context) int
	GetKey(ctx context.Context, key string) (string, error)
	SetKey(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	GetDelKey(ctx context.Context, key string) (string, error)
}

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
