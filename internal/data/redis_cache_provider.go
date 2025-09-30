package data

import (
	"fmt"
	"homelab-dashboard/internal/config"
	"log/slog"
	"time"

	"encoding/json"

	"github.com/gomodule/redigo/redis"
	"github.com/prometheus/common/model"
)

type RedisCache struct {
	pool   *redis.Pool
	logger *slog.Logger
}

// NewRedisCache creates a new Redis-backed cache
func NewRedisCache(cfg *config.Config, logger *slog.Logger) *RedisCache {
	pool := &redis.Pool{
		MaxIdle:     10,
		MaxActive:   50,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			opts := []redis.DialOption{
				redis.DialDatabase(cfg.Redis.CacheIndex),
			}
			if cfg.Redis.Password != "" {
				opts = append(opts, redis.DialPassword(cfg.Redis.Password))
			}
			return redis.Dial("tcp", cfg.Redis.Address, opts...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	return &RedisCache{
		pool:   pool,
		logger: logger,
	}
}

// key generates a namespaced Redis key
func (r *RedisCache) key(queryName string) string {
	return fmt.Sprintf("cache:query:%s", queryName)
}

// ClosePool closes the Redis connection pool
func (r *RedisCache) ClosePool() error {
	return r.pool.Close()
}

func (r *RedisCache) Get(queryName string) (CachedData, bool) {
	conn := r.pool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			r.logger.Error("error closing redis connection", "error", err)
		}
	}(conn)

	data, err := redis.String(conn.Do("GET", r.key(queryName)))
	if err != nil {
		if err != redis.ErrNil {
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

func (r *RedisCache) ListAll() []string {
	conn := r.pool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			r.logger.Error("error closing redis connection", "error", err)
		}
	}(conn)

	pattern := r.key("*")
	keys, err := redis.Strings(conn.Do("KEYS", pattern+"*"))
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

func (r *RedisCache) Set(queryName string, value model.Value, requireAuth bool, requiredGroup string) {
	conn := r.pool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			r.logger.Error("error closing redis connection", "error", err)
		}
	}(conn)

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

	_, err = conn.Do("SET", r.key(queryName), data)
	if err != nil {
		r.logger.Error("error executing redis 'SET'", "error", err)
		return
	}
}

// Delete removes an entry from the cache
func (r *RedisCache) Delete(query string) {
	conn := r.pool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			r.logger.Error("error closing redis connection", "error", err)
		}
	}(conn)

	_, err := conn.Do("DEL", r.key(query))
	if err != nil {
		r.logger.Error("error executing redis 'DEL'", "error", err)
		return
	}
}

// Size returns the current number of elements in the cache
func (r *RedisCache) Size() int {
	conn := r.pool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			r.logger.Error("error closing redis connection", "error", err)
		}
	}(conn)

	pattern := r.key("*")
	keys, err := redis.Values(conn.Do("KEYS", pattern))
	if err != nil {
		r.logger.Error("error executing redis 'KEYS'", "error", err)
		return 0
	}

	return len(keys)
}

// EstimateSize returns the estimated size of the current cache (in bytes)
func (r *RedisCache) EstimateSize() (int, error) {
	conn := r.pool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			r.logger.Error("error closing redis connection", "error", err)
		}
	}(conn)

	pattern := r.key("*")
	keys, err := redis.Strings(conn.Do("KEYS", pattern))
	if err != nil {
		r.logger.Error("error executing redis 'KEYS'", "error", err)
		return 0, err
	}

	totalSize := 0
	for _, key := range keys {
		size, err := redis.Int(conn.Do("STRLEN", key))
		if err != nil {
			r.logger.Error("error executing redis 'STRLEN'", "error", err)
			continue
		}
		totalSize += size
	}

	return totalSize, nil
}
