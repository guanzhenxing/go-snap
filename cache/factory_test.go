package cache

import (
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	// 测试创建内存缓存
	t.Run("MemoryCache", func(t *testing.T) {
		// 默认选项
		cache, err := NewCache(CacheTypeMemory, nil)
		if err != nil {
			t.Fatalf("NewCache(Memory, nil) returned error: %v", err)
		}
		if cache == nil {
			t.Fatal("NewCache(Memory, nil) returned nil cache")
		}
		_, ok := cache.(*MemoryCache)
		if !ok {
			t.Fatalf("Expected *MemoryCache, got %T", cache)
		}

		// 自定义选项
		cache, err = NewCache(CacheTypeMemory, Options{
			DefaultTTL:      time.Minute,
			CleanupInterval: time.Second * 30,
		})
		if err != nil {
			t.Fatalf("NewCache(Memory, Options) returned error: %v", err)
		}
		if cache == nil {
			t.Fatal("NewCache(Memory, Options) returned nil cache")
		}
		_, ok = cache.(*MemoryCache)
		if !ok {
			t.Fatalf("Expected *MemoryCache, got %T", cache)
		}

		// 无效选项
		_, err = NewCache(CacheTypeMemory, "invalid options")
		if err == nil {
			t.Fatal("NewCache(Memory, invalid) should return error")
		}
	})

	// 其他缓存类型的错误情况测试
	t.Run("UnsupportedType", func(t *testing.T) {
		_, err := NewCache("unsupported", nil)
		if err == nil {
			t.Fatal("NewCache(unsupported) should return error")
		}
	})

	// Redis缓存的测试需要模拟，无法直接测试连接
	t.Run("RedisCache", func(t *testing.T) {
		// 跳过测试，因为它需要真实的Redis连接
		t.Skip("Skipping Redis tests that require actual connection")
	})

	// 多级缓存的错误测试
	t.Run("MultiLevelCache", func(t *testing.T) {
		_, err := NewCache(CacheTypeMultiLevel, "invalid options")
		if err == nil {
			t.Fatal("NewCache(MultiLevel, invalid) should return error")
		}
	})
}

func TestBuilderPattern(t *testing.T) {
	// 测试Builder模式创建内存缓存
	t.Run("MemoryCacheBuilder", func(t *testing.T) {
		builder := NewBuilder().
			WithType(CacheTypeMemory).
			WithOptions(Options{
				DefaultTTL:      time.Minute,
				CleanupInterval: time.Second * 30,
			})

		cache, err := builder.Build()
		if err != nil {
			t.Fatalf("Builder.Build() returned error: %v", err)
		}
		if cache == nil {
			t.Fatal("Builder.Build() returned nil cache")
		}
		_, ok := cache.(*MemoryCache)
		if !ok {
			t.Fatalf("Expected *MemoryCache, got %T", cache)
		}
	})

	// 测试无类型的Builder
	t.Run("NoTypeBuilder", func(t *testing.T) {
		builder := NewBuilder()
		_, err := builder.Build()
		if err == nil {
			t.Fatal("Builder.Build() with no type should return error")
		}
	})

	// 测试无效选项的Builder
	t.Run("InvalidOptionsBuilder", func(t *testing.T) {
		builder := NewBuilder().
			WithType(CacheTypeMemory).
			WithOptions("invalid options")

		_, err := builder.Build()
		if err == nil {
			t.Fatal("Builder.Build() with invalid options should return error")
		}
	})

	// 测试多级缓存的错误情况
	t.Run("MultiLevelCacheErrors", func(t *testing.T) {
		// 没有本地缓存
		builder := NewBuilder().
			WithType(CacheTypeMultiLevel).
			WithOptions(MultiLevelOptions{})

		_, err := builder.Build()
		if err == nil {
			t.Fatal("Builder.Build() for multi-level without local cache should return error")
		}
	})

	// 测试序列化器
	t.Run("WithSerializer", func(t *testing.T) {
		// 使用JSON序列化器
		jsonSerializer := &JSONSerializer{}
		builder := NewBuilder().
			WithType(CacheTypeMemory).
			WithSerializer(jsonSerializer)

		cache, err := builder.Build()
		if err != nil {
			t.Fatalf("Builder.Build() with serializer returned error: %v", err)
		}
		if cache == nil {
			t.Fatal("Builder.Build() with serializer returned nil cache")
		}
	})

	// 测试本地缓存设置
	t.Run("WithLocalCache", func(t *testing.T) {
		// 创建一个内存缓存作为本地缓存
		localCache := NewMemoryCache()

		// 尝试使用Builder创建多级缓存，但缺少远程缓存配置
		builder := NewBuilder().
			WithType(CacheTypeMultiLevel).
			WithLocalCache(localCache).
			WithOptions(MultiLevelOptions{})

		// 这应该失败，因为缺少远程缓存
		_, err := builder.Build()
		if err == nil {
			t.Fatal("Builder.Build() for multi-level without remote cache config should return error")
		}
	})
}
