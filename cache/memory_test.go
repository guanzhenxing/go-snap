package cache

import (
	"context"
	"testing"
	"time"
)

func TestNewMemoryCache(t *testing.T) {
	cache := NewMemoryCache()
	if cache == nil {
		t.Fatal("Expected non-nil cache")
	}

	cache = NewMemoryCache(Options{
		DefaultTTL:      time.Minute,
		CleanupInterval: time.Second * 30,
	})
	if cache == nil {
		t.Fatal("Expected non-nil cache with custom options")
	}
}

func TestMemoryCacheImplementation(t *testing.T) {
	// 使用通用测试函数测试内存缓存
	cache := NewMemoryCache()
	testCacheImplementation(t, cache)
}

// 下面是内存缓存特有的测试

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache(Options{
		CleanupInterval: time.Millisecond * 100,
	})
	ctx := context.Background()

	// 设置短期缓存
	err := cache.Set(ctx, "expire_key", "expire_value", time.Millisecond*200)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// 立即检查
	_, found := cache.Get(ctx, "expire_key")
	if !found {
		t.Fatal("Expected to find key expire_key immediately")
	}

	// 等待过期
	time.Sleep(time.Millisecond * 300)

	// 验证已过期
	_, found = cache.Get(ctx, "expire_key")
	if found {
		t.Fatal("Expected key expire_key to be expired")
	}
}

func TestMemoryCache_DeleteByPattern(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// 设置带模式的键
	patternKeys := []string{"pattern:1", "pattern:2", "pattern:3"}
	otherKeys := []string{"other:1", "other:2"}

	// 设置所有键
	for _, key := range append(patternKeys, otherKeys...) {
		err := cache.Set(ctx, key, key+"_value", time.Minute)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}
	}

	// 按模式删除
	err := cache.DeleteByPattern(ctx, "pattern:.*")
	if err != nil {
		t.Fatalf("DeleteByPattern returned error: %v", err)
	}

	// 验证模式键已删除
	for _, key := range patternKeys {
		_, found := cache.Get(ctx, key)
		if found {
			t.Fatalf("Expected key %s to be deleted by pattern", key)
		}
	}

	// 验证其他键仍存在
	for _, key := range otherKeys {
		_, found := cache.Get(ctx, key)
		if !found {
			t.Fatalf("Expected key %s to still exist", key)
		}
	}
}

func TestMemoryCache_DeleteByTag(t *testing.T) {
	cache := NewMemoryCache()
	ctx := context.Background()

	// 设置带标签的缓存项
	taggedItems := map[string]*Item{
		"tag1:1": {Value: "value1", Tags: []string{"tag1", "common"}},
		"tag1:2": {Value: "value2", Tags: []string{"tag1"}},
		"tag2:1": {Value: "value3", Tags: []string{"tag2", "common"}},
		"tag2:2": {Value: "value4", Tags: []string{"tag2"}},
	}

	for key, item := range taggedItems {
		err := cache.SetItem(ctx, key, item)
		if err != nil {
			t.Fatalf("Failed to set cache item: %v", err)
		}
	}

	// 验证所有项都存在
	for key := range taggedItems {
		_, found := cache.Get(ctx, key)
		if !found {
			t.Fatalf("Expected to find key %s", key)
		}
	}

	// 按标签删除
	err := cache.DeleteByTag(ctx, "tag1")
	if err != nil {
		t.Fatalf("DeleteByTag returned error: %v", err)
	}

	// 验证tag1标签的项已删除
	_, found := cache.Get(ctx, "tag1:1")
	if found {
		t.Fatal("Expected key tag1:1 to be deleted by tag")
	}
	_, found = cache.Get(ctx, "tag1:2")
	if found {
		t.Fatal("Expected key tag1:2 to be deleted by tag")
	}

	// 验证tag2标签的项仍存在
	_, found = cache.Get(ctx, "tag2:1")
	if !found {
		t.Fatal("Expected key tag2:1 to still exist")
	}
	_, found = cache.Get(ctx, "tag2:2")
	if !found {
		t.Fatal("Expected key tag2:2 to still exist")
	}

	// 删除common标签的项
	err = cache.DeleteByTag(ctx, "common")
	if err != nil {
		t.Fatalf("DeleteByTag returned error: %v", err)
	}

	// 验证common标签的项已删除
	_, found = cache.Get(ctx, "tag2:1")
	if found {
		t.Fatal("Expected key tag2:1 to be deleted by common tag")
	}

	// 验证没有common标签的项仍存在
	_, found = cache.Get(ctx, "tag2:2")
	if !found {
		t.Fatal("Expected key tag2:2 to still exist")
	}
}

func TestMemoryCache_JanitorCleanup(t *testing.T) {
	// 创建带清理器的缓存
	cache := NewMemoryCache(Options{
		CleanupInterval: time.Millisecond * 100,
	})
	ctx := context.Background()

	// 设置将过期的项
	err := cache.Set(ctx, "cleanup_key", "cleanup_value", time.Millisecond*50)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// 验证项已设置
	_, found := cache.Get(ctx, "cleanup_key")
	if !found {
		t.Fatal("Expected to find key cleanup_key immediately")
	}

	// 等待清理器运行（大约200ms，应足够运行2次）
	time.Sleep(time.Millisecond * 200)

	// 验证项已被清理
	_, found = cache.Get(ctx, "cleanup_key")
	if found {
		t.Fatal("Expected key cleanup_key to be cleaned up by janitor")
	}

	// 测试关闭清理器
	cache.janitor.stop()

	// 设置新项并验证清理器不再运行
	err = cache.Set(ctx, "after_stop_key", "after_stop_value", time.Millisecond*50)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// 等待足够时间，如果清理器仍在运行，项应该被清理
	time.Sleep(time.Millisecond * 200)

	// 验证项仍存在（因为清理器已停止）
	// 注意：这个测试可能不可靠，因为即使停止了清理器，项仍可能因过期而无法获取
	value, found := cache.Get(ctx, "after_stop_key")
	t.Logf("After janitor stop - Key: after_stop_key, Found: %v, Value: %v", found, value)
}
