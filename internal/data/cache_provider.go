package data

import (
	"github.com/prometheus/common/model"
)

//go:generate mockgen -source=cache_provider.go -destination=../mocks/cache.go -package=mocks

type CacheProvider interface {
	Get(queryName string) (CachedData, bool)
	ListAll() []string
	Set(queryName string, value model.Value, requireAuth bool, requiredGroup string)
	Delete(query string)
	Size() int
	EstimateSize() (int, error)
}
