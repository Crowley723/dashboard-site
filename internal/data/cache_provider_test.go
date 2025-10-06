package data

import (
	"context"
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

func setupRedisCache(b *testing.B) *RedisCache {
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
		b.Skipf("Redis not available: %v", err)
	}

	// Clean up any existing test data
	ctx := context.Background()
	err = cache.client.Ping(ctx).Err()
	if err != nil {
		b.Skipf("Redis not available: %v", err)
	}

	b.Cleanup(func() {
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
		{
			name:         "redis cache",
			cacheType:    "redis",
			expectedType: "*data.RedisCache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Cache: config.CacheConfig{
					Type: tt.cacheType,
				},
				Redis: &config.RedisConfig{
					Address:    "127.0.0.1:6379",
					CacheIndex: 1,
				},
			}

			cache, err := NewCacheProvider(cfg, logger)
			if tt.cacheType == "redis" {
				// Redis may not be available in test environment
				if err != nil {
					t.Skipf("Redis not available: %v", err)
				}
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedType, fmt.Sprintf("%T", cache))
		})
	}
}

func TestMemCache_SetAndGet(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()

	// Test vector
	vec := createTestVector(3)
	cache.Set(ctx, "test-vector", vec, false, "")

	result, found := cache.Get(ctx, "test-vector")
	assert.True(t, found)
	assert.Equal(t, "test-vector", result.Name)
	assert.Equal(t, vec, result.Value)
	assert.False(t, result.RequireAuth)
	assert.Empty(t, result.RequiredGroup)
	assert.NotNil(t, result.JSONBytes)

	// Test matrix
	matrix := createTestMatrix(2, 3)
	cache.Set(ctx, "test-matrix", matrix, true, "admin")

	result, found = cache.Get(ctx, "test-matrix")
	assert.True(t, found)
	assert.Equal(t, "test-matrix", result.Name)
	assert.Equal(t, matrix, result.Value)
	assert.True(t, result.RequireAuth)
	assert.Equal(t, "admin", result.RequiredGroup)

	// Test scalar
	scalar := &model.Scalar{
		Value:     123.45,
		Timestamp: model.Time(time.Now().UnixMilli()),
	}
	cache.Set(ctx, "test-scalar", scalar, false, "")

	result, found = cache.Get(ctx, "test-scalar")
	assert.True(t, found)
	assert.Equal(t, "test-scalar", result.Name)
	assert.Equal(t, scalar, result.Value)

	// Test string
	str := &model.String{
		Value:     "test-value",
		Timestamp: model.Time(time.Now().UnixMilli()),
	}
	cache.Set(ctx, "test-string", str, false, "")

	result, found = cache.Get(ctx, "test-string")
	assert.True(t, found)
	assert.Equal(t, "test-string", result.Name)
	assert.Equal(t, str, result.Value)
}

func TestMemCache_GetMiss(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()

	result, found := cache.Get(ctx, "nonexistent")
	assert.False(t, found)
	assert.Empty(t, result.Name)
	assert.Nil(t, result.Value)
}

func TestMemCache_Delete(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()

	vec := createTestVector(2)
	cache.Set(ctx, "test-delete", vec, false, "")

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
	cache.Set(ctx, "query1", vec, false, "")
	cache.Set(ctx, "query2", vec, false, "")
	cache.Set(ctx, "query3", vec, false, "")

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
	cache.Set(ctx, "item1", vec, false, "")
	assert.Equal(t, 1, cache.Size(ctx))

	cache.Set(ctx, "item2", vec, false, "")
	assert.Equal(t, 2, cache.Size(ctx))

	// Delete an item
	cache.Delete(ctx, "item1")
	assert.Equal(t, 1, cache.Size(ctx))
}

func TestMemCache_ConcurrentAccess(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()
	vec := createTestVector(10)

	// Test concurrent writes and reads
	done := make(chan bool, 10)

	// Start 5 writers
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				cache.Set(ctx, fmt.Sprintf("writer-%d-item-%d", id, j), vec, false, "")
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
	cache.Set(ctx, "update-test", vec1, false, "")

	result, found := cache.Get(ctx, "update-test")
	assert.True(t, found)
	assert.Equal(t, vec1, result.Value)
	assert.False(t, result.RequireAuth)

	// Update with new value and auth requirements
	vec2 := createTestVector(5)
	cache.Set(ctx, "update-test", vec2, true, "admin")

	result, found = cache.Get(ctx, "update-test")
	assert.True(t, found)
	assert.Equal(t, vec2, result.Value)
	assert.True(t, result.RequireAuth)
	assert.Equal(t, "admin", result.RequiredGroup)

	// Size should still be 1
	assert.Equal(t, 1, cache.Size(ctx))
}

