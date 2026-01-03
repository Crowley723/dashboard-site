package data

import (
	"context"
	"errors"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/metrics"
	"log/slog"
	"sync"
	"time"
)

func NewMemCache(cfg *config.Config, logger *slog.Logger) (*MemCache, error) {
	return &MemCache{
		cache:   make(map[string]CachedData),
		kvStore: make(map[string]*kvEntry),
		logger:  logger,
	}, nil
}

type kvEntry struct {
	value     string
	expiresAt time.Time
}

type MemCache struct {
	cache   map[string]CachedData
	kvStore map[string]*kvEntry
	mutex   sync.RWMutex
	logger  *slog.Logger
}

// Get returns the data for a currently cached query
func (d *MemCache) Get(ctx context.Context, queryName string) (CachedData, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if cached, exists := d.cache[queryName]; exists {
		metrics.CacheHits.WithLabelValues("memcache").Inc()
		return cached, true
	}

	metrics.CacheMisses.WithLabelValues("memcache").Inc()
	return CachedData{}, false
}

// ListAll returns a slice of keys for the currently cached queries
func (d *MemCache) ListAll(ctx context.Context) []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	keys := make([]string, 0, len(d.cache))
	for k := range d.cache {
		keys = append(keys, k)
	}

	return keys
}

// Set sets (or inserts) the value of a query
func (d *MemCache) Set(ctx context.Context, queryName string, data CachedData) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.cache[queryName] = data
}

// Delete removes an entry from the cache
func (d *MemCache) Delete(ctx context.Context, query string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.cache, query)
}

// Size returns the current number of elements in the cache
func (d *MemCache) Size(ctx context.Context) int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return len(d.cache)
}

// GetKey gets a generic KV entry
func (d *MemCache) GetKey(ctx context.Context, key string) (string, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	entry, exists := d.kvStore[key]
	if !exists {
		return "", errors.New("key not found")
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return "", errors.New("key expired")
	}

	return entry.value, nil
}

// SetKey sets a generic KV entry with TTL
func (d *MemCache) SetKey(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return errors.New("value must be string or []byte")
	}

	entry := &kvEntry{
		value: strValue,
	}

	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)

		time.AfterFunc(ttl, func() {
			d.mutex.Lock()
			defer d.mutex.Unlock()
			delete(d.kvStore, key)
		})
	}

	d.kvStore[key] = entry
	return nil
}

// GetDelKey gets and deletes a key atomically
func (d *MemCache) GetDelKey(ctx context.Context, key string) (string, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	entry, exists := d.kvStore[key]
	if !exists {
		return "", errors.New("key not found")
	}

	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		delete(d.kvStore, key)
		return "", errors.New("key expired")
	}

	value := entry.value
	delete(d.kvStore, key)
	return value, nil
}
