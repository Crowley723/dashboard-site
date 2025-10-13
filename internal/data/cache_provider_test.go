package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"homelab-dashboard/internal/config"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create test data
func createTestVector(size int) model.Vector {
	vec := make(model.Vector, size)
	for i := 0; i < size; i++ {
		vec[i] = &model.Sample{
			Metric: model.Metric{
				"__name__":  "test_metric",
				"instance":  model.LabelValue(fmt.Sprintf("instance-%d", i)),
				"job":       "benchmark",
				"namespace": "default",
			},
			Value:     model.SampleValue(float64(i) * 1.5),
			Timestamp: model.Time(1234567890000 + int64(i)*1000),
		}
	}
	return vec
}

func createTestMatrix(rows, cols int) model.Matrix {
	matrix := make(model.Matrix, rows)
	for i := 0; i < rows; i++ {
		pairs := make([]model.SamplePair, cols)
		for j := 0; j < cols; j++ {
			pairs[j] = model.SamplePair{
				Timestamp: model.Time(1234567890000 + int64(j)*60000),
				Value:     model.SampleValue(float64(i*j) * 0.5),
			}
		}
		matrix[i] = &model.SampleStream{
			Metric: model.Metric{
				"__name__": "test_metric",
				"series":   model.LabelValue(fmt.Sprintf("series-%d", i)),
			},
			Values: pairs,
		}
	}
	return matrix
}

// Setup helpers
func setupMemCache() *MemCache {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cache, _ := NewMemCache(&config.Config{}, logger)
	return cache
}

func setupRedisCache(tb testing.TB) *RedisCache {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{
		Redis: &config.RedisConfig{
			Address:    "127.0.0.1:6379",
			Password:   "",
			CacheIndex: 1,
		},
	}
	cache, err := NewRedisCache(cfg, logger)
	if err != nil {
		tb.Skipf("Redis not available: %v", err)
	}

	// Clean up any existing test data
	ctx := context.Background()
	err = cache.client.Ping(ctx).Err()
	if err != nil {
		tb.Skipf("Redis not available: %v", err)
	}

	tb.Cleanup(func() {
		cache.ClosePool()
	})

	return cache
}

// Unit Tests

func TestNewCacheProvider(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := []struct {
		name         string
		cacheType    string
		expectedType string
	}{
		{
			name:         "memory cache by default",
			cacheType:    "",
			expectedType: "*data.MemCache",
		},
		{
			name:         "memory cache explicitly",
			cacheType:    "memory",
			expectedType: "*data.MemCache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Cache: config.CacheConfig{
					Type: tt.cacheType,
				},
			}

			cache, err := NewCacheProvider(cfg, logger)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, fmt.Sprintf("%T", cache))
		})
	}
}

func TestMemCache_SetAndGet(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()

	tests := []struct {
		name          string
		key           string
		data          interface{}
		requireAuth   bool
		requiredGroup string
	}{
		{
			name:          "vector without auth",
			key:           "test-vector",
			data:          createTestVector(3),
			requireAuth:   false,
			requiredGroup: "",
		},
		{
			name:          "matrix with auth",
			key:           "test-matrix",
			data:          createTestMatrix(2, 3),
			requireAuth:   true,
			requiredGroup: "admin",
		},
		{
			name: "scalar",
			key:  "test-scalar",
			data: &model.Scalar{
				Value:     123.45,
				Timestamp: model.Time(time.Now().UnixMilli()),
			},
			requireAuth:   false,
			requiredGroup: "",
		},
		{
			name: "string",
			key:  "test-string",
			data: &model.String{
				Value:     "test-value",
				Timestamp: model.Time(time.Now().UnixMilli()),
			},
			requireAuth:   false,
			requiredGroup: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.data)
			require.NoError(t, err)

			cachedData := CachedData{
				Name:          tt.key,
				JSONBytes:     jsonBytes,
				RequireAuth:   tt.requireAuth,
				RequiredGroup: tt.requiredGroup,
			}

			cache.Set(ctx, tt.key, cachedData)

			result, found := cache.Get(ctx, tt.key)
			assert.True(t, found)
			assert.Equal(t, tt.key, result.Name)
			assert.Equal(t, jsonBytes, result.JSONBytes)
			assert.Equal(t, tt.requireAuth, result.RequireAuth)
			assert.Equal(t, tt.requiredGroup, result.RequiredGroup)
			assert.NotNil(t, result.JSONBytes)
		})
	}
}

