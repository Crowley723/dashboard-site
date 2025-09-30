package data

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"homelab-dashboard/internal/config"

	"github.com/prometheus/common/model"
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
	return NewMemCache(&config.Config{}, logger)
}

func setupRedisCache(b *testing.B) *RedisCache {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	cfg := &config.Config{
		Redis: &config.RedisConfig{
			Address:    "127.0.1.50:6379",
			Password:   "",
			CacheIndex: 1,
		},
	}
	cache := NewRedisCache(cfg, logger)

	// Clean up any existing test data
	conn := cache.pool.Get()
	conn.Do("FLUSHDB")
	conn.Close()

	b.Cleanup(func() {
		cache.ClosePool()
	})

	return cache
}

// Benchmark Set operations
func BenchmarkMemCache_Set_SmallVector(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("query-%d", i%100), vec, false, "")
	}
}

func BenchmarkRedisCache_Set_SmallVector(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("query-%d", i%100), vec, false, "")
	}
}

func BenchmarkMemCache_Set_LargeVector(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("query-%d", i%100), vec, false, "")
	}
}

func BenchmarkRedisCache_Set_LargeVector(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("query-%d", i%100), vec, false, "")
	}
}

func BenchmarkMemCache_Set_Matrix(b *testing.B) {
	cache := setupMemCache()
	matrix := createTestMatrix(50, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("query-%d", i%100), matrix, false, "")
	}
}

func BenchmarkRedisCache_Set_Matrix(b *testing.B) {
	cache := setupRedisCache(b)
	matrix := createTestMatrix(50, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("query-%d", i%100), matrix, false, "")
	}
}

// Benchmark Get operations
func BenchmarkMemCache_Get(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(100)

	// Pre-populate
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("query-%d", i%100))
	}
}

func BenchmarkRedisCache_Get(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(100)

	// Pre-populate
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("query-%d", i%100))
	}
}

// Benchmark Get misses
func BenchmarkMemCache_GetMiss(b *testing.B) {
	cache := setupMemCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("nonexistent-%d", i))
	}
}

func BenchmarkRedisCache_GetMiss(b *testing.B) {
	cache := setupRedisCache(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("nonexistent-%d", i))
	}
}

// Benchmark Delete operations
func BenchmarkMemCache_Delete(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(10)

	// Pre-populate cache with keys we'll delete
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Delete(fmt.Sprintf("query-%d", i))
	}
}

func BenchmarkRedisCache_Delete(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(10)

	// Pre-populate cache with keys we'll delete
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Delete(fmt.Sprintf("query-%d", i))
	}
}

// Benchmark ListAll
func BenchmarkMemCache_ListAll_10(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(10)

	for i := 0; i < 10; i++ {
		cache.Set(fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ListAll()
	}
}

func BenchmarkRedisCache_ListAll_10(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(10)

	for i := 0; i < 10; i++ {
		cache.Set(fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ListAll()
	}
}

func BenchmarkMemCache_ListAll_100(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(10)

	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ListAll()
	}
}

func BenchmarkRedisCache_ListAll_100(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(10)

	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.ListAll()
	}
}

// Benchmark concurrent operations
func BenchmarkMemCache_ConcurrentReadWrite(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(50)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cache.Set(fmt.Sprintf("query-%d", i%100), vec, false, "")
			} else {
				cache.Get(fmt.Sprintf("query-%d", i%100))
			}
			i++
		}
	})
}

func BenchmarkRedisCache_ConcurrentReadWrite(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(50)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				cache.Set(fmt.Sprintf("query-%d", i%100), vec, false, "")
			} else {
				cache.Get(fmt.Sprintf("query-%d", i%100))
			}
			i++
		}
	})
}

// Benchmark Size operations
func BenchmarkMemCache_Size(b *testing.B) {
	cache := setupMemCache()
	vec := createTestVector(10)

	for i := 0; i < 50; i++ {
		cache.Set(fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Size()
	}
}

func BenchmarkRedisCache_Size(b *testing.B) {
	cache := setupRedisCache(b)
	vec := createTestVector(10)

	for i := 0; i < 50; i++ {
		cache.Set(fmt.Sprintf("query-%d", i), vec, false, "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Size()
	}
}
