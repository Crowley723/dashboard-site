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
	"github.com/prometheus/common/model"
	"github.com/redis/go-redis/extra/redisprometheus/v9"
	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Keys(ctx context.Context, pattern string) *redis.StringSliceCmd
	Ping(ctx context.Context) *redis.StatusCmd
	PoolStats() *redis.PoolStats
	Close() error
}

type RedisCache struct {
	client RedisClient
	logger *slog.Logger
}

// NewRedisCache creates a new Redis-backed cache
func NewRedisCache(cfg *config.Config, logger *slog.Logger) (*RedisCache, error) {
	var client RedisClient

	if cfg.Redis.Sentinel != nil {
		logger.Info("connecting to redis via sentinel",
			"master", cfg.Redis.Sentinel.MasterName,
			"sentinels", cfg.Redis.Sentinel.SentinelAddresses)

		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       cfg.Redis.Sentinel.MasterName,
			SentinelAddrs:    cfg.Redis.Sentinel.SentinelAddresses,
			SentinelUsername: cfg.Redis.Sentinel.SentinelUsername,
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

// key generates a namespaced Redis key
func (r *RedisCache) key(queryName string) string {
	return fmt.Sprintf("cache:query:%s", queryName)
}

// ClosePool closes the Redis connection pool
func (r *RedisCache) ClosePool() error {
	return r.client.Close()
}

func (r *RedisCache) Get(ctx context.Context, queryName string) (CachedData, bool) {
	timer := prometheus.NewTimer(metrics.CacheOperationDuration.WithLabelValues(metrics.CachetypeRedis, metrics.CacheoperationtypeGet))
	defer timer.ObserveDuration()

	data, err := r.client.Get(ctx, r.key(queryName)).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			r.logger.Error("error executing redis GET", "error", err)
		}
		metrics.CacheMisses.WithLabelValues(metrics.CachetypeRedis).Inc()
		return CachedData{}, false
	}

	metrics.CacheHits.WithLabelValues(metrics.CachetypeRedis).Inc()

	var cached CachedData
	if err := json.Unmarshal([]byte(data), &cached); err != nil {
		r.logger.Error("error unmarshalling redis response", "error", err)
		return CachedData{}, false
	}

	// Reconstruct the correct type from JSON
	switch cached.ValueType {
	case "vector":
		var v model.Vector
		if err := json.Unmarshal([]byte(cached.ValueJSON), &v); err != nil {
			r.logger.Error("error unmarshalling vector", "error", err)
			return CachedData{}, false
		}
		cached.Value = v
	case "matrix":
		var m model.Matrix
		if err := json.Unmarshal([]byte(cached.ValueJSON), &m); err != nil {
			r.logger.Error("error unmarshalling matrix", "error", err)
			return CachedData{}, false
		}
		cached.Value = m
	case "scalar":
		var s model.Scalar
		if err := json.Unmarshal([]byte(cached.ValueJSON), &s); err != nil {
			r.logger.Error("error unmarshalling scalar", "error", err)
			return CachedData{}, false
		}
		cached.Value = &s
	case "string":
		var s model.String
		if err := json.Unmarshal([]byte(cached.ValueJSON), &s); err != nil {
			r.logger.Error("error unmarshalling string", "error", err)
			return CachedData{}, false
		}
		cached.Value = &s
	default:
		r.logger.Error("unknown value type in cache", "type", cached.ValueType)
		return CachedData{}, false
	}

	cached.JSONBytes = []byte(cached.ValueJSON)

	return cached, true
}

func (r *RedisCache) ListAll(ctx context.Context) []string {
	timer := prometheus.NewTimer(metrics.CacheOperationDuration.WithLabelValues(metrics.CachetypeRedis, metrics.CacheoperationtypeListall))
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

func (r *RedisCache) Set(ctx context.Context, queryName string, value model.Value, requireAuth bool, requiredGroup string) {
	timer := prometheus.NewTimer(metrics.CacheOperationDuration.WithLabelValues(metrics.CachetypeRedis, metrics.CacheoperationtypeSet))
	defer timer.ObserveDuration()

	var valueType string
	switch value.(type) {
	case model.Vector:
		valueType = "vector"
	case model.Matrix:
		valueType = "matrix"
	case *model.Scalar:
		valueType = "scalar"
	case *model.String:
		valueType = "string"
	default:
		r.logger.Error("unknown prometheus value type", "type", fmt.Sprintf("%T", value))
		return
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		r.logger.Error("error marshalling prometheus value", "error", err)
		return
	}

	cache := CachedData{
		Value:         value,
		ValueJSON:     string(valueJSON),
		ValueType:     valueType,
		JSONBytes:     valueJSON,
		Timestamp:     time.Now(),
		Name:          queryName,
		RequireAuth:   requireAuth,
		RequiredGroup: requiredGroup,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		r.logger.Error("error marshalling cached data", "error", err)
		return
	}

	_, err = r.client.Set(ctx, r.key(queryName), data, 0).Result()
	if err != nil {
		r.logger.Error("error executing redis 'SET'", "error", err)
		return
	}
}

// Delete removes an entry from the cache
func (r *RedisCache) Delete(ctx context.Context, query string) {
	timer := prometheus.NewTimer(metrics.CacheOperationDuration.WithLabelValues(metrics.CachetypeRedis, metrics.CacheoperationtypeDelete))
	defer timer.ObserveDuration()

	_, err := r.client.Del(ctx, r.key(query)).Result()
	if err != nil {
		r.logger.Error("error executing redis 'DEL'", "error", err)
		return
	}
}

// Size returns the current number of elements in the cache
func (r *RedisCache) Size(ctx context.Context) int {
	timer := prometheus.NewTimer(metrics.CacheOperationDuration.WithLabelValues(metrics.CachetypeRedis, metrics.CacheoperationtypeCountEntries))
	defer timer.ObserveDuration()

	pattern := r.key("*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.logger.Error("error executing redis 'KEYS'", "error", err)
		return 0
	}

	return len(keys)
}
