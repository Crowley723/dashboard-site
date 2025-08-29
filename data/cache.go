package data

import (
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4/json"
	"github.com/prometheus/common/model"
)

// CachedData represents a cache entry of the data for a single query.
type CachedData struct {
	Name          string
	Value         model.Value
	Timestamp     time.Time
	RequireAuth   bool
	RequiredGroup string
}

type Cache struct {
	cache map[string]CachedData
	mutex sync.RWMutex
}

// NewCache returns a new Cache
func NewCache() *Cache {
	return &Cache{
		cache: make(map[string]CachedData),
	}
}

// Get returns the data for a currently cached query
func (d *Cache) Get(queryName string) (CachedData, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if cached, exists := d.cache[queryName]; exists {
		return cached, true
	}

	return CachedData{}, false
}

// ListAll returns a slice of keys for the currently cached queries
func (d *Cache) ListAll() []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	keys := make([]string, 0, len(d.cache))
	for k := range d.cache {
		keys = append(keys, k)
	}

	return keys
}

// Set sets (or inserts) the value of a query
func (d *Cache) Set(queryName string, value model.Value, requireAuth bool, requiredGroup string) {
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
func (d *Cache) Delete(query string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.cache, query)
}

// Size returns the current number of elements in the cache
func (d *Cache) Size() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return len(d.cache)
}

// EstimateSize returns the estimated size of the current cache (in bytes) by checking the length of the marshalled cache.
func (d *Cache) EstimateSize() (int, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	data, err := json.Marshal(d.cache)
	if err != nil {
		return 0, err
	}

	return len(data), nil
}
