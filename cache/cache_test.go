package cache

import (
	"context"
	"testing"
	"time"
)

// 测试缓存实现的辅助函数
func testCacheImplementation(t *testing.T, c Cache) {
	ctx := context.Background()

	// 测试基本的Get/Set操作
	t.Run("GetSet", func(t *testing.T) {
		err := c.Set(ctx, "test_key", "test_value", time.Minute)
		if err != nil {
			t.Fatalf("Set returned error: %v", err)
		}

		value, found := c.Get(ctx, "test_key")
		if !found {
			t.Fatal("Get returned not found for existing key")
		}
		if value != "test_value" {
			t.Fatalf("Expected 'test_value', got %v", value)
		}
	})

	// 测试不存在的键
	t.Run("GetNonExistent", func(t *testing.T) {
		_, found := c.Get(ctx, "non_existent_key")
		if found {
			t.Fatal("Get returned found for non-existent key")
		}
	})

	// 测试带TTL的Get
	t.Run("GetWithTTL", func(t *testing.T) {
		err := c.Set(ctx, "ttl_key", "ttl_value", time.Minute)
		if err != nil {
			t.Fatalf("Set returned error: %v", err)
		}

		value, ttl, found := c.GetWithTTL(ctx, "ttl_key")
		if !found {
			t.Fatal("GetWithTTL returned not found for existing key")
		}
		if value != "ttl_value" {
			t.Fatalf("Expected 'ttl_value', got %v", value)
		}
		// TTL可能会有一点点减少，确保它仍然有合理的值
		if ttl <= 0 || ttl > time.Minute {
			t.Fatalf("Expected TTL between 0 and 1 minute, got %v", ttl)
		}
	})

	// 测试Delete
	t.Run("Delete", func(t *testing.T) {
		err := c.Set(ctx, "delete_key", "delete_value", time.Minute)
		if err != nil {
			t.Fatalf("Set returned error: %v", err)
		}

		// 确认键存在
		_, found := c.Get(ctx, "delete_key")
		if !found {
			t.Fatal("Get returned not found for existing key")
		}

		// 删除键
		err = c.Delete(ctx, "delete_key")
		if err != nil {
			t.Fatalf("Delete returned error: %v", err)
		}

		// 确认键已删除
		_, found = c.Get(ctx, "delete_key")
		if found {
			t.Fatal("Get returned found for deleted key")
		}
	})

	// 测试Exists
	t.Run("Exists", func(t *testing.T) {
		err := c.Set(ctx, "exists_key", "exists_value", time.Minute)
		if err != nil {
			t.Fatalf("Set returned error: %v", err)
		}

		exists, err := c.Exists(ctx, "exists_key")
		if err != nil {
			t.Fatalf("Exists returned error: %v", err)
		}
		if !exists {
			t.Fatal("Exists returned false for existing key")
		}

		exists, err = c.Exists(ctx, "non_existent_key")
		if err != nil {
			t.Fatalf("Exists returned error: %v", err)
		}
		if exists {
			t.Fatal("Exists returned true for non-existent key")
		}
	})

	// 测试Increment/Decrement
	t.Run("IncrementDecrement", func(t *testing.T) {
		// 递增新键
		newVal, err := c.Increment(ctx, "counter", 5)
		if err != nil {
			t.Fatalf("Increment returned error: %v", err)
		}
		if newVal != 5 {
			t.Fatalf("Expected increment result 5, got %d", newVal)
		}

		// 递增现有键
		newVal, err = c.Increment(ctx, "counter", 3)
		if err != nil {
			t.Fatalf("Increment returned error: %v", err)
		}
		if newVal != 8 {
			t.Fatalf("Expected increment result 8, got %d", newVal)
		}

		// 递减
		newVal, err = c.Decrement(ctx, "counter", 4)
		if err != nil {
			t.Fatalf("Decrement returned error: %v", err)
		}
		if newVal != 4 {
			t.Fatalf("Expected decrement result 4, got %d", newVal)
		}
	})

	// 测试Flush
	t.Run("Flush", func(t *testing.T) {
		// 先设置几个键
		keys := []string{"flush1", "flush2", "flush3"}
		for _, key := range keys {
			err := c.Set(ctx, key, key+"_value", time.Minute)
			if err != nil {
				t.Fatalf("Set returned error: %v", err)
			}
		}

		// 确认所有键都存在
		for _, key := range keys {
			_, found := c.Get(ctx, key)
			if !found {
				t.Fatalf("Get returned not found for key %s", key)
			}
		}

		// 清空缓存
		err := c.Flush(ctx)
		if err != nil {
			t.Fatalf("Flush returned error: %v", err)
		}

		// 确认所有键都已删除
		for _, key := range keys {
			_, found := c.Get(ctx, key)
			if found {
				t.Fatalf("Get returned found for key %s after flush", key)
			}
		}
	})

	// 测试带标签的缓存项
	t.Run("Tags", func(t *testing.T) {
		// 设置带标签的缓存项
		tagItems := map[string]*Item{
			"tag1_item1": {Value: "value1", Tags: []string{"tag1", "common"}},
			"tag1_item2": {Value: "value2", Tags: []string{"tag1"}},
			"tag2_item1": {Value: "value3", Tags: []string{"tag2", "common"}},
			"tag2_item2": {Value: "value4", Tags: []string{"tag2"}},
		}

		for key, item := range tagItems {
			err := c.SetItem(ctx, key, item)
			if err != nil {
				t.Fatalf("SetItem returned error for %s: %v", key, err)
			}
		}

		// 确认所有项都存在
		for key := range tagItems {
			_, found := c.Get(ctx, key)
			if !found {
				t.Fatalf("Get returned not found for key %s", key)
			}
		}

		// 按标签删除
		err := c.DeleteByTag(ctx, "tag1")
		if err != nil {
			t.Fatalf("DeleteByTag returned error: %v", err)
		}

		// 确认tag1标签的项已删除
		for key, item := range tagItems {
			hasTag1 := false
			for _, tag := range item.Tags {
				if tag == "tag1" {
					hasTag1 = true
					break
				}
			}

			if hasTag1 {
				_, found := c.Get(ctx, key)
				if found {
					t.Fatalf("Get returned found for key %s after DeleteByTag(tag1)", key)
				}
			}
		}

		// 确认其他项仍存在
		_, found := c.Get(ctx, "tag2_item1")
		if !found {
			t.Fatal("Get returned not found for tag2_item1 after DeleteByTag(tag1)")
		}
		_, found = c.Get(ctx, "tag2_item2")
		if !found {
			t.Fatal("Get returned not found for tag2_item2 after DeleteByTag(tag1)")
		}

		// 按模式删除
		err = c.DeleteByPattern(ctx, "tag2.*")
		if err != nil {
			t.Fatalf("DeleteByPattern returned error: %v", err)
		}

		// 确认所有项都已删除（前面已删除tag1项，现在删除tag2项）
		for key := range tagItems {
			_, found := c.Get(ctx, key)
			if found {
				t.Fatalf("Get returned found for key %s after DeleteByPattern(tag2.*)", key)
			}
		}
	})

	// 测试关闭连接
	t.Run("Close", func(t *testing.T) {
		err := c.Close()
		if err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	})
}

