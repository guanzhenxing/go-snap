package main

import (
	"context"
	"fmt"
	"time"

	"github.com/guanzhenxing/go-snap/cache"
)

func main() {
	// 创建一个基本上下文
	ctx := context.Background()

	fmt.Println("===== 多级缓存示例 =====")
	fmt.Println("创建内存缓存和Redis缓存...")

	// 创建内存缓存
	memoryCache := cache.NewMemoryCache(cache.Options{
		DefaultTTL:      time.Minute * 5, // 本地缓存TTL较短
		CleanupInterval: time.Minute,
	})

	// 创建Redis缓存选项
	redisOpts := cache.RedisOptions{
		Mode:       cache.RedisModeStandalone,
		Addr:       "localhost:6379",
		DB:         0,
		KeyPrefix:  "multilevel",
		DefaultTTL: time.Hour, // 远程缓存TTL较长
	}

	// 创建Redis缓存
	var remoteCacheImpl cache.Cache
	redisCache, err := cache.NewRedisCache(redisOpts, nil)
	if err != nil {
		fmt.Printf("创建Redis缓存失败: %v\n", err)
		fmt.Println("继续使用内存缓存进行演示...")
		// 使用另一个内存缓存作为替代
		remoteCacheImpl = cache.NewMemoryCache(cache.Options{
			DefaultTTL:      time.Hour,
			CleanupInterval: time.Minute,
		})
	} else {
		remoteCacheImpl = redisCache
	}

	// 创建多级缓存
	multiCache, err := cache.NewMultiLevelCache(
		memoryCache,
		remoteCacheImpl,
		cache.MultiLevelOptions{
			WriteMode: cache.WriteModeWriteThrough, // 直写模式
			LocalTTL:  time.Minute * 5,             // 本地缓存5分钟
		},
	)
	if err != nil {
		fmt.Printf("创建多级缓存失败: %v\n", err)
		return
	}

	fmt.Println("多级缓存创建成功")

	// 基本多级缓存操作
	fmt.Println("\n===== 基本多级缓存操作 =====")

	// 设置缓存
	err = multiCache.Set(ctx, "multi-key", "multi-value", time.Minute*30)
	if err != nil {
		fmt.Printf("设置多级缓存失败: %v\n", err)
		return
	}
	fmt.Println("设置多级缓存成功: multi-key = multi-value")

	// 从多级缓存获取（应该从本地缓存获取）
	value, found := multiCache.Get(ctx, "multi-key")
	if found {
		fmt.Printf("从多级缓存获取成功: multi-key = %v (从本地缓存)\n", value)
	} else {
		fmt.Println("从多级缓存获取失败")
	}

	// 演示缓存穿透
	fmt.Println("\n===== 缓存穿透演示 =====")

	// 删除本地缓存中的键
	fmt.Println("删除本地缓存中的键...")
	_ = memoryCache.Delete(ctx, "multi-key")

	// 现在再次尝试获取，应该从Redis获取后填充本地缓存
	value, found = multiCache.Get(ctx, "multi-key")
	if found {
		fmt.Printf("从多级缓存获取成功: multi-key = %v (从远程缓存加载到本地缓存)\n", value)
	} else {
		fmt.Println("从多级缓存获取失败")
	}

	// 再次获取，应该直接从本地缓存获取
	value, found = multiCache.Get(ctx, "multi-key")
	if found {
		fmt.Printf("从多级缓存再次获取: multi-key = %v (应该从本地缓存)\n", value)
	}

	// 测试递增/递减操作
	fmt.Println("\n===== 递增/递减操作 =====")

	multiCache.Set(ctx, "counter", 10, time.Minute)
	fmt.Println("设置计数器: counter = 10")

	// 递增
	newValue, err := multiCache.Increment(ctx, "counter", 5)
	if err != nil {
		fmt.Printf("递增失败: %v\n", err)
	} else {
		fmt.Printf("递增成功: counter = %v\n", newValue)
	}

	// 递减
	newValue, err = multiCache.Decrement(ctx, "counter", 3)
	if err != nil {
		fmt.Printf("递减失败: %v\n", err)
	} else {
		fmt.Printf("递减成功: counter = %v\n", newValue)
	}

	// 使用标签
	fmt.Println("\n===== 使用标签 =====")

	multiCache.SetItem(ctx, "tagged1", &cache.Item{
		Value:      "value1",
		Expiration: time.Minute,
		Tags:       []string{"group1"},
	})

	multiCache.SetItem(ctx, "tagged2", &cache.Item{
		Value:      "value2",
		Expiration: time.Minute,
		Tags:       []string{"group1"},
	})

	fmt.Println("设置了两个带标签'group1'的缓存项")

	// 通过标签删除
	err = multiCache.DeleteByTag(ctx, "group1")
	if err != nil {
		fmt.Printf("通过标签删除失败: %v\n", err)
	} else {
		fmt.Println("通过标签删除成功")

		// 验证删除
		_, found1 := multiCache.Get(ctx, "tagged1")
		_, found2 := multiCache.Get(ctx, "tagged2")
		fmt.Printf("tagged1存在: %v, tagged2存在: %v\n", found1, found2)
	}

	// 清空特定级别的缓存
	fmt.Println("\n===== 清空特定级别的缓存 =====")

	multiCache.Set(ctx, "level-test", "test-value", time.Minute)

	err = multiCache.FlushLevel(ctx, cache.CacheLevelLocal)
	if err != nil {
		fmt.Printf("清空本地缓存失败: %v\n", err)
	} else {
		fmt.Println("清空本地缓存成功")

		// 验证本地缓存被清空但远程缓存仍然存在
		_, localFound := memoryCache.Get(ctx, "level-test")
		_, remoteFound := remoteCacheImpl.Get(ctx, "level-test")
		fmt.Printf("本地缓存存在: %v, 远程缓存存在: %v\n", localFound, remoteFound)
	}

	// 关闭缓存
	multiCache.Close()
	fmt.Println("\n多级缓存已关闭")
}