func TestMemCache_GetMiss(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()

	result, found := cache.Get(ctx, "nonexistent")
	assert.False(t, found)
	assert.Empty(t, result.Name)
	assert.Nil(t, result.JSONBytes)
}

func TestMemCache_Delete(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()

	vec := createTestVector(2)
	jsonBytes, _ := json.Marshal(vec)
	data := CachedData{
		Name:      "test-delete",
		JSONBytes: jsonBytes,
	}

	cache.Set(ctx, "test-delete", data)

	// Verify it exists
	_, found := cache.Get(ctx, "test-delete")
	assert.True(t, found)

	// Delete it
	cache.Delete(ctx, "test-delete")

	// Verify it's gone
	_, found = cache.Get(ctx, "test-delete")
	assert.False(t, found)
}

func TestMemCache_ListAll(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()

	// Start with empty cache
	keys := cache.ListAll(ctx)
	assert.Empty(t, keys)

	// Add some items
	vec := createTestVector(1)
	jsonBytes, _ := json.Marshal(vec)
	data := CachedData{
		Name:      "",
		JSONBytes: jsonBytes,
	}

	for i := 1; i <= 3; i++ {
		data.Name = fmt.Sprintf("query%d", i)
		cache.Set(ctx, data.Name, data)
	}

	keys = cache.ListAll(ctx)
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "query1")
	assert.Contains(t, keys, "query2")
	assert.Contains(t, keys, "query3")
}

func TestMemCache_Size(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()

	// Start with empty cache
	assert.Equal(t, 0, cache.Size(ctx))

	// Add items
	vec := createTestVector(1)
	jsonBytes, _ := json.Marshal(vec)
	data := CachedData{
		JSONBytes: jsonBytes,
	}

	data.Name = "item1"
	cache.Set(ctx, "item1", data)
	assert.Equal(t, 1, cache.Size(ctx))

	data.Name = "item2"
	cache.Set(ctx, "item2", data)
	assert.Equal(t, 2, cache.Size(ctx))

	// Delete an item
	cache.Delete(ctx, "item1")
	assert.Equal(t, 1, cache.Size(ctx))
}

func TestMemCache_ConcurrentAccess(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()
	vec := createTestVector(10)
	jsonBytes, _ := json.Marshal(vec)

	// Test concurrent writes and reads
	done := make(chan bool, 10)

	// Start 5 writers
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("writer-%d-item-%d", id, j)
				data := CachedData{
					Name:      key,
					JSONBytes: jsonBytes,
				}
				cache.Set(ctx, key, data)
			}
			done <- true
		}(i)
	}

	// Start 5 readers
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				cache.Get(ctx, fmt.Sprintf("reader-item-%d", j))
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	assert.Equal(t, 50, cache.Size(ctx)) // 5 writers * 10 items each
}

func TestMemCache_UpdateExistingKey(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()

	// Set initial value
	vec1 := createTestVector(2)
	jsonBytes1, _ := json.Marshal(vec1)
	data1 := CachedData{
		Name:      "update-test",
		JSONBytes: jsonBytes1,
	}
	cache.Set(ctx, "update-test", data1)

	result, found := cache.Get(ctx, "update-test")
	assert.True(t, found)
	assert.Equal(t, jsonBytes1, result.JSONBytes)
	assert.False(t, result.RequireAuth)

	// Update with new value and auth requirements
	vec2 := createTestVector(5)
	jsonBytes2, _ := json.Marshal(vec2)
	data2 := CachedData{
		Name:          "update-test",
		JSONBytes:     jsonBytes2,
		RequireAuth:   true,
		RequiredGroup: "admin",
	}
	cache.Set(ctx, "update-test", data2)

	result, found = cache.Get(ctx, "update-test")
	assert.True(t, found)
	assert.Equal(t, jsonBytes2, result.JSONBytes)
	assert.True(t, result.RequireAuth)
	assert.Equal(t, "admin", result.RequiredGroup)

	// Size should still be 1
	assert.Equal(t, 1, cache.Size(ctx))
}

