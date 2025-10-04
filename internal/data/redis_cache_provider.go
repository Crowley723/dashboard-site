package data

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/middlewares"
	"log/slog"
	"time"

	"encoding/json"

	"github.com/prometheus/common/model"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	logger *slog.Logger
}

// NewRedisCache creates a new Redis-backed cache
func NewRedisCache(cfg *config.Config, logger *slog.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Address,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.CacheIndex,
		MinIdleConns: 2,
	})

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

func (r *RedisCache) Get(ctx *middlewares.AppContext, queryName string) (CachedData, bool) {
	data, err := r.client.Get(ctx, r.key(queryName)).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			r.logger.Error("error executing redis GET", "error", err)
		}
		return CachedData{}, false
	}

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

	return cached, true
}

func (r *RedisCache) ListAll(ctx *middlewares.AppContext) []string {
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

func (r *RedisCache) Set(ctx *middlewares.AppContext, queryName string, value model.Value, requireAuth bool, requiredGroup string) {
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
func (r *RedisCache) Delete(ctx *middlewares.AppContext, query string) {
	_, err := r.client.Del(ctx, r.key(query)).Result()
	if err != nil {
		r.logger.Error("error executing redis 'DEL'", "error", err)
		return
	}
}

// Size returns the current number of elements in the cache
func (r *RedisCache) Size(ctx *middlewares.AppContext) int {
	pattern := r.key("*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		r.logger.Error("error executing redis 'KEYS'", "error", err)
		return 0
	}

	return len(keys)
}

// EstimateSize returns the estimated size of the current cache (in bytes)
func (r *RedisCache) EstimateSize(ctx *middlewares.AppContext) (int, error) {
	dbSize, err := r.client.DBSize(ctx).Result()
	if err != nil {
		r.logger.Error("error executing redis 'KEYS'", "error", err)
		return 0, err
	}

	return int(dbSize), nil
}