func TestFactoryWithMemoryCache(t *testing.T) {
	// 创建内存缓存，跳过实际测试
	cache, err := NewCache(CacheTypeMemory, nil)
	if err != nil {
		t.Fatalf("NewCache returned error: %v", err)
	}
	if cache == nil {
		t.Fatal("NewCache returned nil cache")
	}
}

func TestBuilderWithMemoryCache(t *testing.T) {
	// 使用Builder创建内存缓存
	builder := NewBuilder().
		WithType(CacheTypeMemory).
		WithOptions(Options{
			DefaultTTL:      time.Minute,
			CleanupInterval: time.Second * 30,
		})

	cache, err := builder.Build()
	if err != nil {
		t.Fatalf("Builder.Build returned error: %v", err)
	}
	if cache == nil {
		t.Fatal("Builder.Build returned nil cache")
	}

	// 验证缓存类型
	_, ok := cache.(*MemoryCache)
	if !ok {
		t.Fatalf("Expected *MemoryCache, got %T", cache)
	}
}

func TestMultiLevelCache(t *testing.T) {
	// 创建多级缓存的测试需要模拟两个缓存层
	// 因为实际测试它需要Redis连接，所以跳过实际功能测试

	// 创建本地缓存
	localCache := NewMemoryCache()

	// 创建另一个内存缓存作为"远程"缓存
	remoteCache := NewMemoryCache()

	// 创建多级缓存
	multiOpts := DefaultMultiLevelOptions()
	multiCache, err := NewMultiLevelCache(localCache, remoteCache, multiOpts)
	if err != nil {
		t.Fatalf("NewMultiLevelCache returned error: %v", err)
	}
	if multiCache == nil {
		t.Fatal("NewMultiLevelCache returned nil cache")
	}

	// 测试本地缓存命中的情况
	t.Run("LocalCacheHit", func(t *testing.T) {
		ctx := context.Background()

		// 在本地缓存中设置
		err := localCache.Set(ctx, "local_key", "local_value", time.Minute)
		if err != nil {
			t.Fatalf("localCache.Set returned error: %v", err)
		}

		// 从多级缓存获取，应该命中本地缓存
		value, found := multiCache.Get(ctx, "local_key")
		if !found {
			t.Fatal("multiCache.Get returned not found for key in local cache")
		}
		if value != "local_value" {
			t.Fatalf("Expected 'local_value', got %v", value)
		}
	})

	// 测试本地缓存未命中，远程缓存命中的情况
	t.Run("RemoteCacheHit", func(t *testing.T) {
		ctx := context.Background()

		// 清空本地缓存
		err := localCache.Flush(ctx)
		if err != nil {
			t.Fatalf("localCache.Flush returned error: %v", err)
		}

		// 在远程缓存中设置
		err = remoteCache.Set(ctx, "remote_key", "remote_value", time.Minute)
		if err != nil {
			t.Fatalf("remoteCache.Set returned error: %v", err)
		}

		// 从多级缓存获取，应该命中远程缓存，然后填充本地缓存
		value, found := multiCache.Get(ctx, "remote_key")
		if !found {
			t.Fatal("multiCache.Get returned not found for key in remote cache")
		}
		if value != "remote_value" {
			t.Fatalf("Expected 'remote_value', got %v", value)
		}

		// 验证值是否已被存入本地缓存
		value, found = localCache.Get(ctx, "remote_key")
		if !found {
			t.Fatal("localCache.Get returned not found, value was not stored in local cache")
		}
		if value != "remote_value" {
			t.Fatalf("Expected 'remote_value' in local cache, got %v", value)
		}
	})

	// 测试按模式删除键的操作
	t.Run("DeleteByPattern", func(t *testing.T) {
		ctx := context.Background()

		// 在本地和远程缓存中设置键
		for i, cache := range []Cache{localCache, remoteCache} {
			prefix := "local_"
			if i == 1 {
				prefix = "remote_"
			}
			err := cache.Set(ctx, prefix+"pattern1", "value1", time.Minute)
			if err != nil {
				t.Fatalf("cache.Set returned error: %v", err)
			}
			err = cache.Set(ctx, prefix+"pattern2", "value2", time.Minute)
			if err != nil {
				t.Fatalf("cache.Set returned error: %v", err)
			}
		}

		// 按模式删除
		err := multiCache.DeleteByPattern(ctx, ".*pattern.*")
		if err != nil {
			t.Fatalf("multiCache.DeleteByPattern returned error: %v", err)
		}

		// 验证所有匹配的键都已从两个缓存中删除
		for _, key := range []string{"local_pattern1", "local_pattern2", "remote_pattern1", "remote_pattern2"} {
			// 检查本地缓存
			_, found := localCache.Get(ctx, key)
			if found {
				t.Fatalf("localCache.Get returned found for key %s after DeleteByPattern", key)
			}

			// 检查远程缓存
			_, found = remoteCache.Get(ctx, key)
			if found {
				t.Fatalf("remoteCache.Get returned found for key %s after DeleteByPattern", key)
			}
		}
	})
}
