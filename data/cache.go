package data

import (
	"sync"
	"time"

	"github.com/prometheus/common/model"
)

type CachedData struct {
	Value     model.Value
	Timestamp time.Time
	Name      string
	TTL       time.Duration
}

type Cache struct {
	cache map[string]CachedData
	mutex sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		cache: make(map[string]CachedData),
	}
}

func (d *Cache) Get(query string) (CachedData, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if cached, exists := d.cache[query]; exists {
		if time.Since(cached.Timestamp) < cached.TTL {
			return cached, true
		}
	}

	return CachedData{}, false
}

func (d *Cache) Set(name string, value model.Value, ttl time.Duration) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.cache[name] = CachedData{
		Value:     value,
		Timestamp: time.Now(),
		Name:      name,
		TTL:       ttl,
	}
}

func (d *Cache) Delete(query string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.cache, query)
}

func (d *Cache) Size() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return len(d.cache)
}
