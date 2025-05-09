package cache

import (
	"context"
	"testing"
	"time"
)

func TestNewMultiLevelCache(t *testing.T) {
	// 创建本地和远程缓存
	localCache := NewMemoryCache()
	remoteCache := NewMemoryCache()

	// 默认选项
	multiCache, err := NewMultiLevelCache(localCache, remoteCache)
	if err != nil {
		t.Fatalf("NewMultiLevelCache with default options returned error: %v", err)
	}
	if multiCache == nil {
		t.Fatal("NewMultiLevelCache with default options returned nil cache")
	}

	// 自定义选项
	multiCache, err = NewMultiLevelCache(localCache, remoteCache, MultiLevelOptions{
		WriteMode: WriteModeWriteBack,
		LocalTTL:  time.Second * 30,
	})
	if err != nil {
		t.Fatalf("NewMultiLevelCache with custom options returned error: %v", err)
	}
	if multiCache == nil {
		t.Fatal("NewMultiLevelCache with custom options returned nil cache")
	}

	// 空本地缓存
	_, err = NewMultiLevelCache(nil, remoteCache)
	if err == nil {
		t.Fatal("NewMultiLevelCache with nil local cache should return error")
	}

	// 空远程缓存
	_, err = NewMultiLevelCache(localCache, nil)
	if err == nil {
		t.Fatal("NewMultiLevelCache with nil remote cache should return error")
	}
}

func TestMultiLevelCacheImplementation(t *testing.T) {
	// 创建本地和远程缓存
	localCache := NewMemoryCache()
	remoteCache := NewMemoryCache()

	// 创建多级缓存
	multiCache, err := NewMultiLevelCache(localCache, remoteCache)
	if err != nil {
		t.Fatalf("NewMultiLevelCache returned error: %v", err)
	}

	// 使用通用测试函数
	testCacheImplementation(t, multiCache)
}

func TestMultiLevelCache_CacheHierarchy(t *testing.T) {
	// 创建本地和远程缓存
	localCache := NewMemoryCache()
	remoteCache := NewMemoryCache()

	// 创建直写模式的多级缓存
	writeThroughCache, err := NewMultiLevelCache(localCache, remoteCache, MultiLevelOptions{
		WriteMode: WriteModeWriteThrough,
		LocalTTL:  time.Minute,
	})
	if err != nil {
		t.Fatalf("NewMultiLevelCache (write-through) returned error: %v", err)
	}

	ctx := context.Background()

	// 测试直写模式
	t.Run("WriteThrough", func(t *testing.T) {
		// 清空缓存
		localCache.Flush(ctx)
		remoteCache.Flush(ctx)

		// 设置值
		err := writeThroughCache.Set(ctx, "wt_key", "wt_value", time.Minute*5)
		if err != nil {
			t.Fatalf("WriteThrough Set returned error: %v", err)
		}

		// 验证本地缓存已设置（使用本地TTL）
		value, ttl, found := localCache.GetWithTTL(ctx, "wt_key")
		if !found {
			t.Fatal("WriteThrough: Local cache miss")
		}
		if value != "wt_value" {
			t.Fatalf("WriteThrough: Expected local value 'wt_value', got %v", value)
		}
		if ttl > time.Minute {
			t.Fatalf("WriteThrough: Expected local TTL <= 1 minute, got %v", ttl)
		}

		// 验证远程缓存已设置（使用原始TTL）
		value, ttl, found = remoteCache.GetWithTTL(ctx, "wt_key")
		if !found {
			t.Fatal("WriteThrough: Remote cache miss")
		}
		if value != "wt_value" {
			t.Fatalf("WriteThrough: Expected remote value 'wt_value', got %v", value)
		}
		if ttl <= time.Minute || ttl > time.Minute*5 {
			t.Fatalf("WriteThrough: Expected remote TTL between 1-5 minutes, got %v", ttl)
		}
	})

	// 创建回写模式的多级缓存
	writeBackCache, err := NewMultiLevelCache(localCache, remoteCache, MultiLevelOptions{
		WriteMode: WriteModeWriteBack,
		LocalTTL:  time.Minute,
	})
	if err != nil {
		t.Fatalf("NewMultiLevelCache (write-back) returned error: %v", err)
	}

	// 测试回写模式
	t.Run("WriteBack", func(t *testing.T) {
		// 清空缓存
		localCache.Flush(ctx)
		remoteCache.Flush(ctx)

		// 设置值
		err := writeBackCache.Set(ctx, "wb_key", "wb_value", time.Minute*5)
		if err != nil {
			t.Fatalf("WriteBack Set returned error: %v", err)
		}

		// 验证本地缓存未设置
		_, found := localCache.Get(ctx, "wb_key")
		if found {
			t.Fatal("WriteBack: Unexpected local cache hit")
		}

		// 验证远程缓存已设置
		value, found := remoteCache.Get(ctx, "wb_key")
		if !found {
			t.Fatal("WriteBack: Remote cache miss")
		}
		if value != "wb_value" {
			t.Fatalf("WriteBack: Expected remote value 'wb_value', got %v", value)
		}

		// 从多级缓存获取，应该从远程加载并填充本地
		value, found = writeBackCache.Get(ctx, "wb_key")
		if !found {
			t.Fatal("WriteBack: Multi-level cache miss")
		}
		if value != "wb_value" {
			t.Fatalf("WriteBack: Expected multi-level value 'wb_value', got %v", value)
		}

		// 验证本地缓存现在已填充
		value, found = localCache.Get(ctx, "wb_key")
		if !found {
			t.Fatal("WriteBack: Local cache miss after Get")
		}
		if value != "wb_value" {
			t.Fatalf("WriteBack: Expected local value 'wb_value', got %v", value)
		}
	})
}

