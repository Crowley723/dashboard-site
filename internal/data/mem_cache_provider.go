package data

import (
	"context"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/metrics"
	"log/slog"
	"sync"
)

func NewMemCache(cfg *config.Config, logger *slog.Logger) (*MemCache, error) {
	return &MemCache{
		cache:  make(map[string]CachedData),
		logger: logger,
	}, nil
}

type MemCache struct {
	cache  map[string]CachedData
	mutex  sync.RWMutex
	logger *slog.Logger
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