func TestMemCache_TimestampUpdates(t *testing.T) {
	cache := setupMemCache()
	ctx := context.Background()

	vec := createTestVector(1)
	beforeSet := time.Now()
	cache.Set(ctx, "timestamp-test", vec, false, "")
	afterSet := time.Now()

	result, found := cache.Get(ctx, "timestamp-test")
	assert.True(t, found)
	assert.True(t, result.Timestamp.After(beforeSet) || result.Timestamp.Equal(beforeSet))
	assert.True(t, result.Timestamp.Before(afterSet) || result.Timestamp.Equal(afterSet))
}

// Benchmark Set operations
func BenchmarkMemCache_Set_SmallVector(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i%100), vec, false, "")
	}
}

func BenchmarkRedisCache_Set_SmallVector(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i%100), vec, false, "")
	}
}

func BenchmarkMemCache_Set_LargeVector(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(1000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i%100), vec, false, "")
	}
}

func BenchmarkRedisCache_Set_LargeVector(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(1000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i%100), vec, false, "")
	}
}

func BenchmarkMemCache_Set_Matrix(b *testing.B) {
	cache := setupMemCache()
	matrix := createTestMatrix(50, 100)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i%100), matrix, false, "")
	}
}

func BenchmarkRedisCache_Set_Matrix(b *testing.B) {
	cache := setupRedisCache(b)
	matrix := createTestMatrix(50, 100)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i%100), matrix, false, "")
	}
}

// Benchmark Get operations
func BenchmarkMemCache_Get(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(100)
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 100; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, fmt.Sprintf("query-%d", i%100))
	}
}

func BenchmarkRedisCache_Get(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(100)
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 100; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, fmt.Sprintf("query-%d", i%100))
	}
}

// Benchmark Get misses
func BenchmarkMemCache_GetMiss(b *testing.B) {
	cache := setupMemCache()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, fmt.Sprintf("nonexistent-%d", i))
	}
}

func BenchmarkRedisCache_GetMiss(b *testing.B) {
	cache := setupRedisCache(b)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, fmt.Sprintf("nonexistent-%d", i))
	}
}

// Benchmark Delete operations
func BenchmarkMemCache_Delete(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(10)
	ctx := context.Background()

	// Pre-populate cache with keys we'll delete
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Delete(ctx, fmt.Sprintf("query-%d", i))
	}
}

func BenchmarkRedisCache_Delete(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(10)
	ctx := context.Background()

	// Pre-populate cache with keys we'll delete
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Delete(ctx, fmt.Sprintf("query-%d", i))
	}
}

// Benchmark ListAll
func BenchmarkMemCache_ListAll_10(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(10)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ListAll(ctx)
	}
}

func BenchmarkRedisCache_ListAll_10(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(10)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ListAll(ctx)
	}
}

func BenchmarkMemCache_ListAll_100(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(10)
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ListAll(ctx)
	}
}

func BenchmarkRedisCache_ListAll_100(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(10)
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ListAll(ctx)
	}
}

// Benchmark concurrent operations
func BenchmarkMemCache_ConcurrentReadWrite(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(50)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cache.Set(ctx, fmt.Sprintf("query-%d", i%100), vec, false, "")
			} else {
				cache.Get(ctx, fmt.Sprintf("query-%d", i%100))
			}
			i++
		}
	})
}

func BenchmarkRedisCache_ConcurrentReadWrite(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(50)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cache.Set(ctx, fmt.Sprintf("query-%d", i%100), vec, false, "")
			} else {
				cache.Get(ctx, fmt.Sprintf("query-%d", i%100))
			}
			i++
		}
	})
}

// Benchmark Size operations
func BenchmarkMemCache_Size(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(10)
	ctx := context.Background()

	for i := 0; i < 50; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Size(ctx)
	}
}

func BenchmarkRedisCache_Size(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(10)
	ctx := context.Background()

	for i := 0; i < 50; i++ {
		cache.Set(ctx, fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Size(ctx)
	}
}