func TestMultiLevelCache_DeleteOperations(t *testing.T) {
	// 创建本地和远程缓存
	localCache := NewMemoryCache()
	remoteCache := NewMemoryCache()

	// 创建多级缓存
	multiCache, err := NewMultiLevelCache(localCache, remoteCache)
	if err != nil {
		t.Fatalf("NewMultiLevelCache returned error: %v", err)
	}

	ctx := context.Background()

	// 设置一些测试数据
	testKeys := []string{"delete1", "delete2", "pattern_key1", "pattern_key2"}
	for _, key := range testKeys {
		// 直接向两个缓存写入，绕过多级缓存逻辑
		localCache.Set(ctx, key, key+"_local", time.Minute)
		remoteCache.Set(ctx, key, key+"_remote", time.Minute)
	}

	// 测试Delete
	t.Run("Delete", func(t *testing.T) {
		err := multiCache.Delete(ctx, "delete1")
		if err != nil {
			t.Fatalf("Delete returned error: %v", err)
		}

		// 验证本地缓存中已删除
		_, found := localCache.Get(ctx, "delete1")
		if found {
			t.Fatal("Expected key delete1 to be deleted from local cache")
		}

		// 验证远程缓存中已删除
		_, found = remoteCache.Get(ctx, "delete1")
		if found {
			t.Fatal("Expected key delete1 to be deleted from remote cache")
		}
	})

	// 测试DeleteByPattern
	t.Run("DeleteByPattern", func(t *testing.T) {
		err := multiCache.DeleteByPattern(ctx, "pattern_.*")
		if err != nil {
			t.Fatalf("DeleteByPattern returned error: %v", err)
		}

		// 验证本地缓存中的模式键已删除
		for _, key := range []string{"pattern_key1", "pattern_key2"} {
			_, found := localCache.Get(ctx, key)
			if found {
				t.Fatalf("Expected key %s to be deleted from local cache", key)
			}
		}

		// 验证远程缓存中的模式键已删除
		for _, key := range []string{"pattern_key1", "pattern_key2"} {
			_, found := remoteCache.Get(ctx, key)
			if found {
				t.Fatalf("Expected key %s to be deleted from remote cache", key)
			}
		}

		// 验证非模式键仍存在
		_, found := localCache.Get(ctx, "delete2")
		if !found {
			t.Fatal("Expected key delete2 to still exist in local cache")
		}
		_, found = remoteCache.Get(ctx, "delete2")
		if !found {
			t.Fatal("Expected key delete2 to still exist in remote cache")
		}
	})

	// 测试Flush
	t.Run("Flush", func(t *testing.T) {
		err := multiCache.Flush(ctx)
		if err != nil {
			t.Fatalf("Flush returned error: %v", err)
		}

		// 验证所有键都已从本地缓存中删除
		for _, key := range testKeys {
			_, found := localCache.Get(ctx, key)
			if found {
				t.Fatalf("Expected key %s to be deleted from local cache after Flush", key)
			}
		}

		// 验证所有键都已从远程缓存中删除
		for _, key := range testKeys {
			_, found := remoteCache.Get(ctx, key)
			if found {
				t.Fatalf("Expected key %s to be deleted from remote cache after Flush", key)
			}
		}
	})
}

