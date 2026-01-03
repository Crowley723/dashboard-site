package data

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisCacheClient is a mock implementation of RedisCacheClient
type MockRedisCacheClient struct {
	mock.Mock
}

func (m *MockRedisCacheClient) GetDel(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisCacheClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisCacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisCacheClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisCacheClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	args := m.Called(ctx, pattern)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (m *MockRedisCacheClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisCacheClient) PoolStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

func (m *MockRedisCacheClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Helper function to create a StringCmd with a result
func createStringCmd(result string, err error) *redis.StringCmd {
	cmd := redis.NewStringCmd(context.Background())
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(result)
	}
	return cmd
}

// Helper function to create a StatusCmd
func createStatusCmd(err error) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(context.Background())
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal("OK")
	}
	return cmd
}

// Helper function to create an IntCmd
func createIntCmd(result int64, err error) *redis.IntCmd {
	cmd := redis.NewIntCmd(context.Background())
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(result)
	}
	return cmd
}

// Helper function to create a StringSliceCmd
func createStringSliceCmd(result []string, err error) *redis.StringSliceCmd {
	cmd := redis.NewStringSliceCmd(context.Background())
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(result)
	}
	return cmd
}

func TestRedisCache_Key(t *testing.T) {
	mockClient := new(MockRedisCacheClient)
	cache := &RedisCache{
		client: mockClient,
		logger: slog.Default(),
	}

	tests := []struct {
		name      string
		queryName string
		expected  string
	}{
		{
			name:      "simple query name",
			queryName: "users",
			expected:  "cache:query:users",
		},
		{
			name:      "query name with underscores",
			queryName: "active_users_count",
			expected:  "cache:query:active_users_count",
		},
		{
			name:      "empty query name",
			queryName: "",
			expected:  "cache:query:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cache.key(tt.queryName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedisCache_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		cachedData := CachedData{
			ValueJSON: string(`{"test": "data"}`),
			Timestamp: time.Now(),
		}
		jsonData, _ := json.Marshal(cachedData)

		mockClient.On("Get", ctx, "cache:query:test").
			Return(createStringCmd(string(jsonData), nil))

		result, found := cache.Get(ctx, "test")
		assert.True(t, found)
		assert.Equal(t, cachedData.ValueJSON, result.ValueJSON)
		mockClient.AssertExpectations(t)
	})

	t.Run("cache miss - key not found", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		mockClient.On("Get", ctx, "cache:query:missing").
			Return(createStringCmd("", redis.Nil))

		result, found := cache.Get(ctx, "missing")
		assert.False(t, found)
		assert.Equal(t, CachedData{}, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("redis error", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		mockClient.On("Get", ctx, "cache:query:error").
			Return(createStringCmd("", errors.New("connection error")))

		result, found := cache.Get(ctx, "error")
		assert.False(t, found)
		assert.Equal(t, CachedData{}, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("invalid json data", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		mockClient.On("Get", ctx, "cache:query:invalid").
			Return(createStringCmd("invalid json", nil))

		result, found := cache.Get(ctx, "invalid")
		assert.False(t, found)
		assert.Equal(t, CachedData{}, result)
		mockClient.AssertExpectations(t)
	})
}

func TestRedisCache_Set(t *testing.T) {
	ctx := context.Background()

	t.Run("successful set", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		cachedData := CachedData{
			ValueJSON: `{"test": "data"}`,
			Timestamp: time.Now(),
		}

		mockClient.On("Set", ctx, "cache:query:test", mock.Anything, time.Duration(0)).
			Return(createStatusCmd(nil))

		cache.Set(ctx, "test", cachedData)
		mockClient.AssertExpectations(t)
	})

	t.Run("set with redis error", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		cachedData := CachedData{
			ValueJSON: `{"test": "data"}`,
			Timestamp: time.Now(),
		}

		mockClient.On("Set", ctx, "cache:query:test", mock.Anything, time.Duration(0)).
			Return(createStatusCmd(errors.New("connection error")))

		// Should not panic, just log error
		cache.Set(ctx, "test", cachedData)
		mockClient.AssertExpectations(t)
	})
}

func TestRedisCache_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		mockClient.On("Del", ctx, []string{"cache:query:test"}).
			Return(createIntCmd(1, nil))

		cache.Delete(ctx, "test")
		mockClient.AssertExpectations(t)
	})

	t.Run("delete with redis error", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		mockClient.On("Del", ctx, []string{"cache:query:test"}).
			Return(createIntCmd(0, errors.New("connection error")))

		// Should not panic, just log error
		cache.Delete(ctx, "test")
		mockClient.AssertExpectations(t)
	})
}

func TestRedisCache_ListAll(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list all", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		keys := []string{
			"cache:query:users",
			"cache:query:posts",
			"cache:query:comments",
		}

		mockClient.On("Keys", ctx, "cache:query:*").
			Return(createStringSliceCmd(keys, nil))

		result := cache.ListAll(ctx)
		assert.Len(t, result, 3)
		assert.Contains(t, result, "users")
		assert.Contains(t, result, "posts")
		assert.Contains(t, result, "comments")
		mockClient.AssertExpectations(t)
	})

	t.Run("list all with no keys", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		mockClient.On("Keys", ctx, "cache:query:*").
			Return(createStringSliceCmd([]string{}, nil))

		result := cache.ListAll(ctx)
		assert.Len(t, result, 0)
		mockClient.AssertExpectations(t)
	})

	t.Run("list all with redis error", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		mockClient.On("Keys", ctx, "cache:query:*").
			Return(createStringSliceCmd(nil, errors.New("connection error")))

		result := cache.ListAll(ctx)
		assert.Len(t, result, 0)
		mockClient.AssertExpectations(t)
	})
}

func TestRedisCache_Size(t *testing.T) {
	ctx := context.Background()

	t.Run("successful size calculation", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		keys := []string{
			"cache:query:users",
			"cache:query:posts",
			"cache:query:comments",
		}

		mockClient.On("Keys", ctx, "cache:query:*").
			Return(createStringSliceCmd(keys, nil))

		result := cache.Size(ctx)
		assert.Equal(t, 3, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("size with no keys", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		mockClient.On("Keys", ctx, "cache:query:*").
			Return(createStringSliceCmd([]string{}, nil))

		result := cache.Size(ctx)
		assert.Equal(t, 0, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("size with redis error", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		mockClient.On("Keys", ctx, "cache:query:*").
			Return(createStringSliceCmd(nil, errors.New("connection error")))

		result := cache.Size(ctx)
		assert.Equal(t, 0, result)
		mockClient.AssertExpectations(t)
	})
}

func TestRedisCache_ClosePool(t *testing.T) {
	t.Run("successful close", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		mockClient.On("Close").Return(nil)

		err := cache.ClosePool()
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("close with error", func(t *testing.T) {
		mockClient := new(MockRedisCacheClient)
		cache := &RedisCache{
			client: mockClient,
			logger: slog.Default(),
		}

		expectedErr := errors.New("close error")
		mockClient.On("Close").Return(expectedErr)

		err := cache.ClosePool()
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockClient.AssertExpectations(t)
	})
}
