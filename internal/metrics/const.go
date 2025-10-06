package metrics

const Namespace = "homelab_dashboard"

const (
	CacheTypeRedis  = "redis"
	CacheTypeMemory = "memory"
)

const (
	CacheOperationTypeGet          = "get"
	CacheOperationTypeSet          = "set"
	CacheOperationTypeListAll      = "list_all"
	CacheOperationTypeDelete       = "delete"
	CacheOperationTypeCountEntries = "count_entries"
)

const DataSourceTypeMimir = "mimir"