func TestMemCache_TimestampUpdates(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()

	vec := createTestVector(1)
	jsonBytes, _ := json.Marshal(vec)
	data := CachedData{
		Name:      "timestamp-test",
		JSONBytes: jsonBytes,
		Timestamp: time.Now(),
	}

	beforeSet := time.Now()
	cache.Set(ctx, "timestamp-test", data)
	afterSet := time.Now()

	result, found := cache.Get(ctx, "timestamp-test")
	require.True(t, found, "cache entry should exist")

	assert.False(t, result.Timestamp.IsZero(), "timestamp should not be zero value")

	assert.WithinDuration(t, beforeSet, result.Timestamp, time.Second,
		"timestamp should be set around the time Set() was called")
	assert.True(t, !result.Timestamp.After(afterSet),
		"timestamp should not be after Set() completed")
}

// Benchmark Set operations
func BenchmarkCache_Set(b *testing.B) {
	tests := []struct {
		name      string
		setupFunc func(testing.TB) CacheProvider
		dataFunc  func() (string, []byte)
	}{
		{
			name:      "MemCache_SmallVector",
			setupFunc: func(tb testing.TB) CacheProvider { return setupMemCache() },
			dataFunc: func() (string, []byte) {
				vec := createTestVector(10)
				jsonBytes, _ := json.Marshal(vec)
				return "query", jsonBytes
			},
		},
		{
			name:      "RedisCache_SmallVector",
			setupFunc: func(tb testing.TB) CacheProvider { return setupRedisCache(tb) },
			dataFunc: func() (string, []byte) {
				vec := createTestVector(10)
				jsonBytes, _ := json.Marshal(vec)
				return "query", jsonBytes
			},
		},
		{
			name:      "MemCache_LargeVector",
			setupFunc: func(tb testing.TB) CacheProvider { return setupMemCache() },
			dataFunc: func() (string, []byte) {
				vec := createTestVector(1000)
				jsonBytes, _ := json.Marshal(vec)
				return "query", jsonBytes
			},
		},
		{
			name:      "RedisCache_LargeVector",
			setupFunc: func(tb testing.TB) CacheProvider { return setupRedisCache(tb) },
			dataFunc: func() (string, []byte) {
				vec := createTestVector(1000)
				jsonBytes, _ := json.Marshal(vec)
				return "query", jsonBytes
			},
		},
		{
			name:      "MemCache_Matrix",
			setupFunc: func(tb testing.TB) CacheProvider { return setupMemCache() },
			dataFunc: func() (string, []byte) {
				matrix := createTestMatrix(50, 100)
				jsonBytes, _ := json.Marshal(matrix)
				return "query", jsonBytes
			},
		},
		{
			name:      "RedisCache_Matrix",
			setupFunc: func(tb testing.TB) CacheProvider { return setupRedisCache(tb) },
			dataFunc: func() (string, []byte) {
				matrix := createTestMatrix(50, 100)
				jsonBytes, _ := json.Marshal(matrix)
				return "query", jsonBytes
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cache := tt.setupFunc(b)
			_, jsonBytes := tt.dataFunc()
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("query-%d", i%100)
				data := CachedData{
					Name:      key,
					JSONBytes: jsonBytes,
				}
				cache.Set(ctx, key, data)
			}
		})
	}
}

// Benchmark Get operations
func BenchmarkCache_Get(b *testing.B) {
	tests := []struct {
		name      string
		setupFunc func(testing.TB) CacheProvider
	}{
		{
			name:      "MemCache",
			setupFunc: func(tb testing.TB) CacheProvider { return setupMemCache() },
		},
		{
			name:      "RedisCache",
			setupFunc: func(tb testing.TB) CacheProvider { return setupRedisCache(tb) },
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cache := tt.setupFunc(b)
			vec := createTestVector(100)
			jsonBytes, _ := json.Marshal(vec)
			ctx := context.Background()

			for i := 0; i < 100; i++ {
				key := fmt.Sprintf("query-%d", i)
				data := CachedData{
					Name:      key,
					JSONBytes: jsonBytes,
				}
				cache.Set(ctx, key, data)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cache.Get(ctx, fmt.Sprintf("query-%d", i%100))
			}
		})
	}
}

