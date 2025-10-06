package data

import (
	"context"
	"time"

	"github.com/prometheus/common/model"
)

type QueryResult struct {
	Name      string      `json:"name"`
	Value     model.Value `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

type DashboardData struct {
	LastUpdated time.Time     `json:"last_updated"`
	Queries     []QueryResult `json:"queries"`
}

// CachedData represents a cache entry of the data for a single query.
type CachedData struct {
	Name          string      `json:"name"`
	Value         model.Value `json:"-"`          //for memcache use
	ValueJSON     string      `json:"value_json"` // raw JSON for Redis
	ValueType     string      `json:"value_type"` // "vector", "matrix", "scalar", "string"
	JSONBytes     []byte      `json:"json_bytes"`
	Timestamp     time.Time   `json:"timestamp"`
	RequireAuth   bool        `json:"require_auth"`
	RequiredGroup string      `json:"required_group"`
}

type CacheProvider interface {
	Get(ctx context.Context, queryName string) (CachedData, bool)
	ListAll(ctx context.Context) []string
	Set(ctx context.Context, queryName string, value model.Value, requireAuth bool, requiredGroup string)
	Delete(ctx context.Context, query string)
	Size(ctx context.Context) int
}
