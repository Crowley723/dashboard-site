package data

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/metrics"
	"log/slog"
	"time"

	"encoding/json"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/extra/redisprometheus/v9"
	"github.com/redis/go-redis/v9"
)

type RedisCacheClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Keys(ctx context.Context, pattern string) *redis.StringSliceCmd
	Ping(ctx context.Context) *redis.StatusCmd
	PoolStats() *redis.PoolStats
	Close() error
}

type RedisCache struct {
	client RedisCacheClient
	logger *slog.Logger
}

// NewRedisCache creates a new Redis-backed cache
func NewRedisCache(cfg *config.Config, logger *slog.Logger) (*RedisCache, error) {
	var client RedisCacheClient

	if cfg.Redis.Sentinel != nil {
		logger.Info("connecting to redis via sentinel",
			"master", cfg.Redis.Sentinel.MasterName,
			"sentinels", cfg.Redis.Sentinel.SentinelAddresses)

		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       cfg.Redis.Sentinel.MasterName,
			SentinelAddrs:    cfg.Redis.Sentinel.SentinelAddresses,
			SentinelPassword: cfg.Redis.Sentinel.SentinelPassword,
			Password:         cfg.Redis.Password,
			DB:               cfg.Redis.CacheIndex,
			MinIdleConns:     2,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:         cfg.Redis.Address,
			Password:     cfg.Redis.Password,
			DB:           cfg.Redis.CacheIndex,
			MinIdleConns: 2,
		})
	}

	if cfg.Server.Debug != nil && cfg.Server.Debug.Enabled {
		collector := redisprometheus.NewCollector(metrics.Namespace, "cache", client)
		if err := prometheus.Register(collector); err != nil {
			logger.Debug("failed to register redis cache collector: already registered", "error", err)
		}
	}

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

// Key generates a namespaced Redis Key
func (r *RedisCache) key(queryName string) string {
	return fmt.Sprintf("cache:query:%s", queryName)
}

// ClosePool closes the Redis connection pool
func (r *RedisCache) ClosePool() error {
	return r.client.Close()
}

func (r *RedisCache) Get(ctx context.Context, queryName string) (CachedData, bool) {
	timer := prometheus.NewTimer(metrics.CacheOperationDuration.WithLabelValues(metrics.CacheTypeRedis, metrics.CacheOperationTypeGet))
	defer timer.ObserveDuration()

	data, err := r.client.Get(ctx, r.key(queryName)).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			r.logger.Error("error executing redis GET", "error", err)
		}
		metrics.CacheMisses.WithLabelValues(metrics.CacheTypeRedis).Inc()
		return CachedData{}, false
	}

	metrics.CacheHits.WithLabelValues(metrics.CacheTypeRedis).Inc()

	var cached CachedData
	if err := json.Unmarshal([]byte(data), &cached); err != nil {
		r.logger.Error("error unmarshalling cached data", "error", err)
		return CachedData{}, false
	}

	return cached, true
}

func (r *RedisCache) ListAll(ctx context.Context) []string {
	timer := prometheus.NewTimer(metrics.CacheOperationDuration.WithLabelValues(metrics.CacheTypeRedis, metrics.CacheOperationTypeListAll))
	defer timer.ObserveDuration()

	keys, err := r.client.Keys(ctx, r.key("*")).Result()
	if err != nil {
		r.logger.Error("error executing redis 'KEYS'", "error", err)
		return []string{}
	}

	prefix := r.key("")
	prefixLen := len(prefix)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		if len(key) > prefixLen {
			result = append(result, key[prefixLen:])
		}
	}

	return result
}

func (r *RedisCache) Set(ctx context.Context, queryName string, data CachedData) {
	timer := prometheus.NewTimer(metrics.CacheOperationDuration.WithLabelValues(metrics.CacheTypeRedis, metrics.CacheOperationTypeSet))
	defer timer.ObserveDuration()

	jsonData, err := json.Marshal(data)
	if err != nil {
		r.logger.Error("error marshalling cached data", "error", err)
		return
	}

	if err := r.client.Set(ctx, r.key(queryName), jsonData, 0).Err(); err != nil {
		r.logger.Error("error executing redis 'SET'", "error", err)
		return
	}
}

// Delete removes an entry from the cache
func (r *RedisCache) Delete(ctx context.Context, query string) {
	timer := prometheus.NewTimer(metrics.CacheOperationDuration.WithLabelValues(metrics.CacheTypeRedis, metrics.CacheOperationTypeDelete))
	defer timer.ObserveDuration()

	_, err := r.client.Del(ctx, r.key(query)).Result()
	if err != nil {
		r.logger.Error("error executing redis 'DEL'", "error", err)
		return
	}
}

// Size returns the current number of elements in the cache
func (r *RedisCache) Size(ctx context.Context) int {
	timer := prometheus.NewTimer(metrics.CacheOperationDuration.WithLabelValues(metrics.CacheTypeRedis, metrics.CacheOperationTypeCountEntries))
	defer timer.ObserveDuration()

	pattern := r.key("*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.logger.Error("error executing redis 'KEYS'", "error", err)
		return 0
	}

	return len(keys)
}
