package middlewares

import (
	"homelab-dashboard/internal/data"

	"github.com/prometheus/common/model"
)

type CacheProvider interface {
	Get(ctx *AppContext, queryName string) (data.CachedData, bool)
	ListAll(ctx *AppContext) []string
	Set(ctx *AppContext, queryName string, value model.Value, requireAuth bool, requiredGroup string)
	Delete(ctx *AppContext, query string)
	Size(ctx *AppContext) int
	EstimateSize(ctx *AppContext) (int, error)
}
