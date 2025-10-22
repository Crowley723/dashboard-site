package data

import (
	"context"
	"homelab-dashboard/internal/config"
	"log/slog"
	"time"
)

func NewHybridCache(cfg *config.Config, logger *slog.Logger) (*HybridCache, error) {
	memCache, err := NewMemCache(cfg, logger)
	if err != nil {
		return nil, err
	}

	redisCache, err := NewRedisCache(cfg, logger)
	if err != nil {
		return nil, err
	}

	return &HybridCache{
		mem:    memCache,
		redis:  redisCache,
		logger: logger,
	}, nil
}

type HybridCache struct {
	mem    *MemCache
	redis  *RedisCache
	logger *slog.Logger
}

// Get attempts to retrieve cached data from the memory cache and falls back to redis. If data is expired, we invalidate it but return stale data.
func (h HybridCache) Get(ctx context.Context, queryName string) (CachedData, bool) {
	var (
		data CachedData
	)

	data, ok := h.mem.Get(ctx, queryName)
	if ok {
		if time.Now().Before(data.ExpiresAt) {
			return data, true
		}

		h.Delete(ctx, queryName)
	}

	if data, ok = h.redis.Get(ctx, queryName); ok {
		if time.Now().Before(data.ExpiresAt) {
			h.mem.Set(ctx, queryName, data)
			return data, true
		}
	}

	return CachedData{}, false
}

// ListAll retrieves a list of all query keys currently available, both in memory and redis.
func (h HybridCache) ListAll(ctx context.Context) []string {
	redisKeys := h.redis.ListAll(ctx)
	memKeys := h.mem.ListAll(ctx)

	if len(redisKeys) == 0 {
		return memKeys
	}

	if len(memKeys) == 0 {
		return redisKeys
	}

	resultKeys := make(map[string]bool, len(redisKeys)+len(memKeys))

	for _, key := range redisKeys {
		resultKeys[key] = true
	}

	for _, key := range memKeys {
		resultKeys[key] = true
	}

	result := make([]string, 0, len(resultKeys))
	for key := range resultKeys {
		result = append(result, key)
	}

	return result
}

func (h HybridCache) Set(ctx context.Context, queryName string, data CachedData) {
	h.redis.Set(ctx, queryName, data)

	h.mem.Set(ctx, queryName, data)
}

func (h HybridCache) Delete(ctx context.Context, query string) {
	h.mem.Delete(ctx, query)
}

func (h HybridCache) Size(ctx context.Context) int {
	return h.redis.Size(ctx)
}
