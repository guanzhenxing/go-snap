package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_Basic(t *testing.T) {
	// 创建 Redis 缓存实例
	opts := RedisOptions{
		Mode:            RedisModeStandalone,
		Addr:            "localhost:6379",
		Password:        "123456",
		DB:              0,
		KeyPrefix:       "test",
		DefaultTTL:      time.Minute,
		DialTimeout:     time.Second * 10,
		ReadTimeout:     time.Second * 5,
		WriteTimeout:    time.Second * 5,
		PoolSize:        10,
		MinIdleConns:    2,
		MaxRetries:      3,
		MinRetryBackoff: time.Millisecond * 8,
		MaxRetryBackoff: time.Millisecond * 512,
	}

	cache, err := NewRedisCache(opts, nil)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// 测试 Set 和 Get
	t.Run("Set and Get", func(t *testing.T) {
		key := "test-key"
		value := "test-value"

		// 设置值
		err := cache.Set(ctx, key, value, time.Minute)
		require.NoError(t, err)

		// 获取值
		got, found := cache.Get(ctx, key)
		assert.True(t, found)
		assert.Equal(t, value, got)

		// 测试不存在的键
		_, found = cache.Get(ctx, "non-existent")
		assert.False(t, found)
	})

	// 测试 GetWithTTL
	t.Run("GetWithTTL", func(t *testing.T) {
		key := "ttl-key"
		value := "ttl-value"
		ttl := time.Second * 30

		// 设置带 TTL 的值
		err := cache.Set(ctx, key, value, ttl)
		require.NoError(t, err)

		// 获取值和 TTL
		got, gotTTL, found := cache.GetWithTTL(ctx, key)
		assert.True(t, found)
		assert.Equal(t, value, got)
		assert.True(t, gotTTL > 0)
		assert.True(t, gotTTL <= ttl)
	})

	// 测试 Delete
	t.Run("Delete", func(t *testing.T) {
		key := "delete-key"
		value := "delete-value"

		// 设置值
		err := cache.Set(ctx, key, value, time.Minute)
		require.NoError(t, err)

		// 删除值
		err = cache.Delete(ctx, key)
		require.NoError(t, err)

		// 确认值已被删除
		_, found := cache.Get(ctx, key)
		assert.False(t, found)
	})

	// 测试 Exists
	t.Run("Exists", func(t *testing.T) {
		key := "exists-key"
		value := "exists-value"

		// 设置值
		err := cache.Set(ctx, key, value, time.Minute)
		require.NoError(t, err)

		// 检查键是否存在
		exists, err := cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)

		// 检查不存在的键
		exists, err = cache.Exists(ctx, "non-existent")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	// 测试 Increment 和 Decrement
	t.Run("Increment and Decrement", func(t *testing.T) {
		key := "counter-key"

		// 增加计数
		val, err := cache.Increment(ctx, key, 5)
		require.NoError(t, err)
		assert.Equal(t, int64(5), val)

		// 再次增加
		val, err = cache.Increment(ctx, key, 3)
		require.NoError(t, err)
		assert.Equal(t, int64(8), val)

		// 减少计数
		val, err = cache.Decrement(ctx, key, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(6), val)
	})

	// 测试 Flush
	t.Run("Flush", func(t *testing.T) {
		// 设置一些测试数据
		keys := []string{"flush-key1", "flush-key2", "flush-key3"}
		for _, key := range keys {
			err := cache.Set(ctx, key, "value", time.Minute)
			require.NoError(t, err)
		}

		// 清空缓存
		err := cache.Flush(ctx)
		require.NoError(t, err)

		// 确认所有键都被删除
		for _, key := range keys {
			_, found := cache.Get(ctx, key)
			assert.False(t, found)
		}
	})
}

func TestRedisCache_Tags(t *testing.T) {
	opts := RedisOptions{
		Mode:            RedisModeStandalone,
		Addr:            "localhost:6379",
		Password:        "123456",
		DB:              0,
		KeyPrefix:       "test-tags",
		DefaultTTL:      time.Minute,
		DialTimeout:     time.Second * 10,
		ReadTimeout:     time.Second * 5,
		WriteTimeout:    time.Second * 5,
		PoolSize:        10,
		MinIdleConns:    2,
		MaxRetries:      3,
		MinRetryBackoff: time.Millisecond * 8,
		MaxRetryBackoff: time.Millisecond * 512,
	}

	cache, err := NewRedisCache(opts, nil)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// 测试带标签的缓存项
	t.Run("SetItem with Tags", func(t *testing.T) {
		// 创建带标签的缓存项
		item := &Item{
			Value:      "tagged-value",
			Expiration: time.Minute,
			Tags:       []string{"tag1", "tag2"},
		}

		// 设置带标签的缓存项
		err := cache.SetItem(ctx, "tagged-key", item)
		require.NoError(t, err)

		// 验证值可以正常获取
		value, found := cache.Get(ctx, "tagged-key")
		assert.True(t, found)
		assert.Equal(t, "tagged-value", value)
	})

	// 测试 DeleteByTag
	t.Run("DeleteByTag", func(t *testing.T) {
		// 设置多个带标签的缓存项
		items := map[string]*Item{
			"tag1-key1": {Value: "value1", Tags: []string{"tag1"}},
			"tag1-key2": {Value: "value2", Tags: []string{"tag1"}},
			"tag2-key1": {Value: "value3", Tags: []string{"tag2"}},
		}

		for key, item := range items {
			err := cache.SetItem(ctx, key, item)
			require.NoError(t, err)
		}

		// 删除 tag1 的所有缓存项
		err := cache.DeleteByTag(ctx, "tag1")
		require.NoError(t, err)

		// 验证 tag1 的缓存项已被删除
		_, found := cache.Get(ctx, "tag1-key1")
		assert.False(t, found)
		_, found = cache.Get(ctx, "tag1-key2")
		assert.False(t, found)

		// 验证 tag2 的缓存项仍然存在
		value, found := cache.Get(ctx, "tag2-key1")
		assert.True(t, found)
		assert.Equal(t, "value3", value)
	})
}

func TestRedisCache_Pattern(t *testing.T) {
	opts := RedisOptions{
		Mode:            RedisModeStandalone,
		Addr:            "localhost:6379",
		Password:        "123456",
		DB:              0,
		KeyPrefix:       "test-pattern",
		DefaultTTL:      time.Minute,
		DialTimeout:     time.Second * 10,
		ReadTimeout:     time.Second * 5,
		WriteTimeout:    time.Second * 5,
		PoolSize:        10,
		MinIdleConns:    2,
		MaxRetries:      3,
		MinRetryBackoff: time.Millisecond * 8,
		MaxRetryBackoff: time.Millisecond * 512,
	}

	cache, err := NewRedisCache(opts, nil)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// 测试 DeleteByPattern
	t.Run("DeleteByPattern", func(t *testing.T) {
		// 设置多个测试数据
		items := map[string]string{
			"user:1": "value1",
			"user:2": "value2",
			"post:1": "value3",
			"post:2": "value4",
		}

		for key, value := range items {
			err := cache.Set(ctx, key, value, time.Minute)
			require.NoError(t, err)
		}

		// 删除所有 user:* 模式的键
		err := cache.DeleteByPattern(ctx, "user:*")
		require.NoError(t, err)

		// 验证 user:* 模式的键已被删除
		_, found := cache.Get(ctx, "user:1")
		assert.False(t, found)
		_, found = cache.Get(ctx, "user:2")
		assert.False(t, found)

		// 验证 post:* 模式的键仍然存在
		value, found := cache.Get(ctx, "post:1")
		assert.True(t, found)
		assert.Equal(t, "value3", value)
		value, found = cache.Get(ctx, "post:2")
		assert.True(t, found)
		assert.Equal(t, "value4", value)
	})
}

func TestRedisCache_Expiration(t *testing.T) {
	opts := RedisOptions{
		Mode:            RedisModeStandalone,
		Addr:            "localhost:6379",
		Password:        "123456",
		DB:              0,
		KeyPrefix:       "test-expiration",
		DefaultTTL:      time.Minute,
		DialTimeout:     time.Second * 10,
		ReadTimeout:     time.Second * 5,
		WriteTimeout:    time.Second * 5,
		PoolSize:        10,
		MinIdleConns:    2,
		MaxRetries:      3,
		MinRetryBackoff: time.Millisecond * 8,
		MaxRetryBackoff: time.Millisecond * 512,
	}

	cache, err := NewRedisCache(opts, nil)
	require.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()

	// 测试过期时间
	t.Run("Expiration", func(t *testing.T) {
		key := "expiring-key"
		value := "expiring-value"
		ttl := time.Second * 2

		// 设置带短 TTL 的值
		err := cache.Set(ctx, key, value, ttl)
		require.NoError(t, err)

		// 立即获取值
		got, found := cache.Get(ctx, key)
		assert.True(t, found)
		assert.Equal(t, value, got)

		// 等待过期
		time.Sleep(ttl + time.Second)

		// 验证值已过期
		_, found = cache.Get(ctx, key)
		assert.False(t, found)
	})
}