func TestMultiLevelCache_Increment(t *testing.T) {
	// 创建本地和远程缓存
	localCache := NewMemoryCache()
	remoteCache := NewMemoryCache()

	// 创建不同写入模式的多级缓存
	writeThroughCache, _ := NewMultiLevelCache(localCache, remoteCache, MultiLevelOptions{
		WriteMode: WriteModeWriteThrough,
		LocalTTL:  time.Minute,
	})
	writeBackCache, _ := NewMultiLevelCache(localCache, remoteCache, MultiLevelOptions{
		WriteMode: WriteModeWriteBack,
		LocalTTL:  time.Minute,
	})

	ctx := context.Background()

	// 测试直写模式下的Increment
	t.Run("IncrementWriteThrough", func(t *testing.T) {
		// 清空缓存
		localCache.Flush(ctx)
		remoteCache.Flush(ctx)

		// 递增值
		newVal, err := writeThroughCache.Increment(ctx, "wt_counter", 5)
		if err != nil {
			t.Fatalf("Increment returned error: %v", err)
		}
		if newVal != 5 {
			t.Fatalf("Expected Increment result 5, got %d", newVal)
		}

		// 验证本地缓存已更新
		value, found := localCache.Get(ctx, "wt_counter")
		if !found {
			t.Fatal("Expected key wt_counter to exist in local cache")
		}
		if value != int64(5) {
			t.Fatalf("Expected local value 5, got %v", value)
		}

		// 验证远程缓存已更新
		value, found = remoteCache.Get(ctx, "wt_counter")
		if !found {
			t.Fatal("Expected key wt_counter to exist in remote cache")
		}
		if value != int64(5) {
			t.Fatalf("Expected remote value 5, got %v", value)
		}
	})

	// 测试回写模式下的Increment
	t.Run("IncrementWriteBack", func(t *testing.T) {
		// 清空缓存
		localCache.Flush(ctx)
		remoteCache.Flush(ctx)

		// 递增值
		newVal, err := writeBackCache.Increment(ctx, "wb_counter", 7)
		if err != nil {
			t.Fatalf("Increment returned error: %v", err)
		}
		if newVal != 7 {
			t.Fatalf("Expected Increment result 7, got %d", newVal)
		}

		// 验证本地缓存中的键已删除（回写模式下）
		_, found := localCache.Get(ctx, "wb_counter")
		if found {
			t.Fatal("Expected key wb_counter to not exist in local cache in write-back mode")
		}

		// 验证远程缓存已更新
		value, found := remoteCache.Get(ctx, "wb_counter")
		if !found {
			t.Fatal("Expected key wb_counter to exist in remote cache")
		}
		if value != int64(7) {
			t.Fatalf("Expected remote value 7, got %v", value)
		}

		// 在本地缓存中设置值
		localCache.Set(ctx, "wb_counter", int64(3), time.Minute)

		// 再次递增值
		newVal, err = writeBackCache.Increment(ctx, "wb_counter", 2)
		if err != nil {
			t.Fatalf("Increment returned error: %v", err)
		}
		if newVal != 9 {
			t.Fatalf("Expected Increment result 9, got %d", newVal)
		}

		// 验证本地缓存中的键已删除
		_, found = localCache.Get(ctx, "wb_counter")
		if found {
			t.Fatal("Expected key wb_counter to not exist in local cache after second increment")
		}

		// 验证远程缓存已更新
		value, found = remoteCache.Get(ctx, "wb_counter")
		if !found {
			t.Fatal("Expected key wb_counter to exist in remote cache")
		}
		if value != int64(9) {
			t.Fatalf("Expected remote value 9, got %v", value)
		}
	})
}

