package data

import (
	"homelab-dashboard/internal/config"
	"log/slog"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4/json"
	"github.com/prometheus/common/model"
)

func NewMemCache(cfg *config.Config, logger *slog.Logger) *MemCache {
	return &MemCache{
		cache:  make(map[string]CachedData),
		logger: logger,
	}
}

type MemCache struct {
	cache  map[string]CachedData
	mutex  sync.RWMutex
	logger *slog.Logger
}

// Get returns the data for a currently cached query
func (d *MemCache) Get(queryName string) (CachedData, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if cached, exists := d.cache[queryName]; exists {
		return cached, true
	}

	return CachedData{}, false
}

// ListAll returns a slice of keys for the currently cached queries
func (d *MemCache) ListAll() []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	keys := make([]string, 0, len(d.cache))
	for k := range d.cache {
		keys = append(keys, k)
	}

	return keys
}

// Set sets (or inserts) the value of a query
func (d *MemCache) Set(queryName string, value model.Value, requireAuth bool, requiredGroup string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.cache[queryName] = CachedData{
		Value:         value,
		Timestamp:     time.Now(),
		Name:          queryName,
		RequireAuth:   requireAuth,
		RequiredGroup: requiredGroup,
	}
}

// Delete removes an entry from the cache
func (d *MemCache) Delete(query string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.cache, query)
}

// Size returns the current number of elements in the cache
func (d *MemCache) Size() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return len(d.cache)
}

// EstimateSize returns the estimated size of the current cache (in bytes) by checking the length of the marshalled cache.
func (d *MemCache) EstimateSize() (int, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	data, err := json.Marshal(d.cache)
	if err != nil {
		return 0, err
	}

	return len(data), nil
}