// Benchmark Get misses
func BenchmarkCache_GetMiss(b *testing.B) {
	tests := []struct {
		name      string
		setupFunc func(testing.TB) CacheProvider
	}{
		{
			name:      "MemCache",
			setupFunc: func(tb testing.TB) CacheProvider { return setupMemCache() },
		},
		{
			name:      "RedisCache",
			setupFunc: func(tb testing.TB) CacheProvider { return setupRedisCache(tb) },
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cache := tt.setupFunc(b)
			ctx := context.Background()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cache.Get(ctx, fmt.Sprintf("nonexistent-%d", i))
			}
		})
	}
}

// Benchmark Delete operations
func BenchmarkCache_Delete(b *testing.B) {
	tests := []struct {
		name      string
		setupFunc func(testing.TB) CacheProvider
	}{
		{
			name:      "MemCache",
			setupFunc: func(tb testing.TB) CacheProvider { return setupMemCache() },
		},
		{
			name:      "RedisCache",
			setupFunc: func(tb testing.TB) CacheProvider { return setupRedisCache(tb) },
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cache := tt.setupFunc(b)
			vec := createTestVector(10)
			jsonBytes, _ := json.Marshal(vec)
			ctx := context.Background()

			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("query-%d", i)
				data := CachedData{
					Name:      key,
					JSONBytes: jsonBytes,
				}
				cache.Set(ctx, key, data)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cache.Delete(ctx, fmt.Sprintf("query-%d", i))
			}
		})
	}
}

// Benchmark ListAll
func BenchmarkCache_ListAll(b *testing.B) {
	tests := []struct {
		name      string
		setupFunc func(testing.TB) CacheProvider
		numItems  int
	}{
		{
			name:      "MemCache_10",
			setupFunc: func(tb testing.TB) CacheProvider { return setupMemCache() },
			numItems:  10,
		},
		{
			name:      "RedisCache_10",
			setupFunc: func(tb testing.TB) CacheProvider { return setupRedisCache(tb) },
			numItems:  10,
		},
		{
			name:      "MemCache_100",
			setupFunc: func(tb testing.TB) CacheProvider { return setupMemCache() },
			numItems:  100,
		},
		{
			name:      "RedisCache_100",
			setupFunc: func(tb testing.TB) CacheProvider { return setupRedisCache(tb) },
			numItems:  100,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cache := tt.setupFunc(b)
			vec := createTestVector(10)
			jsonBytes, _ := json.Marshal(vec)
			ctx := context.Background()

			for i := 0; i < tt.numItems; i++ {
				key := fmt.Sprintf("query-%d", i)
				data := CachedData{
					Name:      key,
					JSONBytes: jsonBytes,
				}
				cache.Set(ctx, key, data)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cache.ListAll(ctx)
			}
		})
	}
}

// Benchmark concurrent operations
func BenchmarkCache_ConcurrentReadWrite(b *testing.B) {
	tests := []struct {
		name      string
		setupFunc func(testing.TB) CacheProvider
	}{
		{
			name:      "MemCache",
			setupFunc: func(tb testing.TB) CacheProvider { return setupMemCache() },
		},
		{
			name:      "RedisCache",
			setupFunc: func(tb testing.TB) CacheProvider { return setupRedisCache(tb) },
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cache := tt.setupFunc(b)
			vec := createTestVector(50)
			jsonBytes, _ := json.Marshal(vec)
			ctx := context.Background()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := fmt.Sprintf("query-%d", i%100)
					if i%2 == 0 {
						data := CachedData{
							Name:      key,
							JSONBytes: jsonBytes,
						}
						cache.Set(ctx, key, data)
					} else {
						cache.Get(ctx, key)
					}
					i++
				}
			})
		})
	}
}

// Benchmark Size operations
func BenchmarkCache_Size(b *testing.B) {
	tests := []struct {
		name      string
		setupFunc func(testing.TB) CacheProvider
	}{
		{
			name:      "MemCache",
			setupFunc: func(tb testing.TB) CacheProvider { return setupMemCache() },
		},
		{
			name:      "RedisCache",
			setupFunc: func(tb testing.TB) CacheProvider { return setupRedisCache(tb) },
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cache := tt.setupFunc(b)
			vec := createTestVector(10)
			jsonBytes, _ := json.Marshal(vec)
			ctx := context.Background()

			for i := 0; i < 50; i++ {
				key := fmt.Sprintf("query-%d", i)
				data := CachedData{
					Name:      key,
					JSONBytes: jsonBytes,
				}
				cache.Set(ctx, key, data)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cache.Size(ctx)
			}
		})
	}
}