func TestMultiLevelCache_FlushLevel(t *testing.T) {
	// 创建本地和远程缓存
	localCache := NewMemoryCache()
	remoteCache := NewMemoryCache()

	// 创建多级缓存
	multiCache, err := NewMultiLevelCache(localCache, remoteCache)
	if err != nil {
		t.Fatalf("NewMultiLevelCache returned error: %v", err)
	}

	ctx := context.Background()

	// 设置一些测试数据
	testKeys := []string{"flush_level1", "flush_level2"}
	for _, key := range testKeys {
		// 直接向两个缓存写入，绕过多级缓存逻辑
		localCache.Set(ctx, key, key+"_local", time.Minute)
		remoteCache.Set(ctx, key, key+"_remote", time.Minute)
	}

	// 测试只清空本地缓存
	t.Run("FlushLocalOnly", func(t *testing.T) {
		err := multiCache.FlushLevel(ctx, CacheLevelLocal)
		if err != nil {
			t.Fatalf("FlushLevel(Local) returned error: %v", err)
		}

		// 验证本地缓存为空
		for _, key := range testKeys {
			_, found := localCache.Get(ctx, key)
			if found {
				t.Fatalf("Expected key %s to be deleted from local cache", key)
			}
		}

		// 验证远程缓存仍有数据
		for _, key := range testKeys {
			_, found := remoteCache.Get(ctx, key)
			if !found {
				t.Fatalf("Expected key %s to still exist in remote cache", key)
			}
		}
	})

	// 重新设置测试数据
	for _, key := range testKeys {
		localCache.Set(ctx, key, key+"_local", time.Minute)
	}

	// 测试只清空远程缓存
	t.Run("FlushRemoteOnly", func(t *testing.T) {
		err := multiCache.FlushLevel(ctx, CacheLevelRedis)
		if err != nil {
			t.Fatalf("FlushLevel(Redis) returned error: %v", err)
		}

		// 验证本地缓存仍有数据
		for _, key := range testKeys {
			_, found := localCache.Get(ctx, key)
			if !found {
				t.Fatalf("Expected key %s to still exist in local cache", key)
			}
		}

		// 验证远程缓存为空
		for _, key := range testKeys {
			_, found := remoteCache.Get(ctx, key)
			if found {
				t.Fatalf("Expected key %s to be deleted from remote cache", key)
			}
		}
	})

	// 重新设置测试数据
	for _, key := range testKeys {
		remoteCache.Set(ctx, key, key+"_remote", time.Minute)
	}

	// 测试清空所有缓存
	t.Run("FlushAll", func(t *testing.T) {
		err := multiCache.FlushLevel(ctx, CacheLevelAll)
		if err != nil {
			t.Fatalf("FlushLevel(All) returned error: %v", err)
		}

		// 验证本地缓存为空
		for _, key := range testKeys {
			_, found := localCache.Get(ctx, key)
			if found {
				t.Fatalf("Expected key %s to be deleted from local cache", key)
			}
		}

		// 验证远程缓存为空
		for _, key := range testKeys {
			_, found := remoteCache.Get(ctx, key)
			if found {
				t.Fatalf("Expected key %s to be deleted from remote cache", key)
			}
		}
	})
}
